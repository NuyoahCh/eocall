package rag

import (
	"context"
	"fmt"
	"strings"
)

// Document 文档
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Score    float64           `json:"score"`
}

// Embedder 向量嵌入接口
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// VectorStore 向量存储接口
type VectorStore interface {
	Insert(ctx context.Context, docs []Document, vectors [][]float64) error
	Search(ctx context.Context, vector []float64, topK int) ([]Document, error)
	Delete(ctx context.Context, ids []string) error
}

// Reranker 重排序接口
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []Document, topK int) ([]Document, error)
}

// Service RAG 服务
type Service struct {
	embedder     Embedder
	vectorStore  VectorStore
	reranker     Reranker
	chunkSize    int
	chunkOverlap int
}

// Config RAG 配置
type Config struct {
	ChunkSize    int
	ChunkOverlap int
}

// NewService 创建 RAG 服务
func NewService(embedder Embedder, vectorStore VectorStore, reranker Reranker, cfg *Config) *Service {
	return &Service{
		embedder:     embedder,
		vectorStore:  vectorStore,
		reranker:     reranker,
		chunkSize:    cfg.ChunkSize,
		chunkOverlap: cfg.ChunkOverlap,
	}
}

// IndexDocument 索引文档
func (s *Service) IndexDocument(ctx context.Context, content string, metadata map[string]string) error {
	// 1. 分块
	chunks := s.splitIntoChunks(content)

	// 2. 生成向量
	vectors, err := s.embedder.EmbedBatch(ctx, chunks)
	if err != nil {
		return fmt.Errorf("embed failed: %w", err)
	}

	// 3. 构建文档
	docs := make([]Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = Document{
			ID:       fmt.Sprintf("%s_%d", metadata["source"], i),
			Content:  chunk,
			Metadata: metadata,
		}
	}

	// 4. 存储
	return s.vectorStore.Insert(ctx, docs, vectors)
}

// Retrieve 检索相关文档
func (s *Service) Retrieve(ctx context.Context, query string, topK int) ([]Document, error) {
	// 1. 查询向量化
	queryVector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}

	// 2. 向量检索
	docs, err := s.vectorStore.Search(ctx, queryVector, topK*2) // 多检索一些用于重排序
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 3. 重排序 (如果有)
	if s.reranker != nil && len(docs) > 0 {
		docs, err = s.reranker.Rerank(ctx, query, docs, topK)
		if err != nil {
			// 重排序失败，使用原始结果
			if len(docs) > topK {
				docs = docs[:topK]
			}
		}
	} else if len(docs) > topK {
		docs = docs[:topK]
	}

	return docs, nil
}

// FormatContext 格式化检索结果为上下文
func (s *Service) FormatContext(docs []Document) string {
	var sb strings.Builder
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Content))
		if source, ok := doc.Metadata["source"]; ok {
			sb.WriteString(fmt.Sprintf("   来源: %s\n", source))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// splitIntoChunks 文本分块
func (s *Service) splitIntoChunks(text string) []string {
	if s.chunkSize <= 0 {
		return []string{text}
	}

	var chunks []string
	runes := []rune(text)
	length := len(runes)

	for i := 0; i < length; i += s.chunkSize - s.chunkOverlap {
		end := i + s.chunkSize
		if end > length {
			end = length
		}
		chunks = append(chunks, string(runes[i:end]))
		if end == length {
			break
		}
	}

	return chunks
}
