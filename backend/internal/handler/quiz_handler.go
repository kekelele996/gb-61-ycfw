package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/constants"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

// parseQuizOptions decodes a quiz question's JSON-encoded option list.
func parseQuizOptions(raw string) []string {
	var options []string
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return []string{}
	}
	return options
}

// QuizHandler exposes quiz and wrong-answer book endpoints.
type QuizHandler struct {
	svc    *service.QuizService
	logger *slog.Logger
}

// NewQuizHandler creates a QuizHandler.
func NewQuizHandler(svc *service.QuizService, logger *slog.Logger) *QuizHandler {
	return &QuizHandler{svc: svc, logger: logger}
}

// ListQuestions handles GET /quiz/questions (public).
func (h *QuizHandler) ListQuestions(c *gin.Context) {
	questions, err := h.svc.ListQuestions()
	if err != nil {
		c.Error(err)
		return
	}
	items := make([]dto.QuizQuestionItem, 0, len(questions))
	for _, q := range questions {
		items = append(items, dto.QuizQuestionItem{
			ID:       q.ID,
			Question: q.Question,
			Options:  parseQuizOptions(q.Options),
		})
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Submit handles POST /quiz/submit (login required).
func (h *QuizHandler) Submit(c *gin.Context) {
	var req dto.QuizSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	result, err := h.svc.Submit(middleware.GetUserID(c), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

// ListWrongQuestions handles GET /quiz/wrong-questions (login required).
func (h *QuizHandler) ListWrongQuestions(c *gin.Context) {
	var mastered *bool
	switch c.Query("mastered") {
	case "true":
		v := true
		mastered = &v
	case "false":
		v := false
		mastered = &v
	}
	items, err := h.svc.ListWrongQuestions(middleware.GetUserID(c), mastered)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// WrongQuestionStats handles GET /quiz/wrong-questions/stats (login required).
func (h *QuizHandler) WrongQuestionStats(c *gin.Context) {
	stats, err := h.svc.WrongQuestionStats(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(stats))
}

// RetryWrongQuestion handles PUT /quiz/wrong-questions/:questionId/retry (login required).
func (h *QuizHandler) RetryWrongQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid question id"))
		return
	}
	var req dto.WrongQuestionRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	result, err := h.svc.RetryWrongQuestion(middleware.GetUserID(c), uint(questionID), *req.SelectedIndex)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}
