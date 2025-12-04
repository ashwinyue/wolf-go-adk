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

package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/wolf-go-adk/agents/players"
	"github.com/ashwinyue/wolf-go-adk/game"
	"github.com/ashwinyue/wolf-go-adk/memory"
	"github.com/ashwinyue/wolf-go-adk/params"
	"github.com/ashwinyue/wolf-go-adk/tools"
	"github.com/ashwinyue/wolf-go-adk/utils"
)

// dayPhase 白天阶段
func (m *ModeratorAgent) dayPhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	m.sendMessage(gen, "\n--- ☀️ 白天阶段 ---")
	m.state.Phase = "day"
	m.logger.LogPhase("☀️ 白天阶段")

	// 公布夜间死亡
	var dead []string
	if m.state.NightKilled != "" && !m.state.NightSaved {
		dead = append(dead, m.state.NightKilled)
	}
	if m.state.NightPoisoned != "" {
		dead = append(dead, m.state.NightPoisoned)
	}
	if m.state.NightShot != "" {
		dead = append(dead, m.state.NightShot)
	}

	if len(dead) > 0 {
		announcement := fmt.Sprintf(params.Prompts.ToAllDay, strings.Join(dead, ", "))
		m.broadcastToAll(announcement) // 广播给所有玩家
		m.sendMessage(gen, fmt.Sprintf("  📢 %s", announcement))
		m.logger.LogModerator(fmt.Sprintf("昨晚 %s 被淘汰了。", strings.Join(dead, ", ")))

		// 猎人开枪消息
		if m.state.NightShot != "" {
			hunterMsg := fmt.Sprintf(params.Prompts.ToAllHunterShoot, m.state.NightShot)
			m.broadcastToAll(hunterMsg)
			m.sendMessage(gen, fmt.Sprintf("  📢 %s", hunterMsg))
		}

		// 第一晚死者遗言
		if m.state.FirstDay && m.state.NightKilled != "" && !m.state.NightSaved {
			m.lastWords(ctx, gen, m.state.NightKilled)
		}
	} else {
		m.broadcastToAll(params.Prompts.ToAllPeace)
		m.sendMessage(gen, fmt.Sprintf("  📢 %s", params.Prompts.ToAllPeace))
		m.logger.LogModerator("昨晚是平安夜，没有人被淘汰。")
	}

	// 检查胜利条件
	if winner := m.state.CheckWinner(); winner != "" {
		return
	}

	alivePlayers := m.state.GetAlivePlayers()
	m.sendMessage(gen, fmt.Sprintf("  📢 存活玩家: %s", strings.Join(alivePlayers, ", ")))

	// 1. 讨论阶段
	m.discussPhase(ctx, gen, alivePlayers)

	// 2. 投票阶段
	m.votePhase(ctx, gen, alivePlayers)
}

// discussPhase 讨论阶段
func (m *ModeratorAgent) discussPhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], alivePlayers []string) {
	m.sendMessage(gen, "  💬 讨论阶段:")
	m.logger.LogPhase("💬 讨论阶段")
	m.logger.LogModerator("现在进入讨论阶段，请各位玩家依次发言。")

	// 主持人决定发言顺序
	speakingOrder := m.decideSpeakingOrder(ctx, alivePlayers)
	m.sendMessage(gen, fmt.Sprintf("  📢 发言顺序: %s", strings.Join(speakingOrder, " → ")))
	m.logger.LogModerator(fmt.Sprintf("发言顺序: %s", strings.Join(speakingOrder, " → ")))
	// 广播讨论开始
	discussMsg := fmt.Sprintf(params.Prompts.ToAllDiscuss, strings.Join(speakingOrder, ", "), strings.Join(speakingOrder, ", "))
	m.broadcastToAll(discussMsg)

	for _, player := range speakingOrder {
		baseQuery := "轮到你发言了，请分析局势并表达你的观点。"

		// RAG 检索相关记忆
		augmentedQuery := m.augmentQueryWithRAG(ctx, baseQuery, player, "day")

		response := m.callPlayer(ctx, player, augmentedQuery)
		if response != "" {
			// 存储发言到 RAG
			m.storeEpisodeToRAG(ctx, memory.EpisodeSpeech, player, "", response)

			// 检测并存储怀疑关系
			accused := memory.DetectAccusations(player, response, speakingOrder)
			for _, target := range accused {
				m.storeEpisodeToRAG(ctx, memory.EpisodeAccusation, player, target,
					fmt.Sprintf("%s 怀疑 %s", player, target))
			}

			m.sendMessage(gen, fmt.Sprintf("  [%s]: %s", player, utils.Truncate(response, 500)))
			// 广播给所有人
			m.broadcastToAll(fmt.Sprintf("[%s]: %s", player, response))
			m.logger.LogDiscussion(player, response)
		}
	}
}

