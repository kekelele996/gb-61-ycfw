package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func setupQuizRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.QuizQuestion{}, &model.WrongQuestion{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	opts, _ := json.Marshal([]string{"甲", "乙", "丙", "丁"})
	if err := db.Create(&[]model.QuizQuestion{
		{Question: "q1", Options: string(opts), Answer: 1, Explanation: "e1"},
		{Question: "q2", Options: string(opts), Answer: 2, Explanation: "e2"},
	}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewQuizService(
		repository.NewQuizQuestionRepository(db),
		repository.NewWrongQuestionRepository(db),
		logger,
	)
	h := NewQuizHandler(svc, logger)

	cfg := &config.Config{JWTSecret: "test-secret"}
	limiter := middleware.NewRateLimiter(1000, time.Minute)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler(logger))
	v1 := r.Group("/api/v1")
	v1.GET("/quiz/questions", h.ListQuestions)
	v1.POST("/quiz/submit", middleware.AuthOptional(cfg), limiter.Limit(), h.Submit)
	wrongBook := v1.Group("/quiz/wrong-book", middleware.AuthRequired(cfg))
	wrongBook.GET("", h.ListWrongBook)
	wrongBook.GET("/stats", h.WrongBookStats)
	wrongBook.POST("/:questionId/retry", limiter.Limit(), h.RetryWrong)
	return r, cfg.JWTSecret
}

func doJSON(t *testing.T, r *gin.Engine, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var parsed map[string]any
	if w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), &parsed)
	}
	return w.Code, parsed
}

func TestQuizQuestionList(t *testing.T) {
	r, _ := setupQuizRouter(t)
	status, body := doJSON(t, r, http.MethodGet, "/api/v1/quiz/questions", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%v", status, body)
	}
	data := body["data"].([]any)
	if len(data) != 2 {
		t.Fatalf("want 2 questions, got %d", len(data))
	}
}

func TestQuizSubmitAnonymousThenLoggedInAndRetry(t *testing.T) {
	r, secret := setupQuizRouter(t)

	// Anonymous submit: graded, no saving.
	status, body := doJSON(t, r, http.MethodPost, "/api/v1/quiz/submit", "", map[string]any{
		"answers": []map[string]any{
			{"question_id": 1, "selected_index": 0},
			{"question_id": 2, "selected_index": 2},
		},
	})
	if status != http.StatusOK {
		t.Fatalf("anonymous submit status=%d body=%v", status, body)
	}
	result := body["data"].(map[string]any)
	if result["score"].(float64) != 1 || result["saved"].(bool) {
		t.Fatalf("anonymous result unexpected: %v", result)
	}

	// Wrong-book endpoints require login.
	status, _ = doJSON(t, r, http.MethodGet, "/api/v1/quiz/wrong-book", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("want 401 without token, got %d", status)
	}

	token, err := util.GenerateToken(42, "gardener", "user", secret, time.Hour)
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	// Logged-in submit with one wrong answer.
	status, body = doJSON(t, r, http.MethodPost, "/api/v1/quiz/submit", token, map[string]any{
		"answers": []map[string]any{
			{"question_id": 1, "selected_index": 0},
			{"question_id": 2, "selected_index": 2},
		},
	})
	if status != http.StatusOK {
		t.Fatalf("logged-in submit status=%d body=%v", status, body)
	}
	result = body["data"].(map[string]any)
	if !result["saved"].(bool) || result["wrong_count"].(float64) != 1 {
		t.Fatalf("logged-in result unexpected: %v", result)
	}

	// Stats show one unmastered question.
	status, body = doJSON(t, r, http.MethodGet, "/api/v1/quiz/wrong-book/stats", token, nil)
	if status != http.StatusOK {
		t.Fatalf("stats status=%d", status)
	}
	stats := body["data"].(map[string]any)
	if stats["unmastered"].(float64) != 1 || stats["progress"].(float64) != 0 {
		t.Fatalf("stats unexpected: %v", stats)
	}

	// Wrong book carries question detail and the chosen option.
	status, body = doJSON(t, r, http.MethodGet, "/api/v1/quiz/wrong-book", token, nil)
	if status != http.StatusOK {
		t.Fatalf("book status=%d", status)
	}
	book := body["data"].([]any)
	if len(book) != 1 {
		t.Fatalf("want 1 entry, got %d", len(book))
	}
	entry := book[0].(map[string]any)
	if entry["selected_index"].(float64) != 0 || entry["mastered"].(bool) {
		t.Fatalf("entry unexpected: %v", entry)
	}

	// Retry correctly -> mastered.
	status, body = doJSON(t, r, http.MethodPost, "/api/v1/quiz/wrong-book/1/retry", token, map[string]any{
		"selected_index": 1,
	})
	if status != http.StatusOK || !body["data"].(map[string]any)["mastered"].(bool) {
		t.Fatalf("retry correct status=%d body=%v", status, body)
	}

	// Retry wrongly again -> stays unmastered.
	status, body = doJSON(t, r, http.MethodPost, "/api/v1/quiz/wrong-book/1/retry", token, map[string]any{
		"selected_index": 3,
	})
	if status != http.StatusOK || body["data"].(map[string]any)["mastered"].(bool) {
		t.Fatalf("retry wrong status=%d body=%v", status, body)
	}
	status, body = doJSON(t, r, http.MethodGet, "/api/v1/quiz/wrong-book/stats", token, nil)
	if stats := body["data"].(map[string]any); stats["unmastered"].(float64) != 1 {
		t.Fatalf("unmastered should return to 1: %v", stats)
	}
}

func TestQuizSubmitInvalidPayload(t *testing.T) {
	r, _ := setupQuizRouter(t)
	status, _ := doJSON(t, r, http.MethodPost, "/api/v1/quiz/submit", "", map[string]any{
		"answers": []map[string]any{},
	})
	if status != http.StatusBadRequest {
		t.Fatalf("empty answers want 400, got %d", status)
	}
}
