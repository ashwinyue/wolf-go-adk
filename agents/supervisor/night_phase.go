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
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/wolf-go-adk/game"
	"github.com/ashwinyue/wolf-go-adk/memory"
	"github.com/ashwinyue/wolf-go-adk/params"
	"github.com/ashwinyue/wolf-go-adk/tools"
	"github.com/ashwinyue/wolf-go-adk/utils"
)

// nightPhase 夜晚阶段
func (m *ModeratorAgent) nightPhase(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	m.sendMessage(gen, "\n--- 🌙 夜晚阶段 ---")
	m.state.ResetNightState()
	m.state.Phase = "night"
	m.logger.LogPhase("🌙 夜间阶段")

	// 广播夜间开始
	m.broadcastToAll(params.Prompts.ToAllNight)
	m.logger.LogModerator("天黑了，请所有人闭眼。")

	// 1. 狼人行动
	m.logger.LogModerator("狼人请睁眼，请选择今晚要击杀的玩家。")
	m.werewolfAction(ctx, gen)

	// 2. 女巫行动
	m.logger.LogModerator("女巫请睁眼。")
	m.witchAction(ctx, gen)

	// 3. 预言家行动
	m.logger.LogModerator("预言家请睁眼，请选择要查验的玩家。")
	m.seerAction(ctx, gen)

	// 4. 结算夜晚
	m.logger.LogModerator("天亮了，请所有人睁眼。")
	m.resolveNight(ctx, gen)
}

