# 狼人杀 Agent RAG 增强推理设计文档

## 1. 概述

### 1.1 背景

当前 `wolf-go-adk` 项目中，Agent 的记忆管理存在以下局限：
- **无跨轮次记忆**：每轮对话独立，Agent 无法记住历史发言和行为模式
- **无语义检索**：无法根据语义相似度召回相关历史信息
- **决策依据单一**：Agent 仅依赖当前轮次信息做决策，缺乏历史推理能力

### 1.2 目标

使用 **Milvus + qwen-embedding-v4** 实现 RAG 增强推理：
1. 存储游戏事件（发言、投票、击杀等）到向量数据库
2. 在 Agent 决策时检索相关历史信息
3. 增强 Prompt 提升推理质量

### 1.3 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| 向量数据库 | **Milvus** | 生产级分布式向量数据库 |
| Embedding 模型 | **qwen-embedding-v4** | 阿里通义千问嵌入模型 |
| 框架 | **eino** | CloudWeGo 的 LLM 应用框架 |

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        ModeratorAgent                            │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    RAG Memory System                         ││
│  │                                                              ││
│  │  ┌─────────────────┐    ┌─────────────────────────────────┐ ││
│  │  │ Dashscope       │    │ Milvus                          │ ││
│  │  │ Embedder        │───▶│ - episodes collection           │ ││
│  │  │ (qwen-embed-v4) │    │ - 向量索引 + 元数据过滤          │ ││
│  │  └─────────────────┘    └─────────────────────────────────┘ ││
│  │           │                          │                       ││
│  │           └──────────┬───────────────┘                       ││
│  │                      ▼                                       ││
│  │           ┌─────────────────────┐                           ││
│  │           │   RAG Retriever     │                           ││
│  │           │   (eino retriever)  │                           ││
│  │           └─────────────────────┘                           ││
│  └─────────────────────────────────────────────────────────────┘│
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐     │
│  │ PlayerAgent │      │ PlayerAgent │      │ PlayerAgent │     │
│  └─────────────┘      └─────────────┘      └─────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. 依赖安装

```bash
# eino-ext milvus 组件
go get github.com/cloudwego/eino-ext/components/indexer/milvus
go get github.com/cloudwego/eino-ext/components/retriever/milvus

# eino-ext dashscope embedding (qwen-embedding-v4)
go get github.com/cloudwego/eino-ext/components/embedding/dashscope

# milvus SDK
go get github.com/milvus-io/milvus-sdk-go/v2
```

---

## 4. 数据模型

### 4.1 Episode（游戏事件）

```go
// memory/episode.go

package memory

import (
    "time"
    
    "github.com/cloudwego/eino/schema"
)

// EpisodeType 事件类型
type EpisodeType string

const (
    EpisodeSpeech     EpisodeType = "speech"     // 发言
    EpisodeVote       EpisodeType = "vote"       // 投票
    EpisodeKill       EpisodeType = "kill"       // 击杀
    EpisodeSave       EpisodeType = "save"       // 救人
    EpisodePoison     EpisodeType = "poison"     // 毒人
    EpisodeCheck      EpisodeType = "check"      // 查验
    EpisodeDeath      EpisodeType = "death"      // 死亡
    EpisodeAccusation EpisodeType = "accusation" // 指控
)

// Episode 游戏事件
type Episode struct {
    ID        string                 `json:"id"`
    GameID    string                 `json:"game_id"`
    Round     int                    `json:"round"`
    Phase     string                 `json:"phase"` // "night" | "day"
    Type      EpisodeType            `json:"type"`
    Actor     string                 `json:"actor"`   // 行动者
    Target    string                 `json:"target"`  // 目标（可选）
    Content   string                 `json:"content"` // 发言内容/行动描述
    Timestamp time.Time              `json:"timestamp"`
    Visible   []string               `json:"visible"` // 可见玩家列表，空表示公开
}

// ToDocument 转换为 eino Document
func (e *Episode) ToDocument() *schema.Document {
    return &schema.Document{
        ID:      e.ID,
        Content: e.Content,
        MetaData: map[string]any{
            "game_id":   e.GameID,
            "round":     e.Round,
            "phase":     e.Phase,
            "type":      string(e.Type),
            "actor":     e.Actor,
            "target":    e.Target,
            "timestamp": e.Timestamp.Unix(),
            "visible":   e.Visible,
        },
    }
}
```

