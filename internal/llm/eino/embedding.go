package eino

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient 向量嵌入客户端
type EmbeddingClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// EmbeddingConfig 嵌入配置
type EmbeddingConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewEmbeddingClient 创建嵌入客户端
func NewEmbeddingClient(cfg *EmbeddingConfig) *EmbeddingClient {
	return &EmbeddingClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

// Embed 单文本嵌入
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float64, error) {
	vectors, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return vectors[0], nil
}

// EmbedBatch 批量文本嵌入
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := embeddingRequest{
		Model: c.model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := c.baseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed: %s, body: %s", resp.Status, string(body))
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vectors := make([][]float64, len(result.Data))
	for _, d := range result.Data {
		vectors[d.Index] = d.Embedding
	}

	return vectors, nil
}
