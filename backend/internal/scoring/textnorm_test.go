package scoring

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const hhStyleResume = `Анна Смирнова
Frontend-разработчик

Опыт работы — 4 года 2 месяца

Январь 2022 — настоящее время
ООО «Ромашка», Санкт-Петербург
Frontend-разработчик
Разработала кабинет клиента: React, TypeScript, Redux Toolkit.
- 1 -

Ноябрь 2019 — Декабрь 2021
Студия «Цифра»
Верстальщик → Junior frontend
HTML, CSS, JavaScript

Ключевые навыки
React
TypeScript
Vite
CSS Grid

Образование: СПбГУТ им. Бонч-Бруевича, программная инженерия.
Занятость: полная занятость, желательно удалённая работа.
Зарплата: от 180 000 руб.
Тел.: +7 999 000-00-00
anna@example.com
`

const freeformResume = `Привет! Меня зовут Пётр, ищу работу бэкендером на Go.
Живу в Казани, готов к релокации в Питер.
Стек: go, postgres, redis, Go, docker, gRPC.
Опыт: 2 года в стартапе, до этого год интерном.
Формат — гибрид, оформление проектная работа.
Вилка от 200к до 300к руб/мес.
Окончил КФУ, прикладная математика.
Страница 3 из 7
`

func TestNormalizeResumeTextHHStyle(t *testing.T) {
	out := NormalizeResumeText(hhStyleResume)
	assert.NotContains(t, out, "- 1 -")
	assert.NotContains(t, out, "Тел.:")
	assert.NotContains(t, out, "anna@example.com")
	assert.Contains(t, out, "React, TypeScript, Redux Toolkit")
	assert.Contains(t, out, "Ключевые навыки")
	assert.Contains(t, out, "СПбГУТ")
	lines := strings.Split(out, "\n")
	assert.Greater(t, len(lines), 5)
}

func TestNormalizeResumeTextFreeform(t *testing.T) {
	out := NormalizeResumeText(freeformResume)
	assert.NotContains(t, out, "Страница 3 из 7")
	assert.Contains(t, out, "go, postgres, redis, Go, docker, gRPC")
	assert.Contains(t, out, "гибрид")
}

func TestNormalizeResumeTextDeHyphenates(t *testing.T) {
	in := "backend-\nразработчик, senior-\ngo\nСПб-\nГУТ"
	out := NormalizeResumeText(in)
	assert.Contains(t, out, "backend-разработчик")
	assert.Contains(t, out, "senior-go")
	assert.Contains(t, out, "СПб-\nГУТ")
}

func TestNormalizeResumeTextCollapsesBlankLines(t *testing.T) {
	out := NormalizeResumeText("a\n\n\n\n\nb\n\n\nc")
	assert.Equal(t, "a\n\nb\n\nc", out)
}

func TestNormalizeResumeTextEmpty(t *testing.T) {
	assert.Equal(t, "", NormalizeResumeText("   \n \n\t\n"))
}

func TestTruncateTextRuneBoundary(t *testing.T) {
	s := "резюме на русском"
	cut := TruncateText(s, 5)
	assert.Equal(t, "резюм", cut)
	assert.Equal(t, s, TruncateText(s, 1000))
	assert.Equal(t, "", TruncateText("", 10))
}

func TestNormalizeResumeDraftSkillsDedup(t *testing.T) {
	d := NormalizeResumeDraft(ResumeDraft{
		Skills: "Go, go, GO,  go , postgres, , Kubernetes",
		About:  strings.Repeat("x", 9000),
	})
	assert.Equal(t, "Go, postgres, Kubernetes", d.Skills)
	assert.Len(t, d.About, 4000)
}