// votePhase 投票阶段
func (m *ModeratorAgent) votePhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], alivePlayers []string) {
	m.sendMessage(gen, "  🗳️ 投票阶段:")
	m.logger.LogPhase("🗳️ 投票阶段")
	m.logger.LogModerator("讨论结束，现在进入投票阶段，请投票选出你认为的狼人。")

	votes := make(map[string]string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 创建投票工具
	voteTool := tools.NewVoteTool(m.state)

	for _, player := range alivePlayers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			baseQuery := fmt.Sprintf(params.Prompts.ToAllVote, strings.Join(alivePlayers, ", "))

			// RAG 增强投票查询
			augmentedQuery := m.augmentQueryWithRAG(ctx, baseQuery, p, "day")

			var target string
			if voteTool != nil {
				result, err := m.callPlayerWithTool(ctx, p, augmentedQuery, voteTool)
				if err == nil {
					if t, ok := result["target"].(string); ok {
						target = t
					}
				}
			}

			if target != "" && target != p {
				mu.Lock()
				votes[p] = target
				m.logger.LogVote(p, target)

				// 存储投票到 RAG
				m.storeEpisodeToRAG(ctx, memory.EpisodeVote, p, target, fmt.Sprintf("%s 投票给 %s", p, target))

				mu.Unlock()
				m.sendMessage(gen, fmt.Sprintf("  [%s] 投票: %s", p, target))
			}
		}(player)
	}
	wg.Wait()

	if len(votes) == 0 {
		m.sendMessage(gen, "  ➡️ 无有效投票")
		m.logger.LogVoteResult("", "无有效投票")
		return
	}

	votedOut, details := utils.MajorityVote(votes)
	// 广播投票结果
	voteResultMsg := fmt.Sprintf(params.Prompts.ToAllRes, details, votedOut)
	m.broadcastToAll(voteResultMsg)
	m.sendMessage(gen, fmt.Sprintf("  ➡️ 投票结果: %s 被淘汰 (%s)", votedOut, details))
	m.logger.LogVoteResult(votedOut, details)

	if votedOut != "" {
		role := m.state.GetPlayerRole(votedOut)

		// 存储死亡事件到 RAG
		m.storeEpisodeToRAG(ctx, memory.EpisodeDeath, votedOut, "", fmt.Sprintf("%s 被投票淘汰，身份是%s", votedOut, role))

		// 遗言
		m.lastWords(ctx, gen, votedOut)

		m.state.KillPlayer(votedOut)

		// 猎人开枪
		if role == game.RoleHunter && m.state.NightPoisoned != votedOut {
			m.hunterShoot(ctx, gen, votedOut)
		}
	}
}

// lastWords 遗言
func (m *ModeratorAgent) lastWords(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], player string) {
	query := fmt.Sprintf(params.Prompts.ToDeadPlayer, player)
	// 广播遗言提示
	m.broadcastToAll(query)

	response := m.callPlayer(ctx, player, query)
	if response != "" {
		m.sendMessage(gen, fmt.Sprintf("  [%s] (遗言): %s", player, utils.Truncate(response, 200)))
		// 遗言广播给所有人
		m.broadcastToAll(fmt.Sprintf("[%s 遗言]: %s", player, response))
		m.logger.LogLastWords(player, response)

		// 存储遗言到 RAG
		m.storeEpisodeToRAG(ctx, memory.EpisodeLastWords, player, "",
			fmt.Sprintf("%s 的遗言: %s", player, response))
	}
}

