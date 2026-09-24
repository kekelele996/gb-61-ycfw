package model

import "time"

// QuizQuestion is a single multiple-choice question in the care quiz bank.
type QuizQuestion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Question    string    `gorm:"size:255;not null" json:"question"`
	Options     string    `gorm:"type:json;not null" json:"options"`
	AnswerIndex int       `gorm:"not null" json:"answer_index"`
	Explanation string    `gorm:"size:512;not null;default:''" json:"explanation"`
	CreatedAt   time.Time `json:"created_at"`
}

// WrongQuestion records a user's wrong answer for a quiz question. The unique
// index on (user_id, question_id) guarantees one entry per question per user.
type WrongQuestion struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index:idx_wrong_user_question,unique;not null" json:"user_id"`
	QuestionID    uint      `gorm:"index:idx_wrong_user_question,unique;not null" json:"question_id"`
	SelectedIndex int       `gorm:"not null" json:"selected_index"`
	Mastered      bool      `gorm:"not null;default:false" json:"mastered"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
