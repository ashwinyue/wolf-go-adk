/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	milvusRetriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// RAGConfig RAG 配置
type RAGConfig struct {
	MilvusAddr string // Milvus 地址，如 "localhost:19530"
	ArkAPIKey  string // 火山引擎 Ark API Key
	ArkModel   string // Ark 端点 ID，如 "ep-xxxxxxx-xxxxx"
}

// RAGSystem RAG 系统
type RAGSystem struct {
	embedder  embedding.Embedder
	indexer   indexer.Indexer
	retriever retriever.Retriever
	client    milvusClient.Client
	gameID    string
}

// NewRAGSystem 创建 RAG 系统
func NewRAGSystem(ctx context.Context, config *RAGConfig) (*RAGSystem, error) {
	// 验证配置
	if config.MilvusAddr == "" {
		return nil, fmt.Errorf("MILVUS_ADDR 未配置")
	}
	if config.ArkAPIKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY 未配置")
	}
	if config.ArkModel == "" {
		return nil, fmt.Errorf("ARK_MODEL 未配置")
	}

	// 1. 创建 Milvus 客户端
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address: config.MilvusAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Milvus 客户端失败: %w", err)
	}

	// 2. 创建 Ark Embedder (火山引擎)
	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: config.ArkAPIKey,
		Model:  config.ArkModel,
	})
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("创建 Embedder 失败: %w", err)
	}

	// 3. 创建 Milvus Indexer（用于存储，会自动创建 Collection）
	idx, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:              cli,
		Collection:          CollectionName,
		Embedding:           emb,
		Fields:              GetEinoFields(), // 使用 FloatVector 的 schema
		EnableDynamicSchema: true,
		MetricType:          milvus.COSINE,        // qwen-embedding-v4 使用 COSINE
		DocumentConverter:   floatVectorConverter, // 自定义转换器
	})
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("创建 Indexer 失败: %w", err)
	}

	// 4. 创建 Milvus Retriever（用于检索）
	ret, err := milvusRetriever.NewRetriever(ctx, &milvusRetriever.RetrieverConfig{
		Client:     cli,
		Collection: CollectionName,
		Embedding:  emb,
		TopK:       10,
	})
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("创建 Retriever 失败: %w", err)
	}

	return &RAGSystem{
		embedder:  emb,
		indexer:   idx,
		retriever: ret,
		client:    cli,
		gameID:    uuid.New().String(),
	}, nil
}

// floatVectorRow 用于 Milvus 存储的行结构
type floatVectorRow struct {
	ID       string    `json:"id" milvus:"name:id"`
	Vector   []float32 `json:"vector" milvus:"name:vector"`
	Content  string    `json:"content" milvus:"name:content"`
	Metadata []byte    `json:"metadata" milvus:"name:metadata"`
}

// floatVectorConverter 自定义文档转换器（使用 FloatVector）
func floatVectorConverter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	em := make([]floatVectorRow, 0, len(docs))
	rows := make([]interface{}, 0, len(docs))

	for i, doc := range docs {
		// 将 float64 转换为 float32
		vector32 := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			vector32[j] = float32(v)
		}

		// 构建元数据 JSON
		metadataBytes, _ := json.Marshal(doc.MetaData)

		em = append(em, floatVectorRow{
			ID:       doc.ID,
			Vector:   vector32,
			Content:  doc.Content,
			Metadata: metadataBytes,
		})
	}

	// 返回结构体指针
	for idx := range em {
		rows = append(rows, &em[idx])
	}
	return rows, nil
}

// Close 关闭连接
func (r *RAGSystem) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// GetGameID 获取当前游戏 ID
func (r *RAGSystem) GetGameID() string {
	return r.gameID
}

// SetGameID 设置游戏 ID
func (r *RAGSystem) SetGameID(gameID string) {
	r.gameID = gameID
}

// StoreEpisode 存储游戏事件
func (r *RAGSystem) StoreEpisode(ctx context.Context, episode *Episode) error {
	// 生成 ID
	if episode.ID == "" {
		episode.ID = uuid.New().String()
	}
	// 设置游戏 ID
	if episode.GameID == "" {
		episode.GameID = r.gameID
	}
	// 设置时间戳
	if episode.Timestamp.IsZero() {
		episode.Timestamp = time.Now()
	}

	// 转换为 Document
	doc := episode.ToDocument()

	// 存储到 Milvus（eino indexer 会自动处理 embedding）
	_, err := r.indexer.Store(ctx, []*schema.Document{doc})
	return err
}

// StoreEpisodes 批量存储事件
func (r *RAGSystem) StoreEpisodes(ctx context.Context, episodes []*Episode) error {
	docs := make([]*schema.Document, len(episodes))
	for i, ep := range episodes {
		if ep.ID == "" {
			ep.ID = uuid.New().String()
		}
		if ep.GameID == "" {
			ep.GameID = r.gameID
		}
		if ep.Timestamp.IsZero() {
			ep.Timestamp = time.Now()
		}
		docs[i] = ep.ToDocument()
	}

	_, err := r.indexer.Store(ctx, docs)
	return err
}

// RetrieveConfig 检索配置
type RetrieveConfig struct {
	TopK     int    // 返回数量
	GameID   string // 游戏 ID 过滤
	MaxRound int    // 最大回合过滤
	Phase    string // 阶段过滤
	Actor    string // 行动者过滤
}

// RetrieveRelevant 检索相关记忆
func (r *RAGSystem) RetrieveRelevant(ctx context.Context, query string, config *RetrieveConfig) ([]*Episode, error) {
	if config == nil {
		config = &RetrieveConfig{TopK: 10}
	}
	if config.TopK <= 0 {
		config.TopK = 10
	}

	// 设置检索选项
	opts := []retriever.Option{
		retriever.WithTopK(config.TopK),
	}

	// 执行检索
	docs, err := r.retriever.Retrieve(ctx, query, opts...)
	if err != nil {
		return nil, fmt.Errorf("检索失败: %w", err)
	}

	// 转换为 Episode 并应用过滤
	episodes := make([]*Episode, 0, len(docs))
	for _, doc := range docs {
		ep := documentToEpisode(doc)

		// 应用过滤条件
		if config.GameID != "" && ep.GameID != config.GameID {
			continue
		}
		if config.MaxRound > 0 && ep.Round > config.MaxRound {
			continue
		}
		if config.Phase != "" && ep.Phase != config.Phase {
			continue
		}
		if config.Actor != "" && ep.Actor != config.Actor {
			continue
		}

		episodes = append(episodes, ep)
	}

	return episodes, nil
}

// documentToEpisode 将 Document 转换为 Episode
func documentToEpisode(doc *schema.Document) *Episode {
	meta := doc.MetaData
	return &Episode{
		ID:      doc.ID,
		GameID:  getStringMeta(meta, "game_id"),
		Round:   getIntMeta(meta, "round"),
		Phase:   getStringMeta(meta, "phase"),
		Type:    EpisodeType(getStringMeta(meta, "type")),
		Actor:   getStringMeta(meta, "actor"),
		Target:  getStringMeta(meta, "target"),
		Content: doc.Content,
	}
}

func getStringMeta(meta map[string]any, key string) string {
	if v, ok := meta[key].(string); ok {
		return v
	}
	return ""
}

func getIntMeta(meta map[string]any, key string) int {
	if v, ok := meta[key].(int); ok {
		return v
	}
	if v, ok := meta[key].(int32); ok {
		return int(v)
	}
	if v, ok := meta[key].(int64); ok {
		return int(v)
	}
	if v, ok := meta[key].(float64); ok {
		return int(v)
	}
	return 0
}