// werewolfAction 狼人行动
func (m *ModeratorAgent) werewolfAction(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	wolves := m.state.GetAliveWerewolves()
	if len(wolves) == 0 {
		return
	}

	alivePlayers := m.state.GetAlivePlayers()
	nWolves := len(wolves)

	// 广播讨论开始
	discussionPrompt := fmt.Sprintf(params.Prompts.ToWolvesDiscussion,
		strings.Join(wolves, ", "), strings.Join(alivePlayers, ", "))
	m.broadcastToWerewolves(discussionPrompt)

	m.sendMessage(gen, fmt.Sprintf("  狼人 (%s) 正在讨论...", strings.Join(wolves, ", ")))
	m.logger.LogWerewolfDiscussionStart(wolves)

	// 创建讨论工具
	discussTool := tools.NewDiscussTool()

	// 狼人多轮讨论（MaxDiscussionRound * 狼人数量）
	reachAgreement := false
	for round := 1; round <= params.MaxDiscussionRound*nWolves; round++ {
		wolfIdx := (round - 1) % nWolves
		wolf := wolves[wolfIdx]

		// 构建带历史的提示
		history := m.formatWerewolfHistory(wolf)
		// 使用标准提示词，引导狼人进行真正的讨论和思考
		basePrompt := fmt.Sprintf(params.Prompts.ToWolvesDiscussion,
			strings.Join(wolves, ", "), strings.Join(alivePlayers, ", "))
		promptText := basePrompt + history

		// 使用结构化工具调用
		if discussTool != nil {
			result, err := m.callPlayerWithTool(ctx, wolf, promptText, discussTool)
			if err == nil {
				message := ""
				if msg, ok := result["message"].(string); ok {
					message = msg
				}
				m.sendMessage(gen, fmt.Sprintf("  [%s] (狼人第%d轮): %s", wolf, round, utils.Truncate(message, 200)))
				m.broadcastToWerewolves(fmt.Sprintf("[%s]: %s", wolf, message))
				m.logger.LogWerewolfDiscussion(wolf, round, message)

				// 检查是否达成一致
				if round%nWolves == 0 {
					if agree, ok := result["reach_agreement"].(bool); ok && agree {
						reachAgreement = true
						m.sendMessage(gen, "  ✅ 狼人达成一致！")
						break
					}
				}
				continue
			}
		}

		// 回退到普通调用
		response := m.callPlayer(ctx, wolf, promptText)
		if response != "" {
			m.sendMessage(gen, fmt.Sprintf("  [%s] (狼人第%d轮): %s", wolf, round, utils.Truncate(response, 200)))
			m.broadcastToWerewolves(fmt.Sprintf("[%s]: %s", wolf, response))
			m.logger.LogWerewolfDiscussion(wolf, round, response)
		}
	}

	if !reachAgreement {
		m.sendMessage(gen, "  ⚠️ 狼人未达成一致，将进行投票")
	}

	// 狼人投票（并行）
	m.broadcastToWerewolves(params.Prompts.ToWolvesVote)
	m.sendMessage(gen, "  狼人投票中...")
	m.logger.LogPhase("🗳️ 狼人投票")

	votes := make(map[string]string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 创建投票工具
	voteTool := tools.NewVoteTool(m.state)

	for _, wolf := range wolves {
		wg.Add(1)
		go func(w string) {
			defer wg.Done()

			var target string
			if voteTool != nil {
				result, err := m.callPlayerWithTool(ctx, w, params.Prompts.ToWolvesVote, voteTool)
				if err == nil {
					if t, ok := result["target"].(string); ok {
						target = t
					}
				}
			}

			if target != "" {
				mu.Lock()
				votes[w] = target
				mu.Unlock()
				m.sendMessage(gen, fmt.Sprintf("  [%s] 投票: %s", w, target))
				m.logger.LogWerewolfIndividualVote(w, target)
			}
		}(wolf)
	}
	wg.Wait()

	// 统计投票结果
	if len(votes) > 0 {
		killed, details := utils.MajorityVote(votes)
		m.state.SetNightKilled(killed)
		m.broadcastToWerewolves(fmt.Sprintf(params.Prompts.ToWolvesRes, details, killed))
		m.sendMessage(gen, fmt.Sprintf("  ➡️ 狼人决定杀: %s (%s)", killed, details))
		m.logger.LogWerewolfVote(killed, details)

		// 存储狼人击杀到 RAG
		m.storeEpisodeToRAG(ctx, memory.EpisodeKill, strings.Join(wolves, ","), killed,
			fmt.Sprintf("狼人选择击杀 %s", killed))
	}
}

// witchAction 女巫行动
func (m *ModeratorAgent) witchAction(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	witch := m.state.Witch
	if witch == "" || !m.state.IsAlive(witch) {
		return
	}

	// 广播女巫轮次
	m.broadcastToAll(params.Prompts.ToAllWitchTurn)
	m.sendMessage(gen, fmt.Sprintf("  女巫 (%s) 正在决定...", witch))
	killed := m.state.GetNightKilled()
	resurrected := false

	// 救人决策
	if killed != "" && m.state.CanUseHealingPotion() && killed != witch {
		promptText := fmt.Sprintf(params.Prompts.ToWitchResurrect, witch, killed, killed)

		saveTool := tools.NewSaveTool(m.state)
		if saveTool != nil {
			result, err := m.callPlayerWithTool(ctx, witch, promptText, saveTool)
			if err == nil {
				if save, ok := result["save"].(bool); ok && save {
					m.state.SetNightSaved(true) // 内部会设置 HealingPotion = false
					resurrected = true
					m.broadcastToAll(params.Prompts.ToWitchResurrectYes)
					m.sendMessage(gen, fmt.Sprintf("  ➡️ 女巫救了 %s！", killed))
					m.logger.LogWitchSave(killed)

					// 存储女巫救人到 RAG
					m.storeEpisodeToRAG(ctx, memory.EpisodeSave, witch, killed,
						fmt.Sprintf("女巫使用解药救了 %s", killed))
				} else {
					m.broadcastToAll(params.Prompts.ToWitchResurrectNo)
				}
			}
		}
	}

	// 毒人决策（同晚不能同时救毒）
	if m.state.CanUsePoisonPotion() && !resurrected {
		promptText := fmt.Sprintf(params.Prompts.ToWitchPoison, witch)

		poisonTool := tools.NewPoisonTool(m.state)
		if poisonTool != nil {
			result, err := m.callPlayerWithTool(ctx, witch, promptText, poisonTool)
			if err == nil {
				if poison, ok := result["poison"].(bool); ok && poison {
					if target, ok := result["target"].(string); ok && target != "" && target != witch {
						m.state.SetNightPoisoned(target) // 内部会设置 PoisonPotion = false
						m.sendMessage(gen, fmt.Sprintf("  ➡️ 女巫毒了 %s！", target))
						m.logger.LogWitchPoison(target)

						// 存储女巫毒人到 RAG
						m.storeEpisodeToRAG(ctx, memory.EpisodePoison, witch, target,
							fmt.Sprintf("女巫使用毒药毒杀 %s", target))
					}
				}
			}
		}
	}
}

// seerAction 预言家行动
func (m *ModeratorAgent) seerAction(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	seer := m.state.Seer
	if seer == "" || !m.state.IsAlive(seer) {
		return
	}

	// 广播预言家轮次
	m.broadcastToAll(params.Prompts.ToAllSeerTurn)
	m.sendMessage(gen, fmt.Sprintf("  预言家 (%s) 正在查验...", seer))
	promptText := fmt.Sprintf(params.Prompts.ToSeer, seer)

	// 使用结构化工具
	checkTool := tools.NewCheckTool(m.state)
	var target string

	if checkTool != nil {
		result, err := m.callPlayerWithTool(ctx, seer, promptText, checkTool)
		if err == nil {
			if t, ok := result["target"].(string); ok {
				target = t
			}
		}
	}

	if target != "" {
		player := m.state.Players[target]
		result := string(player.Role)
		resultMsg := fmt.Sprintf(params.Prompts.ToSeerResult, target, result)
		m.addToPlayerHistory(seer, schema.User, resultMsg)
		m.sendMessage(gen, fmt.Sprintf("  ➡️ 预言家查验 %s: %s", target, result))
		m.logger.LogSeerCheck(target, result)

		// 存储预言家查验到 RAG（重要信息！）
		isWolf := "好人"
		if player.Role == game.RoleWerewolf {
			isWolf = "狼人"
		}
		m.storeEpisodeToRAG(ctx, memory.EpisodeCheck, seer, target,
			fmt.Sprintf("预言家查验 %s 的身份是 %s", target, isWolf))
	}
}

// resolveNight 结算夜晚
func (m *ModeratorAgent) resolveNight(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	var dead []string
	var saved string
	var shot string

	killed := m.state.GetNightKilled()
	if killed != "" && !m.state.NightSaved {
		// 检查猎人是否被狼人杀死（非毒杀）
		if m.state.GetPlayerRole(killed) == game.RoleHunter && m.state.NightPoisoned != killed {
			if m.state.IsAlive(killed) {
				shot = m.hunterShootNight(ctx, gen, killed)
				m.state.NightShot = shot // 记录猎人开枪目标，供白天阶段使用
			}
		}

		dead = append(dead, killed)
		m.state.KillPlayer(killed)

		if shot != "" {
			dead = append(dead, shot)
			m.state.KillPlayer(shot)
		}
	} else if m.state.NightSaved {
		saved = killed
	}

	if m.state.NightPoisoned != "" {
		dead = append(dead, m.state.NightPoisoned)
		m.state.KillPlayer(m.state.NightPoisoned)
	}

	m.logger.LogNightSummary(killed, m.state.NightPoisoned, saved, shot)

	if len(dead) > 0 {
		m.sendMessage(gen, fmt.Sprintf("  ☠️ 夜晚结算，死亡: %s", strings.Join(dead, ", ")))
	} else {
		m.sendMessage(gen, "  ✨ 平安夜，无人死亡。")
	}
}

// hunterShootNight 猎人夜间被杀时开枪
func (m *ModeratorAgent) hunterShootNight(ctx context.Context, gen *adk.AsyncGenerator[*adk.AgentEvent], hunter string) string {
	alivePlayers := m.state.GetAlivePlayers()
	var targets []string
	for _, p := range alivePlayers {
		if p != hunter {
			targets = append(targets, p)
		}
	}
	if len(targets) == 0 {
		return ""
	}

	promptText := fmt.Sprintf(params.Prompts.ToHunter, hunter)

	// 使用结构化工具
	shootTool := tools.NewShootTool(m.state)
	if shootTool != nil {
		result, err := m.callPlayerWithTool(ctx, hunter, promptText, shootTool)
		if err == nil {
			if shoot, ok := result["shoot"].(bool); ok && shoot {
				if target, ok := result["target"].(string); ok && target != "" {
					m.sendMessage(gen, fmt.Sprintf("  🔫 猎人射杀了 %s！", target))
					m.logger.LogHunterShoot(target)

					// 存储猎人开枪到 RAG
					m.storeEpisodeToRAG(ctx, memory.EpisodeHunterShoot, hunter, target,
						fmt.Sprintf("猎人 %s 开枪射杀了 %s", hunter, target))
					return target
				}
			}
		}
	}
	return ""
}

// callPlayer 调用玩家（保留消息历史）
func (m *ModeratorAgent) callPlayer(ctx context.Context, playerName, promptText string) string {
	m.mu.Lock()
	msgs := m.playerMsgs[playerName]
	msgs = append(msgs, &schema.Message{Role: schema.User, Content: promptText})
	m.playerMsgs[playerName] = msgs
	m.mu.Unlock()

	agent := m.playerAgents[playerName]
	if agent == nil {
		return ""
	}

	iter := agent.Run(ctx, &adk.AgentInput{
		Messages: msgs,
	})

	var response string
	var responseBuilder strings.Builder
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		// 处理错误事件
		if event.Err != nil {
			fmt.Printf("  ⚠️ [%s] 调用错误: %v\n", playerName, event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg := event.Output.MessageOutput.Message; msg != nil && msg.Content != "" {
				// 如果是流式输出，累加内容
				if event.Output.MessageOutput.IsStreaming {
					responseBuilder.WriteString(msg.Content)
				} else {
					// 非流式输出，直接使用完整内容
					response = msg.Content
				}
			}
		}
	}
	// 如果有流式内容，使用累加的结果
	if responseBuilder.Len() > 0 {
		response = responseBuilder.String()
	}

	// 保存响应到历史
	if response != "" {
		m.mu.Lock()
		m.playerMsgs[playerName] = append(m.playerMsgs[playerName], &schema.Message{Role: schema.Assistant, Content: response})
		m.mu.Unlock()
		// 注意：日志记录由各个阶段的专门方法处理，避免重复
	}

	return response
}

// callPlayerWithTool 使用工具调用玩家
// 注意：ADK 的 ChatModelAgent 会自动处理工具调用，返回的是工具执行后的最终响应
func (m *ModeratorAgent) callPlayerWithTool(ctx context.Context, playerName, promptText string, _ interface{}) (map[string]interface{}, error) {
	response := m.callPlayer(ctx, playerName, promptText)
	if response == "" {
		return nil, fmt.Errorf("empty response")
	}

	// 尝试解析 JSON 响应（工具输出通常是 JSON 格式）
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果不是 JSON，尝试从文本中提取关键信息
		result = make(map[string]interface{})
		result["message"] = response
		result["raw"] = response

		// 尝试从文本中提取常见字段
		responseLower := strings.ToLower(response)

		// 检测是否同意/达成一致
		if strings.Contains(responseLower, "agree") || strings.Contains(response, "同意") ||
			strings.Contains(response, "一致") || strings.Contains(responseLower, "yes") {
			result["reach_agreement"] = true
		}

		// 检测是否救人
		if strings.Contains(responseLower, "save") || strings.Contains(response, "救") {
			result["save"] = true
		}

		// 检测是否毒人
		if strings.Contains(responseLower, "poison") || strings.Contains(response, "毒") {
			result["poison"] = true
		}

		// 检测是否开枪
		if strings.Contains(responseLower, "shoot") || strings.Contains(response, "射") ||
			strings.Contains(response, "开枪") {
			result["shoot"] = true
		}

		// 尝试提取目标玩家名
		for i := 1; i <= 9; i++ {
			playerName := fmt.Sprintf("Player%d", i)
			if strings.Contains(response, playerName) {
				result["target"] = playerName
				break
			}
		}
	}

	return result, nil
}