// hunterShoot 猎人白天开枪
func (m *ModeratorAgent) hunterShoot(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], hunter string) {
	alivePlayers := m.state.GetAlivePlayers()
	if len(alivePlayers) == 0 {
		return
	}

	promptText := fmt.Sprintf(params.Prompts.ToHunter, hunter)

	// 使用结构化工具
	shootTool := tools.NewShootTool(m.state)
	if shootTool != nil {
		result, err := m.callPlayerWithTool(ctx, hunter, promptText, shootTool)
		if err == nil {
			if shoot, ok := result["shoot"].(bool); ok && shoot {
				if target, ok := result["target"].(string); ok && target != "" {
					m.state.KillPlayer(target)
					// 广播猎人开枪消息
					m.broadcastToAll(fmt.Sprintf(params.Prompts.ToAllHunterShoot, target))
					m.sendMessage(gen, fmt.Sprintf("  🔫 猎人射杀了 %s！", target))
					m.logger.LogHunterShoot(target)

					// 存储猎人开枪到 RAG
					m.storeEpisodeToRAG(ctx, memory.EpisodeHunterShoot, hunter, target,
						fmt.Sprintf("猎人 %s 开枪射杀了 %s", hunter, target))
					return
				}
			}
		}
	}
}

// SpeakingOrderDecision 发言顺序决策
type SpeakingOrderDecision struct {
	Start     string `json:"start"`     // 起始玩家
	Direction string `json:"direction"` // clockwise 或 counterclockwise
	Reason    string `json:"reason"`    // 决策原因
}

// decideSpeakingOrder 主持人决定发言顺序
func (m *ModeratorAgent) decideSpeakingOrder(ctx context.Context, alivePlayers []string) []string {
	if len(alivePlayers) <= 1 {
		return alivePlayers
	}

	// 构建上下文信息
	var contextInfo strings.Builder
	contextInfo.WriteString(fmt.Sprintf("当前回合: %d\n", m.state.Round))

	// 上轮死亡信息
	var lastDead []string
	if m.state.NightKilled != "" && !m.state.NightSaved {
		lastDead = append(lastDead, m.state.NightKilled)
	}
	if m.state.NightPoisoned != "" {
		lastDead = append(lastDead, m.state.NightPoisoned)
	}
	if len(lastDead) > 0 {
		contextInfo.WriteString(fmt.Sprintf("昨晚死亡: %s\n", strings.Join(lastDead, ", ")))
	} else {
		contextInfo.WriteString("昨晚是平安夜\n")
	}

	// 构建 prompt
	prompt := fmt.Sprintf(`你是狼人杀主持人，需要决定本轮发言顺序。

存活玩家（按座位顺序）: %s
%s
请决定发言顺序，输出 JSON 格式:
{
  "start": "从哪个玩家开始发言",
  "direction": "clockwise 或 counterclockwise",
  "reason": "简短说明决策原因"
}

注意：
- start 必须是存活玩家之一
- 可以考虑从死者旁边的玩家开始
- 第一轮可以随机选择`,
		strings.Join(alivePlayers, ", "),
		contextInfo.String(),
	)

	// 调用 LLM 决定顺序
	decision := m.callModeratorLLM(ctx, prompt)

	// 解析决策
	var order SpeakingOrderDecision
	if err := json.Unmarshal([]byte(decision), &order); err != nil {
		// 解析失败，使用默认顺序
		return alivePlayers
	}

	// 验证起始玩家
	startIdx := -1
	for i, p := range alivePlayers {
		if p == order.Start {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		// 起始玩家无效，随机选择
		startIdx = rand.Intn(len(alivePlayers))
	}

	// 生成发言顺序
	result := make([]string, len(alivePlayers))
	for i := 0; i < len(alivePlayers); i++ {
		if order.Direction == "counterclockwise" {
			result[i] = alivePlayers[(startIdx-i+len(alivePlayers))%len(alivePlayers)]
		} else {
			// 默认顺时针
			result[i] = alivePlayers[(startIdx+i)%len(alivePlayers)]
		}
	}

	return result
}

// callModeratorLLM 调用主持人 LLM（用于决策，不是玩家对话）
func (m *ModeratorAgent) callModeratorLLM(ctx context.Context, prompt string) string {
	// 找一个 AI Agent 来调用 LLM（避免使用人类玩家）
	var agent adk.Agent
	for _, a := range m.playerAgents {
		// 检查是否是人类玩家（通过类型断言）
		if _, isHuman := a.(*players.HumanAgent); !isHuman {
			agent = a
			break
		}
	}
	if agent == nil {
		return "{}"
	}

	msgs := []*schema.Message{
		{Role: schema.System, Content: "你是狼人杀游戏主持人，负责公正地主持游戏。请严格按照要求的 JSON 格式输出。"},
		{Role: schema.User, Content: prompt},
	}

	iter := agent.Run(ctx, &adk.AgentInput{
		Messages: msgs,
	})

	var response string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg := event.Output.MessageOutput.Message; msg != nil && msg.Content != "" {
				response = msg.Content
			}
		}
	}

	// 尝试提取 JSON
	response = extractJSON(response)
	return response
}

