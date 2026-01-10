package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	LLM     LLMConfig     `yaml:"llm"`
	RAG     RAGConfig     `yaml:"rag"`
	Session SessionConfig `yaml:"session"`
	Log     LogConfig     `yaml:"log"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LLMConfig 大模型配置
type LLMConfig struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// RAGConfig RAG 检索配置
type RAGConfig struct {
	VectorDB     VectorDBConfig `yaml:"vector_db"`
	ChunkSize    int            `yaml:"chunk_size"`
	ChunkOverlap int            `yaml:"chunk_overlap"`
	TopK         int            `yaml:"top_k"`
}

// VectorDBConfig 向量数据库配置
type VectorDBConfig struct {
	Type     string `yaml:"type"` // milvus, qdrant, etc.
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	MaxHistory      int           `yaml:"max_history"`      // 最大历史轮数
	SummaryAfter    int           `yaml:"summary_after"`    // 多少轮后开始摘要
	TTL             time.Duration `yaml:"ttl"`              // 会话过期时间
	CleanupInterval time.Duration `yaml:"cleanup_interval"` // 清理间隔
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.setDefaults()
	return cfg, nil
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Session.MaxHistory == 0 {
		c.Session.MaxHistory = 50
	}
	if c.Session.SummaryAfter == 0 {
		c.Session.SummaryAfter = 20
	}
	if c.Session.TTL == 0 {
		c.Session.TTL = 30 * time.Minute
	}
	if c.RAG.TopK == 0 {
		c.RAG.TopK = 5
	}
}
