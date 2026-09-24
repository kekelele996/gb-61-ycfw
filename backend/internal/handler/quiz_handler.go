package handler

import (
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

// QuizHandler exposes quiz and wrong-question book endpoints.
type QuizHandler struct {
	svc    *service.QuizService
	logger *slog.Logger
}

// NewQuizHandler creates a QuizHandler.
func NewQuizHandler(svc *service.QuizService, logger *slog.Logger) *QuizHandler {
	return &QuizHandler{svc: svc, logger: logger}
}

// ListQuestions handles GET /quiz/questions.
func (h *QuizHandler) ListQuestions(c *gin.Context) {
	items, err := h.svc.ListQuestions()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// Submit handles POST /quiz/submit (anonymous allowed; wrong answers are only
// saved when the request carries a valid JWT).
func (h *QuizHandler) Submit(c *gin.Context) {
	var req dto.QuizSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	result, err := h.svc.Submit(middleware.GetUserID(c), req.Answers)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

// ListWrongBook handles GET /quiz/wrong-book.
func (h *QuizHandler) ListWrongBook(c *gin.Context) {
	items, err := h.svc.ListWrongBook(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// WrongBookStats handles GET /quiz/wrong-book/stats.
func (h *QuizHandler) WrongBookStats(c *gin.Context) {
	stats, err := h.svc.WrongBookStats(middleware.GetUserID(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(stats))
}

// RetryWrong handles POST /quiz/wrong-book/:questionId/retry.
func (h *QuizHandler) RetryWrong(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 64)
	if err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, "invalid question id"))
		return
	}
	var req dto.QuizRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, constants.CodeBadRequest, constants.MsgInvalidParam))
		return
	}
	result, err := h.svc.RetryWrong(middleware.GetUserID(c), uint(questionID), *req.SelectedIndex)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}
