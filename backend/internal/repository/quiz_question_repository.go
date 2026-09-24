package repository

import (
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

// List returns all quiz questions ordered by id.
func (r *QuizQuestionRepository) List() ([]model.QuizQuestion, error) {
	var items []model.QuizQuestion
	if err := r.db.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID returns one quiz question.
func (r *QuizQuestionRepository) FindByID(id uint) (*model.QuizQuestion, error) {
	var q model.QuizQuestion
	if err := r.db.First(&q, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &q, nil
}

// FindByIDs returns quiz questions matching the given ids, ordered by id.
func (r *QuizQuestionRepository) FindByIDs(ids []uint) ([]model.QuizQuestion, error) {
	var items []model.QuizQuestion
	if err := r.db.Where("id IN ?", ids).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Count returns the number of quiz questions.
func (r *QuizQuestionRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.QuizQuestion{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
