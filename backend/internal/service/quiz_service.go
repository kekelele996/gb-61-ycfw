package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// QuizService implements quiz taking, wrong-answer book and retry logic.
type QuizService struct {
	questionRepo *repository.QuizQuestionRepository
	wrongRepo    *repository.WrongQuestionRepository
	logger       *slog.Logger
}

// NewQuizService creates a QuizService.
func NewQuizService(questionRepo *repository.QuizQuestionRepository, wrongRepo *repository.WrongQuestionRepository, logger *slog.Logger) *QuizService {
	return &QuizService{questionRepo: questionRepo, wrongRepo: wrongRepo, logger: logger}
}

// parseOptions decodes the JSON-encoded option list of a quiz question.
func parseOptions(raw string) []string {
	var options []string
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return []string{}
	}
	return options
}

// ListQuestions returns the public quiz bank ordered by id.
func (s *QuizService) ListQuestions() ([]model.QuizQuestion, error) {
	items, err := s.questionRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("quiz question list: %w", err)
	}
	return items, nil
}

// gradeItem holds a graded answer together with its question.
type gradeItem struct {
	question      model.QuizQuestion
	selectedIndex int
	correct       bool
}

// grade validates the submitted answers against the question bank and records
// each wrong answer into the user's wrong-answer book (one entry per question).
func (s *QuizService) grade(userID uint, answers []dto.QuizSubmitItem) ([]gradeItem, error) {
	idSet := make(map[uint]bool, len(answers))
	for _, a := range answers {
		if idSet[a.QuestionID] {
			return nil, util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("Quiz submit failed: question_id=%d answered more than once", a.QuestionID))
		}
		idSet[a.QuestionID] = true
	}
	ids := make([]uint, 0, len(answers))
	for _, a := range answers {
		ids = append(ids, a.QuestionID)
	}
	questions, err := s.questionRepo.FindByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("quiz submit question find: %w", err)
	}
	if len(questions) != len(ids) {
		return nil, util.NewAppError(422, constants.CodeValidationError, "Quiz submit failed: unknown question id")
	}
	questionMap := make(map[uint]model.QuizQuestion, len(questions))
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	graded := make([]gradeItem, 0, len(answers))
	for _, a := range answers {
		q := questionMap[a.QuestionID]
		selected := *a.SelectedIndex
		if selected < 0 || selected >= len(parseOptions(q.Options)) {
			return nil, util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("Quiz submit failed: invalid selected_index for question_id=%d", q.ID))
		}
		graded = append(graded, gradeItem{question: q, selectedIndex: selected, correct: selected == q.AnswerIndex})
	}

	// Logged-in users keep one wrong-answer entry per question; anonymous
	// submissions are graded only and leave no learning record.
	if userID != 0 {
		for _, g := range graded {
			if g.correct {
				continue
			}
			if err := s.wrongRepo.Upsert(userID, g.question.ID, g.selectedIndex); err != nil {
				return nil, fmt.Errorf("quiz submit wrong question upsert: %w", err)
			}
		}
	}
	s.logger.Info("quiz submitted", "user_id", userID, "total", len(graded))
	return graded, nil
}

// Submit grades a full quiz and persists wrong answers for a logged-in user.
func (s *QuizService) Submit(userID uint, req dto.QuizSubmitRequest) (*dto.QuizSubmitResult, error) {
	graded, err := s.grade(userID, req.Answers)
	if err != nil {
		return nil, err
	}
	result := &dto.QuizSubmitResult{Total: len(graded), Results: make([]dto.QuizResultItem, 0, len(graded))}
	for _, g := range graded {
		if g.correct {
			result.Correct++
		}
		result.Results = append(result.Results, dto.QuizResultItem{
			QuestionID:    g.question.ID,
			Question:      g.question.Question,
			SelectedIndex: g.selectedIndex,
			AnswerIndex:   g.question.AnswerIndex,
			Correct:       g.correct,
			Explanation:   g.question.Explanation,
		})
	}
	result.Score = result.Correct
	return result, nil
}

// ListWrongQuestions returns the user's wrong-question book optionally filtered by mastery.
func (s *QuizService) ListWrongQuestions(userID uint, mastered *bool) ([]dto.WrongQuestionItem, error) {
	views, err := s.wrongRepo.ListByUser(userID, mastered)
	if err != nil {
		return nil, fmt.Errorf("wrong question list: %w", err)
	}
	items := make([]dto.WrongQuestionItem, 0, len(views))
	for _, v := range views {
		items = append(items, dto.WrongQuestionItem{
			ID:            v.ID,
			QuestionID:    v.QuestionID,
			Question:      v.Question,
			Options:       parseOptions(v.Options),
			AnswerIndex:   v.AnswerIndex,
			SelectedIndex: v.SelectedIndex,
			Mastered:      v.Mastered,
			Explanation:   v.Explanation,
			CreatedAt:     v.CreatedAt,
			UpdatedAt:     v.UpdatedAt,
		})
	}
	return items, nil
}

// WrongQuestionStats returns unmastered count and mastery progress.
func (s *QuizService) WrongQuestionStats(userID uint) (*dto.WrongQuestionStats, error) {
	total, mastered, err := s.wrongRepo.CountByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("wrong question stats: %w", err)
	}
	return &dto.WrongQuestionStats{Total: total, Unmastered: total - mastered, Mastered: mastered}, nil
}

// RetryWrongQuestion grades one retry from the wrong-answer book. A correct
// answer marks the question mastered; a wrong one updates the selected index
// and keeps it unmastered.
func (s *QuizService) RetryWrongQuestion(userID, questionID uint, selectedIndex int) (*dto.QuizResultItem, error) {
	w, err := s.wrongRepo.Find(userID, questionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound,
				fmt.Sprintf("WrongQuestion[user_id=%d question_id=%d] not found", userID, questionID))
		}
		return nil, fmt.Errorf("wrong question retry find: %w", err)
	}
	q, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("wrong question retry question find: %w", err)
	}
	if selectedIndex < 0 || selectedIndex >= len(parseOptions(q.Options)) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("WrongQuestion retry failed: invalid selected_index for question_id=%d", questionID))
	}

	correct := selectedIndex == q.AnswerIndex
	if correct {
		w.Mastered = true
	} else {
		w.Mastered = false
		w.SelectedIndex = selectedIndex
	}
	if err := s.wrongRepo.Update(w); err != nil {
		return nil, fmt.Errorf("wrong question retry update: %w", err)
	}
	s.logger.Info("wrong question retried", "user_id", userID, "question_id", questionID, "correct", correct, "mastered", w.Mastered)
	return &dto.QuizResultItem{
		QuestionID:    q.ID,
		Question:      q.Question,
		SelectedIndex: selectedIndex,
		AnswerIndex:   q.AnswerIndex,
		Correct:       correct,
		Explanation:   q.Explanation,
	}, nil
}
