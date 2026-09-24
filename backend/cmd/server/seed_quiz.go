package main

import (
	"encoding/json"
	"log/slog"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// seedQuizQuestions populates the fixed care quiz bank. It is idempotent so
// the bank is also created for databases initialized before this feature.
func seedQuizQuestions(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.QuizQuestion{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	bank := []struct {
		question    string
		options     []string
		answer      int
		explanation string
	}{
		{"以下哪种植物属于多肉植物？", []string{"月季", "吉娃娃", "碗莲", "龟背竹"}, 1, "吉娃娃为景天科拟石莲属多肉植物。"},
		{"多肉植物夏季施肥的原则是？", []string{"薄肥勤施", "大量施肥", "停止施肥", "只施氮肥"}, 2, "夏季高温多数多肉休眠，应停止施肥避免肥害。"},
		{"月季黑斑病的典型症状是？", []string{"叶片白粉", "黑色圆形斑点", "叶背蛛网", "叶片卷曲"}, 1, "黑斑病叶片出现黑色圆形斑点，边缘放射状。"},
		{"龟背竹适合的光照条件是？", []string{"全日照", "散射光", "完全黑暗", "强直射光"}, 1, "龟背竹耐阴，适合明亮散射光环境。"},
		{"换盆的最佳季节通常是？", []string{"夏季", "深冬", "春季", "雨季"}, 2, "春季气温回升、根系活跃，是换盆最佳时机。"},
		{"“见干见湿”的浇水原则适用于？", []string{"所有植物", "多肉植物", "绝大多数盆栽植物", "水生植物"}, 2, "绝大多数盆栽植物遵循见干见湿原则。"},
	}

	questions := make([]model.QuizQuestion, 0, len(bank))
	for _, b := range bank {
		raw, err := json.Marshal(b.options)
		if err != nil {
			return err
		}
		questions = append(questions, model.QuizQuestion{
			Question:    b.question,
			Options:     string(raw),
			Answer:      b.answer,
			Explanation: b.explanation,
		})
	}
	if err := db.Create(&questions).Error; err != nil {
		return err
	}
	slog.Default().Info("quiz question seed data created", "questions", len(questions))
	return nil
}
