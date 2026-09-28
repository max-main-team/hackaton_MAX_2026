package scoring

import (
	"math"
	"strings"
)

type Vacancy struct {
	RequiredSkills      string
	MinExperienceMonths int
	City                string
	WorkFormat          string
}

type Resume struct {
	Skills           string
	ExperienceMonths int
	City             string
	WorkFormat       string
}

type Breakdown struct {
	Skills     int `json:"skills"`
	City       int `json:"city"`
	Schedule   int `json:"schedule"`
	Experience int `json:"experience"`
}

const (
	weightSkills     = 0.5
	weightCity       = 0.2
	weightSchedule   = 0.2
	weightExperience = 0.1
)

func tokenize(s string) map[string]struct{} {
	out := make(map[string]struct{})
	for part := range strings.SplitSeq(s, ",") {
		t := strings.ToLower(strings.TrimSpace(part))
		if t != "" {
			out[t] = struct{}{}
		}
	}
	return out
}

func Score(v Vacancy, r Resume) (int, Breakdown) {
	required := tokenize(v.RequiredSkills)
	owned := tokenize(r.Skills)

	skills := 1.0
	if len(required) > 0 {
		matched := 0
		for token := range required {
			if _, ok := owned[token]; ok {
				matched++
			}
		}
		skills = float64(matched) / float64(len(required))
	}

	vacancyCity := strings.ToLower(strings.TrimSpace(v.City))
	resumeCity := strings.ToLower(strings.TrimSpace(r.City))
	city := 0.0
	switch {
	case vacancyCity == resumeCity && vacancyCity != "":
		city = 1
	case vacancyCity == "" || resumeCity == "":
		city = 0.5
	}

	schedule := 0.0
	if strings.EqualFold(v.WorkFormat, r.WorkFormat) {
		schedule = 1
	}

	experience := 1.0
	if v.MinExperienceMonths > 0 {
		experience = min(1, max(0, float64(r.ExperienceMonths)/float64(v.MinExperienceMonths)))
	}

	breakdown := Breakdown{
		Skills:     int(math.Round(skills * 100)),
		City:       int(math.Round(city * 100)),
		Schedule:   int(math.Round(schedule * 100)),
		Experience: int(math.Round(experience * 100)),
	}

	total := math.Round(100 * (weightSkills*skills + weightCity*city +
		weightSchedule*schedule + weightExperience*experience))
	return int(total), breakdown
}
