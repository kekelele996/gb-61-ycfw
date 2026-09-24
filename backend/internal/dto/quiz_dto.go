package dto

// QuizOptionItem is one selectable option of a quiz question.
type QuizOptionItem struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// QuizQuestionItem is a quiz question exposed to the frontend.
type QuizQuestionItem struct {
	ID          uint             `json:"id"`
	Question    string           `json:"question"`
	Options     []QuizOptionItem `json:"options"`
	Answer      int              `json:"answer"`
	Explanation string           `json:"explanation"`
}

// QuizSubmitItem is one answered question in a quiz submission.
type QuizSubmitItem struct {
	QuestionID    *uint `json:"question_id" binding:"required"`
	SelectedIndex *int  `json:"selected_index" binding:"required"`
}

// QuizSubmitRequest is the payload for submitting a full quiz.
type QuizSubmitRequest struct {
	Answers []QuizSubmitItem `json:"answers" binding:"required,min=1,dive"`
}

// QuizSubmitResult is the graded quiz result.
type QuizSubmitResult struct {
	Total      int    `json:"total"`
	Correct    int    `json:"correct"`
	Score      int    `json:"score"`
	WrongCount int    `json:"wrong_count"`
	WrongIDs   []uint `json:"wrong_ids"`
	Saved      bool   `json:"saved"`
}

// QuizRetryRequest is the payload for re-answering one wrong question.
type QuizRetryRequest struct {
	SelectedIndex *int `json:"selected_index" binding:"required"`
}

// QuizRetryResult reports whether the re-answered question is now mastered.
type QuizRetryResult struct {
	QuestionID uint `json:"question_id"`
	Correct    bool `json:"correct"`
	Mastered   bool `json:"mastered"`
}

// WrongQuestionItem is a wrong-question book entry with question detail joined.
type WrongQuestionItem struct {
	ID            uint             `json:"id"`
	QuestionID    uint             `json:"question_id"`
	SelectedIndex int              `json:"selected_index"`
	Mastered      bool             `json:"mastered"`
	Question      string           `json:"question"`
	Options       []QuizOptionItem `json:"options"`
	Answer        int              `json:"answer"`
	Explanation   string           `json:"explanation"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
}

// WrongBookStats summarizes a user's wrong-question book.
type WrongBookStats struct {
	Total      int     `json:"total"`
	Unmastered int     `json:"unmastered"`
	Mastered   int     `json:"mastered"`
	Progress   float64 `json:"progress"`
}
