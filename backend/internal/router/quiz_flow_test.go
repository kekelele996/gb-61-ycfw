package router_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/router"
	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

type apiEnvelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func setupQuizRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	// A named in-memory database scoped to one test function.
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.PlantSpecies{}, &model.CareArticle{}, &model.DiseasePest{},
		&model.CareReminder{}, &model.Favorite{}, &model.UserGarden{},
		&model.Question{}, &model.Answer{}, &model.QuizQuestion{}, &model.WrongQuestion{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	questions := []model.QuizQuestion{
		{Question: "q1", Options: `["a","b","c","d"]`, AnswerIndex: 1, Explanation: "e1"},
		{Question: "q2", Options: `["a","b","c","d"]`, AnswerIndex: 2, Explanation: "e2"},
	}
	if err := db.Create(&questions).Error; err != nil {
		t.Fatalf("seed questions: %v", err)
	}
	user := &model.User{Username: "tester", Email: "t@e.com", Role: "user"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cfg := &config.Config{
		JWTSecret:    "test-secret",
		JWTExpire:    time.Hour,
		RateLimitReq: 1000,
		RateLimitWin: time.Minute,
		UploadDir:    t.TempDir(),
		CORSOrigins:  "*",
		ServerPort:   "8080",
	}
	return router.Setup(cfg, db, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))), db
}

func doJSON(t *testing.T, r *gin.Engine, method, path, token string, body interface{}) (apiEnvelope, *httptest.ResponseRecorder) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("%s %s -> %d: %s", method, path, w.Code, w.Body.String())
	}
	var env apiEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, w.Body.String())
	}
	return env, w
}

