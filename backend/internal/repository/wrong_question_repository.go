package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// WrongQuestionRepository handles persistence of users' wrong-answer book.
type WrongQuestionRepository struct {
	db *gorm.DB
}

// NewWrongQuestionRepository creates a WrongQuestionRepository.
func NewWrongQuestionRepository(db *gorm.DB) *WrongQuestionRepository {
	return &WrongQuestionRepository{db: db}
}

// Upsert inserts a wrong-question entry or updates the selected index and
// resets mastery when the (user_id, question_id) pair already exists.
func (r *WrongQuestionRepository) Upsert(userID, questionID uint, selectedIndex int) error {
	w := model.WrongQuestion{
		UserID:        userID,
		QuestionID:    questionID,
		SelectedIndex: selectedIndex,
		Mastered:      false,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"selected_index": selectedIndex,
			"mastered":       false,
			"updated_at":     time.Now(),
		}),
	}).Create(&w).Error
}

// Find locates a user's wrong-question entry for a quiz question.
func (r *WrongQuestionRepository) Find(userID, questionID uint) (*model.WrongQuestion, error) {
	var w model.WrongQuestion
	if err := r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &w, nil
}

// Update persists a wrong-question entry.
func (r *WrongQuestionRepository) Update(w *model.WrongQuestion) error {
	return r.db.Save(w).Error
}

// ListByUser returns a user's wrong-question entries joined with quiz questions.
// A nil mastered pointer means no mastery filter.
func (r *WrongQuestionRepository) ListByUser(userID uint, mastered *bool) ([]WrongQuestionView, error) {
	var items []WrongQuestionView
	q := r.db.Table("wrong_questions AS w").
		Select("w.id AS id, w.user_id AS user_id, w.question_id AS question_id, "+
			"w.selected_index AS selected_index, w.mastered AS mastered, "+
			"w.created_at AS created_at, w.updated_at AS updated_at, "+
			"q.question AS question, q.options AS options, q.answer_index AS answer_index, "+
			"q.explanation AS explanation").
		Joins("LEFT JOIN quiz_questions AS q ON q.id = w.question_id").
		Where("w.user_id = ?", userID)
	if mastered != nil {
		q = q.Where("w.mastered = ?", *mastered)
	}
	if err := q.Order("w.mastered ASC, w.updated_at DESC").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountByUser returns the total and mastered counts of a user's wrong questions.
func (r *WrongQuestionRepository) CountByUser(userID uint) (total int64, mastered int64, err error) {
	if err = r.db.Model(&model.WrongQuestion{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ? AND mastered = ?", userID, true).Count(&mastered).Error; err != nil {
		return 0, 0, err
	}
	return total, mastered, nil
}

// WrongQuestionView joins a wrong-question entry with its quiz question content.
type WrongQuestionView struct {
	ID            uint      `gorm:"column:id"`
	UserID        uint      `gorm:"column:user_id"`
	QuestionID    uint      `gorm:"column:question_id"`
	SelectedIndex int       `gorm:"column:selected_index"`
	Mastered      bool      `gorm:"column:mastered"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	Question      string    `gorm:"column:question"`
	Options       string    `gorm:"column:options"`
	AnswerIndex   int       `gorm:"column:answer_index"`
	Explanation   string    `gorm:"column:explanation"`
}
