# 讨论阶段多轮对话设计文档

## 1. 背景

当前狼人杀游戏的讨论阶段采用**单轮发言**模式：每个玩家按顺序发言一次，然后直接进入投票阶段。这种模式存在以下问题：

### 1.1 当前模式的局限性

| 问题 | 描述 |
|------|------|
| **信息不对称** | 先发言的玩家无法回应后发言玩家的质疑 |
| **缺乏互动** | 玩家之间没有真正的"讨论"，只是单向陈述 |
| **策略受限** | 无法通过追问、反驳来揭露狼人破绽 |
| **游戏体验** | 与真实狼人杀游戏的讨论氛围差距较大 |

### 1.2 真实狼人杀的讨论模式

真实游戏中，讨论阶段通常包含：
- **首轮发言**：每人陈述观点
- **自由讨论**：玩家可以互相质疑、辩论
- **总结发言**：投票前的最后陈述

## 2. 多轮对话方案设计

### 2.1 方案概述

引入**多轮对话机制**，让讨论阶段更加真实和完整：

```
┌─────────────────────────────────────────────────────────────┐
│                     讨论阶段流程                              │
├─────────────────────────────────────────────────────────────┤
│  第一轮：顺序发言（每人一次）                                  │
│     ↓                                                        │
│  第二轮：自由讨论（可选择发言或跳过）                          │
│     ↓                                                        │
│  第三轮：总结陈词（每人一次，可选）                            │
│     ↓                                                        │
│  投票阶段                                                     │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 详细设计

#### 2.2.1 讨论轮次配置

```go
type DiscussionConfig struct {
    // 首轮发言：必须，每人一次
    FirstRoundRequired bool
    
    // 自由讨论轮数：0-3 轮
    FreeDiscussionRounds int
    
    // 总结陈词：可选
    SummaryRoundEnabled bool
    
    // 每轮最大发言人数（自由讨论）
    MaxSpeakersPerRound int
    
    // 发言超时时间
    SpeakTimeout time.Duration
}

// 默认配置
var DefaultDiscussionConfig = DiscussionConfig{
    FirstRoundRequired:    true,
    FreeDiscussionRounds:  2,
    SummaryRoundEnabled:   false,
    MaxSpeakersPerRound:   5,
    SpeakTimeout:          30 * time.Second,
}
```

#### 2.2.2 自由讨论机制

在自由讨论阶段，需要解决**谁来发言**的问题：

**方案 A：主持人决定（推荐）**
```go
// 主持人根据上下文决定下一个发言者
func (m *ModeratorAgent) decideNextSpeaker(ctx context.Context, 
    alivePlayers []string, 
    lastSpeaker string,
    discussionHistory []Speech) string {
    
    prompt := fmt.Sprintf(`
当前讨论历史：
%s

请决定下一个发言者，考虑：
1. 被质疑的玩家应该有机会回应
2. 沉默的玩家可能需要被点名
3. 讨论是否已经充分

输出 JSON：{"next_speaker": "PlayerX", "reason": "..."}
或 {"end_discussion": true, "reason": "讨论已充分"}
`, formatHistory(discussionHistory))
    
    return m.callModeratorLLM(ctx, prompt)
}
```

**方案 B：玩家自主申请**
```go
// 每轮询问所有玩家是否想发言
func (m *ModeratorAgent) collectSpeakRequests(ctx context.Context, 
    alivePlayers []string) []string {
    
    var wantToSpeak []string
    for _, player := range alivePlayers {
        response := m.callPlayer(ctx, player, 
            "是否想要发言？回复 '是' 或 '否'")
        if strings.Contains(response, "是") {
            wantToSpeak = append(wantToSpeak, player)
        }
    }
    return wantToSpeak
}
```

**方案 C：混合模式（最佳）**
```go
// 结合主持人引导和玩家意愿
func (m *ModeratorAgent) freeDiscussionRound(ctx context.Context,
    gen *adk.AsyncGenerator[*adk.AgentEvent],
    alivePlayers []string,
    history []Speech) {
    
    // 1. 主持人分析谁应该发言
    suggestions := m.analyzeSpeakingNeeds(ctx, history)
    
    // 2. 询问被建议的玩家
    for _, player := range suggestions {
        prompt := fmt.Sprintf(
            "有玩家质疑了你的观点，你是否想要回应？")
        response := m.callPlayer(ctx, player, prompt)
        if response != "" && !isSkip(response) {
            m.broadcastSpeech(gen, player, response)
            history = append(history, Speech{Player: player, Content: response})
        }
    }
    
    // 3. 询问其他玩家是否有补充
    for _, player := range alivePlayers {
        if !contains(suggestions, player) {
            prompt := "是否有补充发言？回复内容或'跳过'"
            response := m.callPlayer(ctx, player, prompt)
            if response != "" && !isSkip(response) {
                m.broadcastSpeech(gen, player, response)
            }
        }
    }
}
```

#### 2.2.3 上下文管理

多轮对话需要更好的上下文管理：

```go
type DiscussionContext struct {
    Round       int                    // 当前讨论轮次
    Speeches    []Speech               // 所有发言记录
    Accusations map[string][]string    // 质疑关系：被质疑者 -> 质疑者列表
    KeyPoints   []string               // 关键信息点
}

