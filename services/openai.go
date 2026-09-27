package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openAIEmbeddingsURL = "https://api.openai.com/v1/embeddings"

// OpenAIClient calls the OpenAI embeddings endpoint directly
type OpenAIClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenAIClient(apiKey string, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 2 * time.Minute},
	}
}

type embeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingsResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// Embed returns one vector per input, in the same order as inputs
func (o *OpenAIClient) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body, err := json.Marshal(embeddingsRequest{Model: o.model, Input: inputs})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEmbeddingsURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embeddings: %s: %s", resp.Status, msg)
	}

	var embResp embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}
	if len(embResp.Data) != len(inputs) {
		return nil, fmt.Errorf("openai embeddings: got %d embeddings for %d inputs", len(embResp.Data), len(inputs))
	}

	embeddings := make([][]float32, len(inputs))
	for _, d := range embResp.Data {
		if d.Index < 0 || d.Index >= len(inputs) {
			return nil, fmt.Errorf("openai embeddings: index %d out of range", d.Index)
		}
		embeddings[d.Index] = d.Embedding
	}
	return embeddings, nil
}