### 4.2 Milvus Collection Schema

```go
// memory/milvus_schema.go

package memory

import (
    "github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
    CollectionName = "wolf_episodes"
    VectorDim      = 1024 // qwen-embedding-v4 维度
)

// GetCollectionSchema 获取 Milvus collection schema
func GetCollectionSchema() *entity.Schema {
    return &entity.Schema{
        CollectionName: CollectionName,
        Description:    "狼人杀游戏事件向量存储",
        Fields: []*entity.Field{
            {
                Name:       "id",
                DataType:   entity.FieldTypeVarChar,
                PrimaryKey: true,
                AutoID:     false,
                TypeParams: map[string]string{"max_length": "64"},
            },
            {
                Name:     "embedding",
                DataType: entity.FieldTypeFloatVector,
                TypeParams: map[string]string{
                    "dim": "1024", // qwen-embedding-v4
                },
            },
            {
                Name:       "game_id",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "32"},
            },
            {
                Name:     "round",
                DataType: entity.FieldTypeInt32,
            },
            {
                Name:       "phase",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "16"},
            },
            {
                Name:       "type",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "32"},
            },
            {
                Name:       "actor",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "32"},
            },
            {
                Name:       "target",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "32"},
            },
            {
                Name:       "content",
                DataType:   entity.FieldTypeVarChar,
                TypeParams: map[string]string{"max_length": "4096"},
            },
            {
                Name:     "timestamp",
                DataType: entity.FieldTypeInt64,
            },
        },
    }
}
```

---

## 5. RAG 实现（基于 eino）

### 5.1 初始化 Embedder 和 Milvus

```go
// memory/rag.go

package memory

import (
    "context"
    "fmt"
    "os"

    "github.com/cloudwego/eino-ext/components/embedding/dashscope"
    "github.com/cloudwego/eino-ext/components/indexer/milvus"
    milvusRetriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
    "github.com/cloudwego/eino/components/embedding"
    "github.com/cloudwego/eino/components/indexer"
    "github.com/cloudwego/eino/components/retriever"
    milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// RAGConfig RAG 配置
type RAGConfig struct {
    MilvusAddr     string // Milvus 地址，如 "localhost:19530"
    DashscopeKey   string // 阿里云 Dashscope API Key
    EmbeddingModel string // 嵌入模型，如 "text-embedding-v4"
}

// RAGSystem RAG 系统
type RAGSystem struct {
    embedder  embedding.Embedder
    indexer   indexer.Indexer
    retriever retriever.Retriever
    client    milvusClient.Client
}

// NewRAGSystem 创建 RAG 系统
func NewRAGSystem(ctx context.Context, config *RAGConfig) (*RAGSystem, error) {
    // 1. 创建 Milvus 客户端
    cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
        Address: config.MilvusAddr,
    })
    if err != nil {
        return nil, fmt.Errorf("创建 Milvus 客户端失败: %w", err)
    }

    // 2. 创建 Dashscope Embedder (qwen-embedding-v4)
    emb, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
        APIKey: config.DashscopeKey,
        Model:  config.EmbeddingModel, // "text-embedding-v4"
    })
    if err != nil {
        cli.Close()
        return nil, fmt.Errorf("创建 Embedder 失败: %w", err)
    }

    // 3. 创建 Milvus Indexer（用于存储）
    idx, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
        Client:     cli,
        Collection: CollectionName,
        Embedding:  emb,
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
        OutputFields: []string{
            "id", "game_id", "round", "phase", "type",
            "actor", "target", "content", "timestamp",
        },
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
    }, nil
}

// Close 关闭连接
func (r *RAGSystem) Close() error {
    return r.client.Close()
}
```

