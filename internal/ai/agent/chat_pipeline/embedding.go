package chat_pipeline

import (
	"context"
	"github.com/NuyoahCh/eocall/internal/ai/embedder"
	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	return embedder.DoubaoEmbedding(ctx)
}