// extractJSON 从响应中提取 JSON
func extractJSON(s string) string {
	// 查找 { 和 } 的位置
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}

// augmentQueryWithRAG 使用 RAG（语义记忆）和短期记忆（情景记忆）增强查询
func (m *ModeratorAgent) augmentQueryWithRAG(ctx context.Context, baseQuery, playerName, phase string) string {
	var allEpisodes []*memory.Episode

	// 1. 从短期情景记忆获取最近事件（始终可用）
	if m.shortMem != nil {
		// 获取当前轮次的事件
		roundEvents := m.shortMem.GetByRound(m.state.Round)
		allEpisodes = append(allEpisodes, roundEvents...)

		// 获取针对该玩家的怀疑
		if accusers := m.shortMem.GetAccusers(playerName); len(accusers) > 0 {
			// 添加提示：有人怀疑你
			for _, accuser := range accusers {
				allEpisodes = append(allEpisodes, &memory.Episode{
					Type:    memory.EpisodeAccusation,
					Actor:   accuser,
					Target:  playerName,
					Content: fmt.Sprintf("%s 怀疑你", accuser),
					Round:   m.state.Round,
				})
			}
		}
	}

	// 2. 从 RAG 语义记忆检索相关事件（如果可用）
	if m.rag != nil {
		// 构建检索查询
		searchQuery := memory.BuildQueryFromContext(playerName, phase, m.state.Round)

		// 检索相关记忆
		episodes, err := m.rag.RetrieveRelevant(ctx, searchQuery, &memory.RetrieveConfig{
			TopK:     5,
			GameID:   m.logger.GetGameID(),
			MaxRound: m.state.Round,
		})
		if err == nil && len(episodes) > 0 {
			allEpisodes = append(allEpisodes, episodes...)
		}
	}

	// 如果没有任何记忆，返回原始查询
	if len(allEpisodes) == 0 {
		return baseQuery
	}

	// 去重（基于内容）
	seen := make(map[string]bool)
	var uniqueEpisodes []*memory.Episode
	for _, ep := range allEpisodes {
		key := fmt.Sprintf("%s-%s-%s", ep.Type, ep.Actor, ep.Content)
		if !seen[key] {
			seen[key] = true
			uniqueEpisodes = append(uniqueEpisodes, ep)
		}
	}

	// 构建增强 Prompt
	memCtx := &memory.MemoryContext{
		RelevantEpisodes: uniqueEpisodes,
		CurrentRound:     m.state.Round,
		PlayerName:       playerName,
	}
	return memory.BuildAugmentedPrompt(baseQuery, memCtx)
}

// storeEpisodeToRAG 存储事件到 RAG（语义记忆）和短期记忆（情景记忆）
func (m *ModeratorAgent) storeEpisodeToRAG(ctx context.Context, episodeType memory.EpisodeType, actor, target, content string) {
	episode := memory.NewEpisode(
		m.logger.GetGameID(),
		m.state.Round,
		m.state.Phase,
		episodeType,
		actor,
		target,
		content,
	)

	// 存储到短期情景记忆（始终执行）
	if m.shortMem != nil {
		m.shortMem.Add(episode)
	}

	// 存储到 RAG 语义记忆（如果可用）
	if m.rag != nil {
		if err := m.rag.StoreEpisode(ctx, episode); err != nil {
			// 存储失败不影响游戏，只记录警告
			fmt.Printf("⚠️ 存储事件到 RAG 失败: %v\n", err)
		}
	}
}

// playerReflection 玩家反思
func (m *ModeratorAgent) playerReflection(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	m.sendMessage(gen, "\n=== 🎭 玩家反思 ===")

	var wg sync.WaitGroup
	var mu sync.Mutex

	for name := range m.state.Players {
		wg.Add(1)
		go func(playerName string) {
			defer wg.Done()

			response := m.callPlayer(ctx, playerName, params.Prompts.ToAllReflect)

			if response != "" {
				mu.Lock()
				role := string(m.state.GetPlayerRole(playerName))
				m.sendMessage(gen, fmt.Sprintf("  [%s] 反思: %s", playerName, utils.Truncate(response, 200)))
				m.logger.LogReflection(playerName, role, response)
				mu.Unlock()
			}
		}(name)
	}
	wg.Wait()
}
