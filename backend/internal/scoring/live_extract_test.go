package scoring

import (
	"context"
	"testing"
	"time"
)

func TestExtractResumeLive(t *testing.T) {
	ai := NewAIClient("3c4d8334b980410b8dc0b21c71a87ceb.cT9SM1klD9lgdPpo", "https://api.z.ai/api/paas/v4", "glm-4.5-flash")
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
