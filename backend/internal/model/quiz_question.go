package model

import "time"

// QuizQuestion is a fixed care-knowledge quiz question.
type QuizQuestion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Question    string    `gorm:"size:255;not null" json:"question"`
	Options     string    `gorm:"type:json;not null" json:"-"`
	Answer      int       `gorm:"not null" json:"answer"`
	Explanation string    `gorm:"size:512;not null" json:"explanation"`
	CreatedAt   time.Time `json:"created_at"`
}
