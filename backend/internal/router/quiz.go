package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerQuizRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.QuizHandler, limiter *middleware.RateLimiter) {
	quiz := v1.Group("/quiz")
	// The question bank is public and submission works anonymously too:
	// anonymous users get their score but no wrong-answer record is kept.
	quiz.GET("/questions", h.ListQuestions)
	quiz.POST("/submit", middleware.AuthOptional(cfg), limiter.Limit(), h.Submit)

	auth := quiz.Group("", middleware.AuthRequired(cfg))
	auth.GET("/wrong-questions", h.ListWrongQuestions)
	auth.GET("/wrong-questions/stats", h.WrongQuestionStats)
	auth.PUT("/wrong-questions/:questionId/retry", h.RetryWrongQuestion)
}
