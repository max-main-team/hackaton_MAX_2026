package scoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeResumeDraftAliases(t *testing.T) {
	d := NormalizeResumeDraft(ResumeDraft{
		Title:            "  Backend-разработчик  ",
		Skills:           " go , postgres , , kafka ",
		ExperienceMonths: 900,
		WorkFormat:       "Удалённо",
		EmploymentType:   "полный день",
		City:             "  Санкт-Петербург ",
		SalaryMin:        ptrInt(300),
		SalaryMax:        ptrInt(100),
	})
	assert.Equal(t, "Backend-разработчик", d.Title)
	assert.Equal(t, "go, postgres, kafka", d.Skills)
	assert.Equal(t, 600, d.ExperienceMonths)
	assert.Equal(t, "remote", d.WorkFormat)
	assert.Equal(t, "full_time", d.EmploymentType)
	assert.Equal(t, "Санкт-Петербург", d.City)
	assert.Equal(t, 100, *d.SalaryMin)
	assert.Equal(t, 300, *d.SalaryMax)
}

func TestNormalizeResumeDraftUnknownEnum(t *testing.T) {
	d := NormalizeResumeDraft(ResumeDraft{WorkFormat: "на марсе", EmploymentType: "стажировка"})
	assert.Equal(t, "onsite", d.WorkFormat)
	assert.Equal(t, "internship", d.EmploymentType)
}

func TestNormalizeResumeDraftNegativeSalaryDropped(t *testing.T) {
	d := NormalizeResumeDraft(ResumeDraft{SalaryMin: ptrInt(-5)})
	assert.Nil(t, d.SalaryMin)
}

func TestParseDraftJSONWithFences(t *testing.T) {
	content := "Ответ:\n```json\n{\"title\": \"QA\", \"experience_months\": 12, \"notes\": \"ок\"}\n```"
	d, err := ParseDraftJSON(content)
	require.NoError(t, err)
	assert.Equal(t, "QA", d.Title)
	assert.Equal(t, 12, d.ExperienceMonths)
	assert.Equal(t, "ок", d.Notes)
}

func TestParseDraftJSONNoObject(t *testing.T) {
	_, err := ParseDraftJSON("никакого json тут нет")
	assert.Error(t, err)
}

func ptrInt(i int) *int { return &i }
