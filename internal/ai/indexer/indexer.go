package indexer

import (
	"context"
	embedder2 "github.com/NuyoahCh/eocall/internal/ai/embedder"
	"github.com/NuyoahCh/eocall/utility/client"
	"github.com/NuyoahCh/eocall/utility/common"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// NewMilvusIndexer 初始化数据库索引组件
func NewMilvusIndexer(ctx context.Context) (*milvus.Indexer, error) {
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return nil, err
	}
	eb, err := embedder2.DoubaoEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	config := &milvus.IndexerConfig{
		Client:     cli,
		Collection: common.MilvusCollectionName,
		Fields:     fields,
		Embedding:  eb,
	}
	indexer, err := milvus.NewIndexer(ctx, config)
	if err != nil {
		return nil, err
	}
	return indexer, nil
}

var fields = []*entity.Field{
	{
		Name:     "id",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "255",
		},
		PrimaryKey: true,
	},
	{
		Name:     "vector", // 确保字段名匹配
		DataType: entity.FieldTypeBinaryVector,
		TypeParams: map[string]string{
			"dim": "65536",
		},
	},
	{
		Name:     "content",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "8192",
		},
	},
	{
		Name:     "metadata",
		DataType: entity.FieldTypeJSON,
	},
}
