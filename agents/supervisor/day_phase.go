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
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"

	"github.com/ashwinyue/wolf-go-adk/game"
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

	// 广播讨论开始
	discussMsg := fmt.Sprintf(params.Prompts.ToAllDiscuss, strings.Join(alivePlayers, ", "), strings.Join(alivePlayers, ", "))
	m.broadcastToAll(discussMsg)

	for _, player := range alivePlayers {
		query := "轮到你发言了，请分析局势并表达你的观点。"

		response := m.callPlayer(ctx, player, query)
		if response != "" {
			m.sendMessage(gen, fmt.Sprintf("  [%s]: %s", player, utils.Truncate(response, 200)))
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

			query := fmt.Sprintf(params.Prompts.ToAllVote, strings.Join(alivePlayers, ", "))

			var target string
			if voteTool != nil {
				result, err := m.callPlayerWithTool(ctx, p, query, voteTool)
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
					return
				}
			}
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
