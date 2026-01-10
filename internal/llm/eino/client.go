package eino

import (
	"context"
	"errors"
	"io"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
)

// Config Eino 客户端配置
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Client Eino LLM 客户端
type Client struct {
	chatModel *ark.ChatModel
	config    *Config
}

// NewClient 创建 Eino 客户端
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		chatModel: chatModel,
		config:    cfg,
	}, nil
}

// Generate 生成回复
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	msgs := []*schema.Message{
		{Role: schema.User, Content: prompt},
	}

	resp, err := c.chatModel.Generate(ctx, msgs)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateWithMessages 使用消息列表生成回复
func (c *Client) GenerateWithMessages(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return c.chatModel.Generate(ctx, messages)
}

// GenerateStream 流式生成
func (c *Client) GenerateStream(ctx context.Context, prompt string, callback func(chunk string)) error {
	msgs := []*schema.Message{
		{Role: schema.User, Content: prompt},
	}

	stream, err := c.chatModel.Stream(ctx, msgs)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		callback(chunk.Content)
	}

	return nil
}

// StreamWithMessages 使用消息列表流式生成
func (c *Client) StreamWithMessages(ctx context.Context, messages []*schema.Message, callback func(chunk *schema.Message)) error {
	stream, err := c.chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		callback(chunk)
	}

	return nil
}

// BindTools 绑定工具
func (c *Client) BindTools(tools []*schema.ToolInfo) error {
	return c.chatModel.BindTools(tools)
}

// GetChatModel 获取底层 ChatModel
func (c *Client) GetChatModel() *ark.ChatModel {
	return c.chatModel
}