type Speech struct {
    Player    string
    Content   string
    Round     int
    Timestamp time.Time
    ReplyTo   string  // 回复谁的发言
}

// 构建玩家的讨论上下文
func (m *ModeratorAgent) buildDiscussionPrompt(
    player string, 
    ctx *DiscussionContext) string {
    
    var prompt strings.Builder
    
    // 1. 之前的发言摘要
    prompt.WriteString("=== 讨论记录 ===\n")
    for _, speech := range ctx.Speeches {
        prompt.WriteString(fmt.Sprintf("[%s]: %s\n", 
            speech.Player, speech.Content))
    }
    
    // 2. 针对该玩家的质疑
    if accusations, ok := ctx.Accusations[player]; ok {
        prompt.WriteString(fmt.Sprintf(
            "\n⚠️ 以下玩家质疑了你: %s\n", 
            strings.Join(accusations, ", ")))
    }
    
    // 3. 提示
    prompt.WriteString("\n请发表你的观点，可以：\n")
    prompt.WriteString("- 回应对你的质疑\n")
    prompt.WriteString("- 质疑其他玩家\n")
    prompt.WriteString("- 分析局势\n")
    prompt.WriteString("- 回复'跳过'放弃发言\n")
    
    return prompt.String()
}
```

#### 2.2.4 质疑检测

自动检测发言中的质疑关系：

```go
// 分析发言中的质疑
func (m *ModeratorAgent) detectAccusations(
    speaker string, 
    content string, 
    alivePlayers []string) []string {
    
    var accused []string
    
    // 方法1：关键词匹配
    accusationKeywords := []string{
        "怀疑", "可疑", "狼人", "有问题", 
        "不对劲", "撒谎", "投票淘汰",
    }
    
    for _, player := range alivePlayers {
        if player == speaker {
            continue
        }
        // 检查是否提到该玩家 + 质疑关键词
        if strings.Contains(content, player) {
            for _, kw := range accusationKeywords {
                if strings.Contains(content, kw) {
                    accused = append(accused, player)
                    break
                }
            }
        }
    }
    
    // 方法2：LLM 分析（更准确）
    if len(accused) == 0 {
        accused = m.llmDetectAccusations(content, alivePlayers)
    }
    
    return accused
}
```

### 2.3 人类玩家的多轮对话

对于人类玩家，需要特殊处理：

```go
func (m *ModeratorAgent) humanFreeDiscussion(
    ctx context.Context,
    gen *adk.AsyncGenerator[*adk.AgentEvent],
    humanPlayer string,
    discussionCtx *DiscussionContext) {
    
    // 显示当前讨论状态
    prompt := fmt.Sprintf(`
=== 当前讨论 ===
%s

你可以：
1. 输入发言内容
2. 输入 '@PlayerX 内容' 回复特定玩家
3. 输入 '跳过' 放弃发言
4. 输入 '投票' 结束讨论进入投票
`, formatDiscussion(discussionCtx))
    
    response := m.callPlayer(ctx, humanPlayer, prompt)
    
    // 解析人类输入
    if strings.HasPrefix(response, "@") {
        // 回复特定玩家
        parts := strings.SplitN(response[1:], " ", 2)
        if len(parts) == 2 {
            replyTo := parts[0]
            content := parts[1]
            m.broadcastReply(gen, humanPlayer, replyTo, content)
        }
    } else if response == "投票" {
        // 人类请求结束讨论
        m.endDiscussion = true
    } else if response != "跳过" {
        m.broadcastSpeech(gen, humanPlayer, response)
    }
}
```

## 3. 实现步骤

### 3.1 第一阶段：基础多轮

1. 修改 `discussPhase` 支持多轮发言
2. 添加 `DiscussionConfig` 配置
3. 实现基本的轮次控制

```go
// agents/supervisor/day_phase.go