### 5.2 存储事件

```go
// memory/rag.go (续)

import (
    "github.com/cloudwego/eino/schema"
    "github.com/google/uuid"
)

// StoreEpisode 存储游戏事件
func (r *RAGSystem) StoreEpisode(ctx context.Context, episode *Episode) error {
    // 生成 ID
    if episode.ID == "" {
        episode.ID = uuid.New().String()
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
        docs[i] = ep.ToDocument()
    }

    _, err := r.indexer.Store(ctx, docs)
    return err
}
```

### 5.3 检索相关记忆

```go
// memory/rag.go (续)

import (
    "github.com/cloudwego/eino/components/retriever"
)

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
    // 设置检索选项
    opts := []retriever.Option{}
    if config.TopK > 0 {
        opts = append(opts, retriever.WithTopK(config.TopK))
    }

    // 执行检索
    docs, err := r.retriever.Retrieve(ctx, query, opts...)
    if err != nil {
        return nil, fmt.Errorf("检索失败: %w", err)
    }

    // 转换为 Episode
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
    return 0
}
```

### 5.4 构建增强 Prompt

```go
// memory/prompt_builder.go

package memory

import (
    "fmt"
    "strings"
)

// MemoryContext 记忆上下文
type MemoryContext struct {
    RelevantEpisodes []*Episode
    CurrentRound     int
    PlayerName       string
}

// BuildAugmentedPrompt 构建增强 Prompt
func BuildAugmentedPrompt(basePrompt string, memCtx *MemoryContext) string {
    if len(memCtx.RelevantEpisodes) == 0 {
        return basePrompt
    }

    var sb strings.Builder

    sb.WriteString("## 📚 相关历史记忆\n\n")

    // 按回合分组
    roundEpisodes := make(map[int][]*Episode)
    for _, ep := range memCtx.RelevantEpisodes {
        roundEpisodes[ep.Round] = append(roundEpisodes[ep.Round], ep)
    }

    // 按回合输出
    for round := 1; round <= memCtx.CurrentRound; round++ {
        eps, ok := roundEpisodes[round]
        if !ok {
            continue
        }
        sb.WriteString(fmt.Sprintf("### 第 %d 轮\n", round))
        for _, ep := range eps {
            sb.WriteString(formatEpisode(ep))
        }
        sb.WriteString("\n")
    }

    sb.WriteString("---\n\n")
    sb.WriteString(basePrompt)

    return sb.String()
}

// formatEpisode 格式化单个事件
func formatEpisode(ep *Episode) string {
    switch ep.Type {
    case EpisodeSpeech:
        return fmt.Sprintf("- 💬 [%s] 发言: \"%s\"\n", ep.Actor, truncate(ep.Content, 150))
    case EpisodeVote:
        return fmt.Sprintf("- 🗳️ [%s] 投票给 [%s]\n", ep.Actor, ep.Target)
    case EpisodeAccusation:
        return fmt.Sprintf("- ⚠️ [%s] 指控 [%s]: \"%s\"\n", ep.Actor, ep.Target, truncate(ep.Content, 100))
    case EpisodeDeath:
        return fmt.Sprintf("- 💀 [%s] 死亡\n", ep.Actor)
    case EpisodeCheck:
        return fmt.Sprintf("- 🔍 查验 [%s]: %s\n", ep.Target, ep.Content)
    default:
        return fmt.Sprintf("- [%s] %s\n", ep.Actor, ep.Content)
    }
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

---

## 6. 集成到 ModeratorAgent

### 6.1 初始化

```go
// agents/supervisor/moderator.go

import (
    "github.com/ashwinyue/wolf-go-adk/memory"
)

