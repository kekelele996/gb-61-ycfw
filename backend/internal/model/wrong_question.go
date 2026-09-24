package model

import "time"

// WrongQuestion is one entry of a user's quiz wrong-question book.
// There is at most one row per (user, quiz_question): re-answering
// wrongly updates selected_index and keeps the question unmastered.
type WrongQuestion struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index:idx_wrong_user_question,unique;not null" json:"user_id"`
	QuestionID    uint      `gorm:"index:idx_wrong_user_question,unique;not null" json:"question_id"`
	SelectedIndex int       `gorm:"not null" json:"selected_index"`
	Mastered      bool      `gorm:"not null;default:false" json:"mastered"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
