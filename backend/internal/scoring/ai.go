package scoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewAIClient(apiKey, baseURL, model string) *AIClient {
	return &AIClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 180 * time.Second},
	}
}

func (c *AIClient) Enabled() bool { return c != nil && c.apiKey != "" }

type AIResult struct {
	Score   int
	Comment string
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// chat — один запрос к OpenAI-совместимому /chat/completions,
// возвращает очищенный текст ответа модели.
func (c *AIClient) chat(ctx context.Context, system, user string) (string, error) {
	reqBody, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm status %d: %s", resp.StatusCode, string(body))
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	content = strings.Trim(content, "`")
	content = strings.TrimSpace(strings.TrimPrefix(content, "json"))
	return content, nil
}

// ScoreAI оценивает соответствие резюме вакансии через LLM.
func (c *AIClient) ScoreAI(ctx context.Context, vacancy, resume any) (AIResult, error) {
	content, err := c.chat(ctx,
		"Ты — HR-аналитик. Оцени соответствие резюме вакансии. Верни СТРОГО валидный JSON без markdown: {\"score\": 0-100, \"comment\": \"краткое обоснование на русском\"}.",
		fmt.Sprintf("ВАКАНСИЯ: %v\n\nРЕЗЮМЕ: %v", vacancy, resume))
	if err != nil {
		return AIResult{}, err
	}

	var result struct {
		Score   int    `json:"score"`
		Comment string `json:"comment"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return AIResult{}, fmt.Errorf("parse llm answer: %w", err)
	}
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}
	return AIResult{Score: result.Score, Comment: result.Comment}, nil
}

const extractSystemPrompt = `Ты — HR-парсер резюме. Извлеки из текста структурированные данные и верни СТРОГО валидный JSON без markdown ровно по схеме:
{"title": string, "skills": string, "experience_months": int, "about": string, "education": string, "city": string, "work_format": "onsite|hybrid|remote", "employment_type": "full_time|part_time|contract|internship", "salary_min": int|null, "salary_max": int|null, "notes": string}
Правила:
- skills — через запятую, как в резюме;
- experience_months — суммарный опыт в месяцах (годы работы умножай на 12);
- зарплата — рублей в месяц, только числа;
- work_format/employment_type — только перечисленные значения, ближайшие к тексту;
- ничего не выдумывай: нет данных в тексте — пустая строка / 0 / null;
- notes — по-русски, 1-2 предложения: что извлекли и чего не нашли.`

type ResumeDraft struct {
	Title            string `json:"title"`
	Skills           string `json:"skills"`
	ExperienceMonths int    `json:"experience_months"`
	About            string `json:"about"`
	Education        string `json:"education"`
	City             string `json:"city"`
	WorkFormat       string `json:"work_format"`
	EmploymentType   string `json:"employment_type"`
	SalaryMin        *int   `json:"salary_min"`
	SalaryMax        *int   `json:"salary_max"`
	Notes            string `json:"notes"`
}

// ExtractResume структурирует текст резюме в черновик через LLM.
func (c *AIClient) ExtractResume(ctx context.Context, text string) (ResumeDraft, error) {
	if !c.Enabled() {
		return ResumeDraft{}, fmt.Errorf("AI is not configured")
	}

	content, err := c.chat(ctx, extractSystemPrompt, "ТЕКСТ РЕЗЮМЕ:\n"+text)
	if err != nil {
		return ResumeDraft{}, err
	}
	return ParseDraftJSON(content)
}

func ParseDraftJSON(content string) (ResumeDraft, error) {
	var d ResumeDraft
	i := strings.Index(content, "{")
	j := strings.LastIndex(content, "}")
	if i < 0 || j <= i {
		return d, fmt.Errorf("no JSON object in llm answer")
	}
	if err := json.Unmarshal([]byte(content[i:j+1]), &d); err != nil {
		return d, fmt.Errorf("parse llm answer: %w", err)
	}
	return NormalizeResumeDraft(d), nil
}

var workFormatAliases = map[string]string{
	"удалённо": "remote", "удаленно": "remote", "удалёнка": "remote", "удаленка": "remote",
	"дистанционно": "remote", "remote": "remote", "из дома": "remote",
	"гибрид": "hybrid", "гибридный график": "hybrid", "hybrid": "hybrid",
	"офис": "onsite", "в офисе": "onsite", "onsite": "onsite",
}

var employmentTypeAliases = map[string]string{
	"полный день": "full_time", "полная занятость": "full_time", "полная": "full_time",
	"фуллтайм": "full_time", "full_time": "full_time",
	"частичная занятость": "part_time", "частичная": "part_time", "part_time": "part_time",
	"контракт": "contract", "договор": "contract", "contract": "contract",
	"стажировка": "internship", "internship": "internship",
}

// NormalizeResumeDraft приводит ответ LLM к нашим enum'ам и границам.
func NormalizeResumeDraft(d ResumeDraft) ResumeDraft {
	d.Title = strings.TrimSpace(d.Title)
	d.Skills = normalizeSkills(d.Skills)
	d.ExperienceMonths = clampInt(d.ExperienceMonths, 0, 600)
	d.About = strings.TrimSpace(d.About)
	d.Education = strings.TrimSpace(d.Education)
	d.City = strings.TrimSpace(d.City)
	d.WorkFormat = normalizeEnum(d.WorkFormat, workFormatAliases, "onsite")
	d.EmploymentType = normalizeEnum(d.EmploymentType, employmentTypeAliases, "full_time")
	if d.SalaryMin != nil && *d.SalaryMin < 0 {
		d.SalaryMin = nil
	}
	if d.SalaryMax != nil && *d.SalaryMax < 0 {
		d.SalaryMax = nil
	}
	if d.SalaryMin != nil && d.SalaryMax != nil && *d.SalaryMin > *d.SalaryMax {
		d.SalaryMin, d.SalaryMax = d.SalaryMax, d.SalaryMin
	}
	d.Notes = strings.TrimSpace(d.Notes)
	return d
}

func normalizeSkills(s string) string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, ", ")
}

func normalizeEnum(v string, aliases map[string]string, def string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return def
	}
	if mapped, ok := aliases[v]; ok {
		return mapped
	}
	for alias, canonical := range aliases {
		if strings.Contains(v, alias) {
			return canonical
		}
	}
	return def
}

func clampInt(v, lo, hi int) int {
	return min(hi, max(lo, v))
}