// addToPlayerHistory 添加消息到玩家历史
func (m *ModeratorAgent) addToPlayerHistory(playerName string, role schema.RoleType, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerMsgs[playerName] = append(m.playerMsgs[playerName], &schema.Message{Role: role, Content: content})
}

// broadcastToWerewolves 广播消息给所有狼人
func (m *ModeratorAgent) broadcastToWerewolves(content string) {
	wolves := m.state.GetAliveWerewolves()
	for _, wolf := range wolves {
		m.addToPlayerHistory(wolf, schema.User, content)
	}
}

// formatWerewolfHistory 格式化狼人历史消息
func (m *ModeratorAgent) formatWerewolfHistory(wolf string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs := m.playerMsgs[wolf]
	if len(msgs) <= 1 {
		return ""
	}

	var history strings.Builder
	history.WriteString("\n\n[Previous discussion]:\n")
	for _, msg := range msgs[1:] { // 跳过系统消息
		if msg.Role == schema.User {
			history.WriteString(fmt.Sprintf("Moderator: %s\n", utils.Truncate(msg.Content, 100)))
		} else if msg.Role == schema.Assistant {
			history.WriteString(fmt.Sprintf("You: %s\n", utils.Truncate(msg.Content, 100)))
		}
	}
	return history.String()
}