type ModeratorAgent struct {
    state        *game.GameState
    logger       *game.GameLogger
    playerAgents map[string]adk.Agent
    playerMsgs   map[string][]*schema.Message

    // 新增：RAG 系统
    rag *memory.RAGSystem

    mu sync.RWMutex
}

func NewModeratorAgent(ctx context.Context) (*ModeratorAgent, error) {
    // ... 现有初始化代码 ...

    // 初始化 RAG 系统
    ragConfig := &memory.RAGConfig{
        MilvusAddr:     os.Getenv("MILVUS_ADDR"),     // "localhost:19530"
        DashscopeKey:   os.Getenv("DASHSCOPE_API_KEY"),
        EmbeddingModel: "text-embedding-v4",
    }
    
    rag, err := memory.NewRAGSystem(ctx, ragConfig)
    if err != nil {
        return nil, fmt.Errorf("初始化 RAG 系统失败: %w", err)
    }

    return &ModeratorAgent{
        state:        state,
        logger:       logger,
        playerAgents: playerAgents,
        playerMsgs:   playerMsgs,
        rag:          rag,
    }, nil
}

func (m *ModeratorAgent) Close() error {
    if m.rag != nil {
        return m.rag.Close()
    }
    return nil
}
```

### 6.2 讨论阶段集成

```go
// agents/supervisor/day_phase.go

func (m *ModeratorAgent) discussPhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], alivePlayers []string) {
    m.sendMessage(gen, "  💬 讨论阶段:")
    m.logger.LogPhase("💬 讨论阶段")

    for _, player := range alivePlayers {
        baseQuery := "轮到你发言了，请分析局势并表达你的观点。"

        // RAG 检索相关记忆
        augmentedQuery := baseQuery
        if m.rag != nil {
            episodes, err := m.rag.RetrieveRelevant(ctx, "分析局势 怀疑 投票", &memory.RetrieveConfig{
                TopK:     5,
                GameID:   m.logger.GetGameID(),
                MaxRound: m.state.Round,
            })
            if err == nil && len(episodes) > 0 {
                memCtx := &memory.MemoryContext{
                    RelevantEpisodes: episodes,
                    CurrentRound:     m.state.Round,
                    PlayerName:       player,
                }
                augmentedQuery = memory.BuildAugmentedPrompt(baseQuery, memCtx)
            }
        }

        // 调用玩家 Agent
        response := m.callPlayer(ctx, player, augmentedQuery)

        if response != "" {
            // 存储发言到 RAG
            if m.rag != nil {
                m.rag.StoreEpisode(ctx, &memory.Episode{
                    GameID:  m.logger.GetGameID(),
                    Round:   m.state.Round,
                    Phase:   "day",
                    Type:    memory.EpisodeSpeech,
                    Actor:   player,
                    Content: response,
                })
            }

            m.sendMessage(gen, fmt.Sprintf("  [%s]: %s", player, utils.Truncate(response, 200)))
            m.broadcastToAll(fmt.Sprintf("[%s]: %s", player, response))
            m.logger.LogDiscussion(player, response)
        }
    }
}
```

### 6.3 投票阶段集成

```go
// agents/supervisor/day_phase.go

func (m *ModeratorAgent) votePhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], alivePlayers []string) {
    m.sendMessage(gen, "  🗳️ 投票阶段:")

    for _, player := range alivePlayers {
        baseQuery := fmt.Sprintf(params.Prompts.ToAllVote, strings.Join(alivePlayers, ", "))

        // RAG 检索投票相关记忆
        augmentedQuery := baseQuery
        if m.rag != nil {
            episodes, err := m.rag.RetrieveRelevant(ctx, "投票 怀疑 狼人 可疑", &memory.RetrieveConfig{
                TopK:     3,
                GameID:   m.logger.GetGameID(),
                MaxRound: m.state.Round,
            })
            if err == nil && len(episodes) > 0 {
                memCtx := &memory.MemoryContext{
                    RelevantEpisodes: episodes,
                    CurrentRound:     m.state.Round,
                    PlayerName:       player,
                }
                augmentedQuery = memory.BuildAugmentedPrompt(baseQuery, memCtx)
            }
        }

        // 调用投票
        target := m.callPlayerVote(ctx, player, augmentedQuery)

        if target != "" {
            // 存储投票到 RAG
            if m.rag != nil {
                m.rag.StoreEpisode(ctx, &memory.Episode{
                    GameID:  m.logger.GetGameID(),
                    Round:   m.state.Round,
                    Phase:   "day",
                    Type:    memory.EpisodeVote,
                    Actor:   player,
                    Target:  target,
                    Content: fmt.Sprintf("%s 投票给 %s", player, target),
                })
            }
        }
    }
}
```

---

## 7. 环境配置

### 7.1 环境变量

```bash
# Milvus 配置
export MILVUS_ADDR="localhost:19530"