func (m *ModeratorAgent) discussPhase(ctx context.Context, 
    gen *adk.AsyncGenerator[*adk.AgentEvent], 
    alivePlayers []string) {
    
    config := DefaultDiscussionConfig
    discussionCtx := &DiscussionContext{}
    
    // 第一轮：顺序发言
    m.sendMessage(gen, "  💬 第一轮发言:")
    speakingOrder := m.decideSpeakingOrder(ctx, alivePlayers)
    for _, player := range speakingOrder {
        response := m.callPlayer(ctx, player, 
            m.buildDiscussionPrompt(player, discussionCtx))
        if response != "" {
            discussionCtx.Speeches = append(discussionCtx.Speeches, 
                Speech{Player: player, Content: response, Round: 1})
            m.broadcastSpeech(gen, player, response)
            
            // 检测质疑
            accused := m.detectAccusations(player, response, alivePlayers)
            for _, a := range accused {
                discussionCtx.Accusations[a] = append(
                    discussionCtx.Accusations[a], player)
            }
        }
    }
    
    // 自由讨论轮次
    for round := 0; round < config.FreeDiscussionRounds; round++ {
        m.sendMessage(gen, fmt.Sprintf("  💬 自由讨论 (第 %d 轮):", round+1))
        
        if !m.freeDiscussionRound(ctx, gen, alivePlayers, discussionCtx) {
            break // 讨论已充分，提前结束
        }
    }
}
```

### 3.2 第二阶段：智能引导

1. 实现主持人智能决定发言者
2. 添加质疑检测
3. 优化上下文提示

### 3.3 第三阶段：人类交互优化

1. 改进人类玩家的交互界面
2. 支持 `@` 回复语法
3. 支持人类主动结束讨论

## 4. 配置选项

```go
// params/game.go

type GameConfig struct {
    // ... 其他配置
    
    // 讨论配置
    Discussion DiscussionConfig
}

// 环境变量支持
// DISCUSSION_FREE_ROUNDS=2
// DISCUSSION_SUMMARY=true
```

## 5. 预期效果

### 5.1 游戏体验提升

| 方面 | 单轮模式 | 多轮模式 |
|------|----------|----------|
| 互动性 | ⭐ | ⭐⭐⭐⭐ |
| 策略深度 | ⭐⭐ | ⭐⭐⭐⭐ |
| 真实感 | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| 游戏时长 | 短 | 中等 |
| Token 消耗 | 低 | 中高 |

### 5.2 示例对话

```
=== 第一轮发言 ===
[Player1]: 昨晚平安夜，我认为我们应该关注发言内容...
[Player2]: 我同意 Player1 的观点，但我注意到 Player5 一直很沉默...
[Player3]: Player2 你为什么要针对 Player5？你是不是想转移注意力？

=== 自由讨论 (第 1 轮) ===
[Player2]: 我只是提出疑问，Player3 你反应这么大是不是心虚？
[Player5]: 我来回应一下，我之前沉默是因为在观察...
[Player3]: 我没有心虚，我只是觉得 Player2 的逻辑有问题

=== 自由讨论 (第 2 轮) ===
[主持人]: 讨论已充分，进入投票阶段
```

## 6. 风险与对策

| 风险 | 对策 |
|------|------|
| 讨论过长 | 设置最大轮次和超时 |
| Token 消耗高 | 压缩历史上下文 |
| 死循环争论 | 主持人强制结束 |
| 人类等待过久 | 并行处理 AI 发言 |

## 7. 总结

引入多轮对话机制可以显著提升狼人杀游戏的真实感和策略深度。建议采用**混合模式**（主持人引导 + 玩家意愿），分阶段实现，优先保证游戏流畅性。

### 推荐配置

- **快速模式**：1 轮首发 + 1 轮自由讨论
- **标准模式**：1 轮首发 + 2 轮自由讨论
- **深度模式**：1 轮首发 + 3 轮自由讨论 + 总结陈词