func TestQuizFullFlow(t *testing.T) {
	r, _ := setupQuizRouter(t)
	token, err := util.GenerateToken(1, "tester", "user", "test-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Question bank is public and must not leak the correct answer.
	env, _ := doJSON(t, r, "GET", "/api/v1/quiz/questions", "", nil)
	var bank []struct {
		ID          uint     `json:"id"`
		AnswerIndex int      `json:"answer_index"`
		Options     []string `json:"options"`
	}
	if err := json.Unmarshal(env.Data, &bank); err != nil {
		t.Fatal(err)
	}
	if len(bank) != 2 || bank[0].AnswerIndex != 0 || len(bank[0].Options) != 4 {
		t.Fatalf("unexpected question bank: %+v", bank)
	}

	// 2. Anonymous submit is graded but leaves no learning record.
	anonBody := map[string]interface{}{"answers": []map[string]interface{}{
		{"question_id": 1, "selected_index": 0}, // wrong
		{"question_id": 2, "selected_index": 2}, // correct
	}}
	env, _ = doJSON(t, r, "POST", "/api/v1/quiz/submit", "", anonBody)
	var anonResult struct {
		Total   int `json:"total"`
		Correct int `json:"correct"`
	}
	if err := json.Unmarshal(env.Data, &anonResult); err != nil {
		t.Fatal(err)
	}
	if anonResult.Total != 2 || anonResult.Correct != 1 {
		t.Fatalf("anonymous grading wrong: %+v", anonResult)
	}

	// 3. Logged-in submit: q1 wrong -> wrong-question book gains one entry.
	loginBody := map[string]interface{}{"answers": []map[string]interface{}{
		{"question_id": 1, "selected_index": 0}, // wrong (correct is 1)
		{"question_id": 2, "selected_index": 2}, // correct
	}}
	env, _ = doJSON(t, r, "POST", "/api/v1/quiz/submit", token, loginBody)
	var result struct {
		Total   int `json:"total"`
		Correct int `json:"correct"`
	}
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Correct != 1 {
		t.Fatalf("expected 1 correct, got %d", result.Correct)
	}

	// 4. Stats show 1 total, 1 unmastered, 0 mastered.
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions/stats", token, nil)
	var stats struct {
		Total      int64 `json:"total"`
		Unmastered int64 `json:"unmastered"`
		Mastered   int64 `json:"mastered"`
	}
	if err := json.Unmarshal(env.Data, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.Total != 1 || stats.Unmastered != 1 || stats.Mastered != 0 {
		t.Fatalf("unexpected stats after submit: %+v", stats)
	}

	// 5. Wrong-question list returns question content and the chosen index.
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions", token, nil)
	var wrong []struct {
		QuestionID    uint     `json:"question_id"`
		SelectedIndex int      `json:"selected_index"`
		Mastered      bool     `json:"mastered"`
		Options       []string `json:"options"`
		AnswerIndex   int      `json:"answer_index"`
	}
	if err := json.Unmarshal(env.Data, &wrong); err != nil {
		t.Fatal(err)
	}
	if len(wrong) != 1 || wrong[0].QuestionID != 1 || wrong[0].SelectedIndex != 0 || wrong[0].Mastered ||
		len(wrong[0].Options) != 4 || wrong[0].AnswerIndex != 1 {
		t.Fatalf("unexpected wrong list: %+v", wrong)
	}

	// 6. Retry wrong again -> selected index updates, stays unmastered, still one row.
	env, _ = doJSON(t, r, "PUT", "/api/v1/quiz/wrong-questions/1/retry", token,
		map[string]interface{}{"selected_index": 3})
	var retryRes struct {
		Correct bool `json:"correct"`
	}
	if err := json.Unmarshal(env.Data, &retryRes); err != nil {
		t.Fatal(err)
	}
	if retryRes.Correct {
		t.Fatal("retry with index 3 should be wrong")
	}
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions", token, nil)
	if err := json.Unmarshal(env.Data, &wrong); err != nil {
		t.Fatal(err)
	}
	if len(wrong) != 1 || wrong[0].SelectedIndex != 3 || wrong[0].Mastered {
		t.Fatalf("entry should update to index 3, still unmastered, single row: %+v", wrong)
	}

	// 7. Retry correctly -> mastered, stats update.
	env, _ = doJSON(t, r, "PUT", "/api/v1/quiz/wrong-questions/1/retry", token,
		map[string]interface{}{"selected_index": 1})
	if err := json.Unmarshal(env.Data, &retryRes); err != nil {
		t.Fatal(err)
	}
	if !retryRes.Correct {
		t.Fatal("retry with index 1 should be correct")
	}
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions/stats", token, nil)
	if err := json.Unmarshal(env.Data, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.Total != 1 || stats.Unmastered != 0 || stats.Mastered != 1 {
		t.Fatalf("expected 1 mastered, got %+v", stats)
	}

	// 8. Mastered then wrong again in a fresh quiz -> back to unmastered, still one row.
	env, _ = doJSON(t, r, "POST", "/api/v1/quiz/submit", token,
		map[string]interface{}{"answers": []map[string]interface{}{
			{"question_id": 1, "selected_index": 2},
		}})
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatal(err)
	}
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions", token, nil)
	if err := json.Unmarshal(env.Data, &wrong); err != nil {
		t.Fatal(err)
	}
	if len(wrong) != 1 || wrong[0].SelectedIndex != 2 || wrong[0].Mastered {
		t.Fatalf("entry should be unmastered again with index 2, single row: %+v", wrong)
	}
	env, _ = doJSON(t, r, "GET", "/api/v1/quiz/wrong-questions/stats", token, nil)
	if err := json.Unmarshal(env.Data, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.Unmastered != 1 || stats.Mastered != 0 {
		t.Fatalf("expected 1 unmastered 0 mastered, got %+v", stats)
	}
}

func TestQuizWrongBookRequiresAuth(t *testing.T) {
	r, _ := setupQuizRouter(t)
	req := httptest.NewRequest("GET", "/api/v1/quiz/wrong-questions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestQuizRetryUnknownQuestion(t *testing.T) {
	r, _ := setupQuizRouter(t)
	token, _ := util.GenerateToken(1, "tester", "user", "test-secret", time.Hour)
	req := httptest.NewRequest("PUT", "/api/v1/quiz/wrong-questions/99/retry",
		bytes.NewBufferString(`{"selected_index":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown wrong question, got %d: %s", w.Code, w.Body.String())
	}
}
