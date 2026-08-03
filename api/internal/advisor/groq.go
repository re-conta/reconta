package advisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const groqEndpoint = "https://api.groq.com/openai/v1/chat/completions"

// defaultModel é um modelo leve e rápido disponível no plano gratuito do
// Groq, adequado para gerar texto curto como estas recomendações.
const defaultModel = "llama-3.1-8b-instant"

// GroqClient fala com a API de chat completions do Groq (compatível com o
// formato da OpenAI). Não faz nenhum controle de taxa por conta própria —
// isso é responsabilidade exclusiva da Queue, para que exista um único ponto
// de controle de quota no processo inteiro.
type GroqClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGroqClient(apiKey, model string) *GroqClient {
	if model == "" {
		model = defaultModel
	}
	return &GroqClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	Temperature    float64        `json:"temperature"`
	MaxTokens      int            `json:"max_completion_tokens"`
	ResponseFormat responseFormat `json:"response_format"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GetRecommendations pede ao Groq recomendações a partir do prompt informado
// e retorna a lista já sanitizada. Faz exatamente UMA chamada HTTP — o
// chamador (Queue) é quem decide quando é seguro chamar este método.
func (c *GroqClient) GetRecommendations(ctx context.Context, userPrompt string) ([]Recommendation, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.6,
		MaxTokens:      1200,
		ResponseFormat: responseFormat{Type: "json_object"},
	})
	if err != nil {
		return nil, fmt.Errorf("serializando requisição ao groq: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("criando requisição ao groq: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chamando groq: %w", err)
	}
	defer resp.Body.Close()

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decodificando resposta do groq: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("status %d", resp.StatusCode)
		if cr.Error != nil && cr.Error.Message != "" {
			msg = cr.Error.Message
		}
		return nil, fmt.Errorf("groq retornou erro: %s", msg)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("groq retornou resposta sem conteúdo")
	}

	var wrapper struct {
		Recommendations []Recommendation `json:"recommendations"`
	}
	if err := json.Unmarshal([]byte(cr.Choices[0].Message.Content), &wrapper); err != nil {
		return nil, fmt.Errorf("interpretando recomendações do groq: %w", err)
	}

	return sanitizeRecommendations(wrapper.Recommendations), nil
}
