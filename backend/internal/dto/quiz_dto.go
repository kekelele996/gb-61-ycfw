package dto

import "time"

// QuizQuestionItem is a quiz question without the correct answer for taking
// the quiz. The answer and explanation are returned after submission.
type QuizQuestionItem struct {
	ID          uint     `json:"id"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
}

// QuizSubmitItem is one answer inside a quiz submission.
type QuizSubmitItem struct {
	QuestionID    uint `json:"question_id" binding:"required"`
	SelectedIndex *int `json:"selected_index" binding:"required"`
}

// QuizSubmitRequest is the payload for submitting a full quiz.
type QuizSubmitRequest struct {
	Answers []QuizSubmitItem `json:"answers" binding:"required,min=1,dive"`
}

// QuizResultItem is the grading result for one answered question.
type QuizResultItem struct {
	QuestionID    uint   `json:"question_id"`
	Question      string `json:"question"`
	SelectedIndex int    `json:"selected_index"`
	AnswerIndex   int    `json:"answer_index"`
	Correct       bool   `json:"correct"`
	Explanation   string `json:"explanation"`
}

// QuizSubmitResult is the overall grading result.
type QuizSubmitResult struct {
	Total   int              `json:"total"`
	Correct int              `json:"correct"`
	Score   int              `json:"score"`
	Results []QuizResultItem `json:"results"`
}

// WrongQuestionRetryRequest is the payload for redoing one wrong question.
type WrongQuestionRetryRequest struct {
	SelectedIndex *int `json:"selected_index" binding:"required"`
}

// WrongQuestionItem is one entry in the wrong-question book.
type WrongQuestionItem struct {
	ID            uint      `json:"id"`
	QuestionID    uint      `json:"question_id"`
	Question      string    `json:"question"`
	Options       []string  `json:"options"`
	AnswerIndex   int       `json:"answer_index"`
	SelectedIndex int       `json:"selected_index"`
	Mastered      bool      `json:"mastered"`
	Explanation   string    `json:"explanation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WrongQuestionStats summarizes the user's mastery progress.
type WrongQuestionStats struct {
	Total      int64 `json:"total"`
	Unmastered int64 `json:"unmastered"`
	Mastered   int64 `json:"mastered"`
}
