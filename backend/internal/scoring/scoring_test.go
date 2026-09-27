package scoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreFullMatch(t *testing.T) {
	v := Vacancy{RequiredSkills: "go, postgres", MinExperienceMonths: 12, City: "Москва", WorkFormat: "remote"}
	r := Resume{Skills: "go, postgres", ExperienceMonths: 24, City: "Москва", WorkFormat: "remote"}

	score, breakdown := Score(v, r)
	assert.Equal(t, 100, score)
	assert.Equal(t, Breakdown{Skills: 100, City: 100, Schedule: 100, Experience: 100}, breakdown)
}

func TestScoreEmptyResume(t *testing.T) {
	v := Vacancy{RequiredSkills: "go, postgres", MinExperienceMonths: 12, City: "Москва", WorkFormat: "remote"}

	score, _ := Score(v, Resume{})
	assert.Equal(t, 10, score)
}

func TestScorePartialSkills(t *testing.T) {
	v := Vacancy{RequiredSkills: "go, postgres, kafka, docker", MinExperienceMonths: 0,
		City: "Санкт-Петербург", WorkFormat: "onsite"}
	r := Resume{Skills: "go, sql", ExperienceMonths: 24, City: "санкт-петербург", WorkFormat: "onsite"}

	score, breakdown := Score(v, r)
	assert.Equal(t, 25, breakdown.Skills)
	// 0.5*25 + 0.2*100 + 0.2*100 + 0.1*100 = 62.5 -> 63
	assert.Equal(t, 63, score)
}

func TestScoreCityMismatch(t *testing.T) {
	v := Vacancy{City: "Москва", WorkFormat: "onsite"}
	r := Resume{City: "Казань", WorkFormat: "remote"}

	score, breakdown := Score(v, r)
	// skills: нет требований -> 1.0 (50), city 0, schedule 0, experience 100 (10)
	assert.Equal(t, 60, score)
	assert.Equal(t, Breakdown{Skills: 100, City: 0, Schedule: 0, Experience: 100}, breakdown)
}

func TestScoreEmptyRequirements(t *testing.T) {
	v := Vacancy{}
	r := Resume{City: "Москва", WorkFormat: "full_time"}

	score, breakdown := Score(v, r)
	// skills 50 + city 10 (одно поле пустое) + schedule 0 + experience 10 = 70
	assert.Equal(t, 70, score)
	assert.Equal(t, Breakdown{Skills: 100, City: 50, Schedule: 0, Experience: 100}, breakdown)
}
