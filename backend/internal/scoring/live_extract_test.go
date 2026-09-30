package scoring

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestExtractResumeLive(t *testing.T) {
	if os.Getenv("AI_LIVE_TEST") == "" {
		t.Skip("live AI check: set AI_LIVE_TEST=1")
	}
	key := os.Getenv("AI_API_KEY")
	if key == "" {
		key = "3c4d8334b980410b8dc0b21c71a87ceb.cT9SM1klD9lgdPpo"
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "glm-4.5-flash"
	}
	ai := NewAIClient(key, "https://api.z.ai/api/paas/v4", model)
	text := "Михаил Соковых\nBackend-разработчик\nОпыт 5 лет: Go, PostgreSQL, Redis, Docker, Kubernetes. Живу в Санкт-Петербурге.\nОбразование: ИТМО. Полный день, удалёнка. Зарплата от 250 тысяч рублей."
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	start := time.Now()
	d, err := ai.ExtractResume(ctx, text)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	t.Logf("результат за %.1fs: title=%q skills=%q exp=%d city=%q wf=%q et=%q",
		time.Since(start).Seconds(), d.Title, d.Skills, d.ExperienceMonths, d.City, d.WorkFormat, d.EmploymentType)
	if d.Title == "" && d.Skills == "" {
		t.Error("пустой черновик — извлечение не сработало")
	}
}