# 阿里云 Dashscope 配置 (qwen-embedding-v4)
export DASHSCOPE_API_KEY="sk-xxx"
```

### 7.2 Docker 启动 Milvus

```bash
# 使用 docker-compose 启动 Milvus standalone
wget https://github.com/milvus-io/milvus/releases/download/v2.3.0/milvus-standalone-docker-compose.yml -O docker-compose.yml

docker-compose up -d
```

### 7.3 初始化 Collection

```go
// 首次运行时创建 collection
func (r *RAGSystem) InitCollection(ctx context.Context) error {
    // 检查 collection 是否存在
    has, err := r.client.HasCollection(ctx, CollectionName)
    if err != nil {
        return err
    }
    if has {
        return nil
    }

    // 创建 collection
    schema := GetCollectionSchema()
    err = r.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
    if err != nil {
        return err
    }

    // 创建索引
    idx, _ := entity.NewIndexIvfFlat(entity.L2, 128)
    err = r.client.CreateIndex(ctx, CollectionName, "embedding", idx, false)
    if err != nil {
        return err
    }

    // 加载 collection
    return r.client.LoadCollection(ctx, CollectionName, false)
}
```

---

## 8. 目录结构

```
wolf-go-adk/
├── memory/
│   ├── episode.go          # Episode 数据模型
│   ├── milvus_schema.go    # Milvus schema 定义
│   ├── rag.go              # RAG 系统（Milvus + Dashscope）
│   └── prompt_builder.go   # Prompt 增强构建器
├── agents/
│   └── supervisor/
│       ├── moderator.go    # 集成 RAG 系统
│       └── day_phase.go    # 讨论/投票阶段使用 RAG
└── docs/
    └── MEMORY_RAG_DESIGN.md
```

---

## 9. 实现路线图

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| **Phase 1** | 基础 RAG 实现 | 3 天 |
| | - 创建 `memory/` 包 | |
| | - 实现 Episode 数据模型 | |
| | - 集成 eino milvus indexer/retriever | |
| | - 集成 dashscope embedder | |
| **Phase 2** | Agent 集成 | 2 天 |
| | - ModeratorAgent 初始化 RAG | |
| | - 讨论阶段 RAG 增强 | |
| | - 投票阶段 RAG 增强 | |
| **Phase 3** | 测试优化 | 2 天 |
| | - 单元测试 | |
| | - 效果评估 | |
| | - 性能优化 | |

---

## 10. 成本估算

### qwen-embedding-v4 定价

| 模型 | 价格 |
|------|------|
| text-embedding-v4 | ¥0.0007 / 1000 tokens |

### 单局游戏估算（9人局，5轮）

| 项目 | 数量 | Token 估算 | 成本 |
|------|------|-----------|------|
| 发言存储 | 45 条 | ~9000 | ¥0.0063 |
| 投票存储 | 45 条 | ~2250 | ¥0.0016 |
| 检索查询 | 90 次 | ~900 | ¥0.0006 |
| **总计** | | ~12150 | **¥0.0085** |

Embedding 成本极低，主要成本在 LLM 调用。
