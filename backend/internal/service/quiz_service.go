package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// QuizService implements quiz listing/grading and wrong-question book logic.
type QuizService struct {
	questionRepo *repository.QuizQuestionRepository
	wrongRepo    *repository.WrongQuestionRepository
	logger       *slog.Logger
}

// NewQuizService creates a QuizService.
func NewQuizService(questionRepo *repository.QuizQuestionRepository, wrongRepo *repository.WrongQuestionRepository, logger *slog.Logger) *QuizService {
	return &QuizService{questionRepo: questionRepo, wrongRepo: wrongRepo, logger: logger}
}

// ListQuestions returns the quiz bank for rendering the quiz page.
func (s *QuizService) ListQuestions() ([]dto.QuizQuestionItem, error) {
	items, err := s.questionRepo.List()
	if err != nil {
		return nil, fmt.Errorf("quiz question list: %w", err)
	}
	out := make([]dto.QuizQuestionItem, 0, len(items))
	for i := range items {
		item, err := toQuestionItem(&items[i])
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

// Submit grades a full quiz. When userID is non-zero the wrong answers are
// upserted into the user's wrong-question book; anonymous submissions are
// graded without leaving any learning record.
func (s *QuizService) Submit(userID uint, answers []dto.QuizSubmitItem) (*dto.QuizSubmitResult, error) {
	ids := make([]uint, 0, len(answers))
	selected := make(map[uint]int, len(answers))
	for _, a := range answers {
		ids = append(ids, *a.QuestionID)
		selected[*a.QuestionID] = *a.SelectedIndex
	}
	questions, err := s.questionRepo.FindByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("quiz submit question find: %w", err)
	}
	if len(questions) != len(ids) {
		return nil, util.NewAppError(422, constants.CodeValidationError, "quiz submit failed: unknown question id")
	}

	result := &dto.QuizSubmitResult{
		Total:    len(questions),
		WrongIDs: []uint{},
	}
	wrong := make([]struct {
		id       uint
		selected int
	}, 0)
	for _, q := range questions {
		options, err := decodeOptions(q.Options)
		if err != nil {
			return nil, err
		}
		sel := selected[q.ID]
		if sel < 0 || sel >= len(options) {
			return nil, util.NewAppError(422, constants.CodeValidationError,
				fmt.Sprintf("quiz submit failed: invalid option for question %d", q.ID))
		}
		if sel == q.Answer {
			result.Correct++
		} else {
			result.WrongCount++
			result.WrongIDs = append(result.WrongIDs, q.ID)
			wrong = append(wrong, struct {
				id       uint
				selected int
			}{q.ID, sel})
		}
	}
	result.Score = result.Correct

	if userID != 0 && len(wrong) > 0 {
		for _, w := range wrong {
			if err := s.wrongRepo.UpsertWrong(userID, w.id, w.selected); err != nil {
				s.logger.Error("wrong question upsert failed", "error", err, "user_id", userID, "question_id", w.id)
				return nil, fmt.Errorf("quiz submit wrong upsert: %w", err)
			}
		}
		result.Saved = true
		s.logger.Info("quiz submitted, wrong questions saved", "user_id", userID,
			"total", result.Total, "correct", result.Correct, "wrong", result.WrongCount)
	}
	return result, nil
}

// ListWrongBook returns a user's wrong-question entries with question detail.
func (s *QuizService) ListWrongBook(userID uint) ([]dto.WrongQuestionItem, error) {
	entries, err := s.wrongRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("wrong book list: %w", err)
	}
	if len(entries) == 0 {
		return []dto.WrongQuestionItem{}, nil
	}
	ids := make([]uint, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.QuestionID)
	}
	questions, err := s.questionRepo.FindByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("wrong book question find: %w", err)
	}
	questionMap := make(map[uint]*model.QuizQuestion, len(questions))
	for i := range questions {
		questionMap[questions[i].ID] = &questions[i]
	}
	out := make([]dto.WrongQuestionItem, 0, len(entries))
	for _, e := range entries {
		q, ok := questionMap[e.QuestionID]
		if !ok {
			continue
		}
		item, err := toQuestionItem(q)
		if err != nil {
			return nil, err
		}
		out = append(out, dto.WrongQuestionItem{
			ID:            e.ID,
			QuestionID:    e.QuestionID,
			SelectedIndex: e.SelectedIndex,
			Mastered:      e.Mastered,
			Question:      item.Question,
			Options:       item.Options,
			Answer:        item.Answer,
			Explanation:   item.Explanation,
			CreatedAt:     util.FormatDateTime(e.CreatedAt),
			UpdatedAt:     util.FormatDateTime(e.UpdatedAt),
		})
	}
	return out, nil
}

