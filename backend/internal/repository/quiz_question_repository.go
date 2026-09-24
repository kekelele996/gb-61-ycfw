package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// QuizQuestionRepository handles persistence of quiz questions.
type QuizQuestionRepository struct {
	db *gorm.DB
}

// NewQuizQuestionRepository creates a QuizQuestionRepository.
func NewQuizQuestionRepository(db *gorm.DB) *QuizQuestionRepository {
	return &QuizQuestionRepository{db: db}
}

// ListAll returns all quiz questions ordered by id.
func (r *QuizQuestionRepository) ListAll() ([]model.QuizQuestion, error) {
	var items []model.QuizQuestion
	if err := r.db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID locates a quiz question by id.
func (r *QuizQuestionRepository) FindByID(id uint) (*model.QuizQuestion, error) {
	var q model.QuizQuestion
	if err := r.db.First(&q, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &q, nil
}

// FindByIDs locates quiz questions by their ids.
func (r *QuizQuestionRepository) FindByIDs(ids []uint) ([]model.QuizQuestion, error) {
	var items []model.QuizQuestion
	if len(ids) == 0 {
		return items, nil
	}
	if err := r.db.Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
