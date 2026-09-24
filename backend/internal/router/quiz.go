package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerQuizRoutes(v1 *gin.RouterGroup, cfg *config.Config, h *handler.QuizHandler, limiter *middleware.RateLimiter) {
	quiz := v1.Group("/quiz")
	quiz.GET("/questions", h.ListQuestions)
	quiz.POST("/submit", middleware.AuthOptional(cfg), limiter.Limit(), h.Submit)

	wrongBook := quiz.Group("/wrong-book", middleware.AuthRequired(cfg))
	wrongBook.GET("", h.ListWrongBook)
	wrongBook.GET("/stats", h.WrongBookStats)
	wrongBook.POST("/:questionId/retry", limiter.Limit(), h.RetryWrong)
}
