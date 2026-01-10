package eino

import (
	"context"

	"github.com/NuyoahCh/eocall/internal/llm"
)

// Config Eino 配置
type Config = llm.Config

// Client Eino LLM 客户端
// 基于 ByteDance Eino 框架实现
type Client struct {
	config *llm.Config
	// TODO: 添加 Eino 相关字段
}

// NewClient 创建 Eino 客户端
func NewClient(cfg *Config) (*Client, error) {
	return &Client{
		config: cfg,
	}, nil
}

// Chat 聊天
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// TODO: 实现 Eino 调用
	// 参考: https://github.com/cloudwego/eino
	return &llm.ChatResponse{
		Content: "Eino client not implemented yet",
	}, nil
}

// ChatStream 流式聊天
func (c *Client) ChatStream(ctx context.Context, req *llm.ChatRequest, callback func(chunk string)) error {
	// TODO: 实现 Eino 流式调用
	callback("Eino streaming not implemented yet")
	return nil
}

// Generate 简单生成 (实现 planner.LLMClient 接口)
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	resp, err := c.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateStream 流式生成 (实现 planner.LLMClient 接口)
func (c *Client) GenerateStream(ctx context.Context, prompt string, callback func(chunk string)) error {
	return c.ChatStream(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}, callback)
}
