package scoring

import (
	"bytes"
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

type AIResult struct {
	Score   int
	Comment string
}

func NewAIClient(apiKey, baseURL, model string) *AIClient {
	return &AIClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *AIClient) Enabled() bool { return c != nil && c.apiKey != "" }

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

// ScoreAI оценивает соответствие резюме вакансии через LLM
// (OpenAI-совместимый API: z.ai / bigmodel / любой другой).
func (c *AIClient) ScoreAI(vacancy, resume any) (AIResult, error) {
	if !c.Enabled() {
		return AIResult{}, fmt.Errorf("AI is not configured")
	}

	payload, err := json.Marshal(struct {
		Vacancy any `json:"vacancy"`
		Resume  any `json:"resume"`
	}{vacancy, resume})
	if err != nil {
		return AIResult{}, err
	}

	prompt := "Ты — HR-аналитик. Оцени соответствие резюме вакансии. " +
		"Верни СТРОГО валидный JSON без markdown: {\"score\": 0-100, \"comment\": \"краткое обоснование на русском\"}. " +
		"Данные:\n" + string(payload)

	reqBody, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Отвечай только валидным JSON без пояснений."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return AIResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return AIResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return AIResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return AIResult{}, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(body))
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return AIResult{}, err
	}
	if len(cr.Choices) == 0 {
		return AIResult{}, fmt.Errorf("empty choices")
	}

	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	content = strings.Trim(content, "`")
	content = strings.TrimSpace(strings.TrimPrefix(content, "json"))

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
