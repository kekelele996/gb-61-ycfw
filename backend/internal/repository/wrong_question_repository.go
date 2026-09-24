package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// WrongQuestionRepository handles persistence of the wrong-question book.
type WrongQuestionRepository struct {
	db *gorm.DB
}

// NewWrongQuestionRepository creates a WrongQuestionRepository.
func NewWrongQuestionRepository(db *gorm.DB) *WrongQuestionRepository {
	return &WrongQuestionRepository{db: db}
}

// UpsertWrong inserts a wrong entry or, for an existing (user, question)
// row, refreshes the selected option and resets it to unmastered.
func (r *WrongQuestionRepository) UpsertWrong(userID, questionID uint, selectedIndex int) error {
	w := model.WrongQuestion{
		UserID:        userID,
		QuestionID:    questionID,
		SelectedIndex: selectedIndex,
		Mastered:      false,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"selected_index", "mastered", "updated_at"}),
	}).Create(&w).Error
}

// MarkMastered marks an existing wrong entry as mastered.
func (r *WrongQuestionRepository) MarkMastered(userID, questionID uint) error {
	res := r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Updates(map[string]interface{}{"mastered": true})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByUser returns a user's wrong-question entries ordered newest first.
func (r *WrongQuestionRepository) ListByUser(userID uint) ([]model.WrongQuestion, error) {
	var items []model.WrongQuestion
	if err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// CountByUser returns the total and unmastered number of a user's wrong entries.
func (r *WrongQuestionRepository) CountByUser(userID uint) (total int64, unmastered int64, err error) {
	if err = r.db.Model(&model.WrongQuestion{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	if err = r.db.Model(&model.WrongQuestion{}).
		Where("user_id = ? AND mastered = ?", userID, false).Count(&unmastered).Error; err != nil {
		return 0, 0, err
	}
	return total, unmastered, nil
}