// WrongBookStats returns the totals and mastery progress of a user's book.
func (s *QuizService) WrongBookStats(userID uint) (*dto.WrongBookStats, error) {
	total, unmastered, err := s.wrongRepo.CountByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("wrong book stats: %w", err)
	}
	mastered := total - unmastered
	stats := &dto.WrongBookStats{Total: int(total), Unmastered: int(unmastered), Mastered: int(mastered)}
	if total > 0 {
		stats.Progress = math.Round(float64(mastered)/float64(total)*1000) / 10
	}
	return stats, nil
}

// RetryWrong grades a re-answer of one wrong question. A correct answer marks
// the entry as mastered; a wrong answer refreshes the selected option and
// keeps it unmastered.
func (s *QuizService) RetryWrong(userID, questionID uint, selectedIndex int) (*dto.QuizRetryResult, error) {
	q, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(404, constants.CodeNotFound, fmt.Sprintf("QuizQuestion[id=%d] not found", questionID))
		}
		return nil, fmt.Errorf("quiz retry question find: %w", err)
	}
	options, err := decodeOptions(q.Options)
	if err != nil {
		return nil, err
	}
	if selectedIndex < 0 || selectedIndex >= len(options) {
		return nil, util.NewAppError(422, constants.CodeValidationError,
			fmt.Sprintf("quiz retry failed: invalid option for question %d", questionID))
	}

	correct := selectedIndex == q.Answer
	if correct {
		if err := s.wrongRepo.MarkMastered(userID, questionID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, util.NewAppError(404, constants.CodeNotFound,
					fmt.Sprintf("WrongQuestion[user_id=%d question_id=%d] not found", userID, questionID))
			}
			return nil, fmt.Errorf("quiz retry mark mastered: %w", err)
		}
		s.logger.Info("wrong question mastered", "user_id", userID, "question_id", questionID)
	} else {
		if err := s.wrongRepo.UpsertWrong(userID, questionID, selectedIndex); err != nil {
			return nil, fmt.Errorf("quiz retry upsert: %w", err)
		}
		s.logger.Info("wrong question retried, still wrong", "user_id", userID, "question_id", questionID)
	}
	return &dto.QuizRetryResult{QuestionID: questionID, Correct: correct, Mastered: correct}, nil
}

func toQuestionItem(q *model.QuizQuestion) (dto.QuizQuestionItem, error) {
	texts, err := decodeOptions(q.Options)
	if err != nil {
		return dto.QuizQuestionItem{}, err
	}
	options := make([]dto.QuizOptionItem, 0, len(texts))
	for i, t := range texts {
		options = append(options, dto.QuizOptionItem{Index: i, Text: t})
	}
	return dto.QuizQuestionItem{
		ID:          q.ID,
		Question:    q.Question,
		Options:     options,
		Answer:      q.Answer,
		Explanation: q.Explanation,
	}, nil
}

func decodeOptions(raw string) ([]string, error) {
	var options []string
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil, fmt.Errorf("quiz options decode: %w", err)
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("quiz options decode: empty options")
	}
	return options, nil
}
