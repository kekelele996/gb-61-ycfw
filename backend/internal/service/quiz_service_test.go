package service

import (
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/model"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newQuizTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.QuizQuestion{}, &model.WrongQuestion{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	opts, _ := json.Marshal([]string{"甲", "乙", "丙", "丁"})
	questions := []model.QuizQuestion{
		{Question: "q1", Options: string(opts), Answer: 1, Explanation: "e1"},
		{Question: "q2", Options: string(opts), Answer: 2, Explanation: "e2"},
		{Question: "q3", Options: string(opts), Answer: 0, Explanation: "e3"},
	}
	if err := db.Create(&questions).Error; err != nil {
		t.Fatalf("seed questions: %v", err)
	}
	return db
}

func newQuizService(db *gorm.DB) *QuizService {
	return NewQuizService(
		repository.NewQuizQuestionRepository(db),
		repository.NewWrongQuestionRepository(db),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func ptrInt(i int) *int    { return &i }
func ptrUint(u uint) *uint { return &u }

func TestListQuestions(t *testing.T) {
	svc := newQuizService(newQuizTestDB(t))
	items, err := svc.ListQuestions()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 questions, got %d", len(items))
	}
	if items[0].Options[1].Text != "乙" || items[0].Answer != 1 {
		t.Fatalf("unexpected first question: %+v", items[0])
	}
}

func TestSubmitAnonymousSavesNothing(t *testing.T) {
	db := newQuizTestDB(t)
	svc := newQuizService(db)

	res, err := svc.Submit(0, []dto.QuizSubmitItem{
		{QuestionID: ptrUint(1), SelectedIndex: ptrInt(1)}, // correct
		{QuestionID: ptrUint(2), SelectedIndex: ptrInt(0)}, // wrong
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Score != 1 || res.Correct != 1 || res.WrongCount != 1 || res.Saved {
		t.Fatalf("unexpected graded result: %+v", res)
	}
	var count int64
	db.Model(&model.WrongQuestion{}).Count(&count)
	if count != 0 {
		t.Fatalf("anonymous submit must leave no records, got %d", count)
	}
}

func TestSubmitLoggedInAndRetryFlow(t *testing.T) {
	db := newQuizTestDB(t)
	svc := newQuizService(db)
	const user uint = 42

	// First full submission: q1 wrong, q2 wrong, q3 correct.
	res, err := svc.Submit(user, []dto.QuizSubmitItem{
		{QuestionID: ptrUint(1), SelectedIndex: ptrInt(0)},
		{QuestionID: ptrUint(2), SelectedIndex: ptrInt(1)},
		{QuestionID: ptrUint(3), SelectedIndex: ptrInt(0)},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Score != 1 || !res.Saved || len(res.WrongIDs) != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}

	stats, _ := svc.WrongBookStats(user)
	if stats.Total != 2 || stats.Unmastered != 2 || stats.Mastered != 0 || stats.Progress != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	// Resubmit the same quiz with q1 still wrong but a different option:
	// the single existing row must be updated, not duplicated.
	if _, err := svc.Submit(user, []dto.QuizSubmitItem{
		{QuestionID: ptrUint(1), SelectedIndex: ptrInt(3)},
		{QuestionID: ptrUint(2), SelectedIndex: ptrInt(2)}, // now correct -> untouched
	}); err != nil {
		t.Fatalf("second submit: %v", err)
	}
	var entry model.WrongQuestion
	if err := db.Where("user_id = ? AND question_id = ?", user, 1).First(&entry).Error; err != nil {
		t.Fatalf("load q1 entry: %v", err)
	}
	if entry.SelectedIndex != 3 || entry.Mastered {
		t.Fatalf("q1 entry should be updated to option 3 and stay unmastered: %+v", entry)
	}
	var total int64
	db.Model(&model.WrongQuestion{}).Where("user_id = ?", user).Count(&total)
	if total != 2 {
		t.Fatalf("one row per question expected, got %d", total)
	}

	// Retry q1 correctly -> mastered.
	retry, err := svc.RetryWrong(user, 1, 1)
	if err != nil || !retry.Correct || !retry.Mastered {
		t.Fatalf("retry correct failed: %+v err=%v", retry, err)
	}
	stats, _ = svc.WrongBookStats(user)
	if stats.Mastered != 1 || stats.Unmastered != 1 || stats.Progress != 50 {
		t.Fatalf("unexpected stats after mastery: %+v", stats)
	}

	// Retry q1 wrongly again -> selected option updated, back to unmastered.
	retry, err = svc.RetryWrong(user, 1, 2)
	if err != nil || retry.Correct || retry.Mastered {
		t.Fatalf("retry wrong failed: %+v err=%v", retry, err)
	}
	db.Where("user_id = ? AND question_id = ?", user, 1).First(&entry)
	if entry.SelectedIndex != 2 || entry.Mastered {
		t.Fatalf("q1 entry should reset to option 2 unmastered: %+v", entry)
	}

	// Wrong book joins question detail.
	book, err := svc.ListWrongBook(user)
	if err != nil || len(book) != 2 || book[0].Question == "" || len(book[0].Options) != 4 {
		t.Fatalf("wrong book unexpected: %d items err=%v", len(book), err)
	}
}

func TestRetryUnknownAndValidation(t *testing.T) {
	db := newQuizTestDB(t)
	svc := newQuizService(db)
	const user uint = 7

	if _, err := svc.RetryWrong(user, 999, 0); err == nil {
		t.Fatal("expected error for unknown question")
	}
	if _, err := svc.Submit(user, []dto.QuizSubmitItem{
		{QuestionID: ptrUint(1), SelectedIndex: ptrInt(9)},
	}); err == nil {
		t.Fatal("expected validation error for out-of-range option")
	}
	// Retry on a question not in the book is rejected.
	if _, err := svc.RetryWrong(user, 1, 1); err == nil {
		t.Fatal("expected not found when retrying a question absent from the book")
	}
}
