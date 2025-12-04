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
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/wolf-go-adk/agents/players"
	"github.com/ashwinyue/wolf-go-adk/game"
	"github.com/ashwinyue/wolf-go-adk/memory"
	"github.com/ashwinyue/wolf-go-adk/params"
)

// ModeratorAgent 主持人 Agent（自定义实现 adk.Agent 接口）
type ModeratorAgent struct {
	state        *game.GameState
	logger       *game.GameLogger
	playerAgents map[string]adk.Agent
	playerMsgs   map[string][]*schema.Message // 玩家消息历史
	rag          *memory.RAGSystem            // RAG 系统
	mu           sync.RWMutex
}

// NewModeratorAgent 创建主持人 Agent（全 AI 模式）
func NewModeratorAgent(ctx context.Context) (*ModeratorAgent, error) {
	return NewModeratorAgentWithHuman(ctx, "")
}

// NewModeratorAgentWithHuman 创建主持人 Agent，支持人类玩家
// humanPlayer 指定人类玩家的名字，为空则全部为 AI
func NewModeratorAgentWithHuman(ctx context.Context, humanPlayer string) (*ModeratorAgent, error) {
	state := game.NewGameState()
	logger := game.NewGameLogger()

	// 初始化玩家名单
	playerNames := []string{
		"Player1", "Player2", "Player3",
		"Player4", "Player5", "Player6",
		"Player7", "Player8", "Player9",
	}

	// 验证人类玩家名字
	if humanPlayer != "" {
		valid := false
		for _, name := range playerNames {
			if name == humanPlayer {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("无效的玩家名: %s，可选: %v", humanPlayer, playerNames)
		}
	}

	// 角色分配：3狼人 + 3村民 + 1预言家 + 1女巫 + 1猎人
	roles := []game.Role{
		game.RoleWerewolf, game.RoleWerewolf, game.RoleWerewolf,
		game.RoleVillager, game.RoleVillager, game.RoleVillager,
		game.RoleSeer, game.RoleWitch, game.RoleHunter,
	}

	// 洗牌
	rand.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	state.InitPlayers(playerNames, roles)

	// 记录角色分配
	playerRoles := make(map[string]game.Role)
	for name, player := range state.Players {
		playerRoles[name] = player.Role
	}
	logger.SetPlayers(playerRoles)

	// 创建玩家 Agent
	playerAgents, err := players.CreatePlayerAgents(ctx, state, humanPlayer)
	if err != nil {
		return nil, fmt.Errorf("创建玩家 Agent 失败: %w", err)
	}

	// 如果有人类玩家，显示其角色
	if humanPlayer != "" {
		role := state.GetPlayerRole(humanPlayer)
		fmt.Printf("\n🎮 你是 %s，角色是: %s\n\n", humanPlayer, getRoleDisplayName(role))
	}

	// 初始化玩家消息历史
	playerMsgs := make(map[string][]*schema.Message)
	for name, player := range state.Players {
		playerMsgs[name] = []*schema.Message{
			{Role: schema.System, Content: params.BuildPlayerInstruction(name, game.Role(player.Role))},
		}
	}

	// 初始化 RAG 系统（可选，如果环境变量未配置则跳过）
	var rag *memory.RAGSystem
	if milvusAddr := os.Getenv("MILVUS_ADDR"); milvusAddr != "" {
		arkAPIKey := os.Getenv("ARK_API_KEY")
		arkModel := os.Getenv("ARK_MODEL")
		if arkAPIKey != "" && arkModel != "" {
			ragConfig := &memory.RAGConfig{
				MilvusAddr: milvusAddr,
				ArkAPIKey:  arkAPIKey,
				ArkModel:   arkModel,
			}
			rag, err = memory.NewRAGSystem(ctx, ragConfig)
			if err != nil {
				// RAG 初始化失败不影响游戏运行，只记录警告
				fmt.Printf("⚠️ RAG 系统初始化失败: %v\n", err)
				rag = nil
			} else {
				// 设置游戏 ID
				rag.SetGameID(logger.GetGameID())
				fmt.Println("✅ RAG 系统初始化成功")
			}
		}
	}

	return &ModeratorAgent{
		state:        state,
		logger:       logger,
		playerAgents: playerAgents,
		playerMsgs:   playerMsgs,
		rag:          rag,
	}, nil
}

// Name 返回 Agent 名称
func (m *ModeratorAgent) Name(ctx context.Context) string {
	return "Moderator"
}

// Description 返回 Agent 描述
func (m *ModeratorAgent) Description(ctx context.Context) string {
	return "狼人杀游戏主持人 Agent，负责编排游戏流程和协调玩家 Agent"
}

// Run 运行游戏
func (m *ModeratorAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		// panic 恢复（ADK 最佳实践）
		defer func() {
			if e := recover(); e != nil {
				gen.Send(&adk.AgentEvent{
					Err: fmt.Errorf("recover from panic: %v", e),
				})
			}
			gen.Close()
		}()

		// 宣布游戏开始
		m.announceGameStart(gen)

		// 游戏主循环
		for round := 1; round <= params.MaxGameRound; round++ {
			m.state.Round = round
			m.sendMessage(gen, fmt.Sprintf("\n========== 第 %d 回合 ==========", round))
			m.logger.LogRound(round)

			// 夜晚阶段
			m.nightPhase(ctx, gen)

			// 检查胜利条件
			if winner := m.state.CheckWinner(); winner != "" {
				m.announceWinner(gen, winner)
				m.playerReflection(ctx, gen)
				_ = m.logger.Save()
				return
			}

			// 白天阶段
			m.dayPhase(ctx, gen)

			// 检查胜利条件
			if winner := m.state.CheckWinner(); winner != "" {
				m.announceWinner(gen, winner)
				m.playerReflection(ctx, gen)
				_ = m.logger.Save()
				return
			}

			m.state.FirstDay = false
		}

		m.sendMessage(gen, "\n⚠️ 游戏超过最大回合数，强制结束")
		_ = m.logger.Save()
	}()

	return iter
}

// announceGameStart 宣布游戏开始
func (m *ModeratorAgent) announceGameStart(gen *adk.AsyncGenerator[*adk.AgentEvent]) {
	playerNames := m.state.GetAlivePlayers()

	m.sendMessage(gen, "\n=== 🐺 狼人杀游戏开始 🐺 ===")
	m.sendMessage(gen, fmt.Sprintf("玩家: %s", strings.Join(playerNames, ", ")))

	// 广播游戏开始（与原版 to_all_new_game 一致）
	m.broadcastToAll(fmt.Sprintf(params.Prompts.ToAllNewGame, strings.Join(playerNames, ", ")))

	m.sendMessage(gen, "\n=== 角色分配 ===")
	for name, player := range m.state.Players {
		m.sendMessage(gen, fmt.Sprintf("  %s: %s", name, getRoleName(player.Role)))
	}
	m.sendMessage(gen, "=======================")
}

// announceWinner 宣布胜利者
func (m *ModeratorAgent) announceWinner(gen *adk.AsyncGenerator[*adk.AgentEvent], winner game.Faction) {
	rolesStr := m.state.GetRolesString()
	aliveCount := len(m.state.GetAlivePlayers())
	aliveWolves := len(m.state.GetAliveWerewolves())

	m.sendMessage(gen, "\n========================================")
	if winner == game.FactionWerewolf {
		// 广播狼人胜利消息
		msg := fmt.Sprintf(params.Prompts.ToAllWolfWin, aliveCount, aliveWolves, rolesStr)
		m.broadcastToAll(msg)
		m.sendMessage(gen, "🐺 狼人阵营获胜！")
	} else {
		// 广播村民胜利消息
		msg := fmt.Sprintf(params.Prompts.ToAllVillageWin, rolesStr)
		m.broadcastToAll(msg)
		m.sendMessage(gen, "👨‍🌾 好人阵营获胜！")
	}

	m.sendMessage(gen, "\n=== 最终角色揭示 ===")
	for name, player := range m.state.Players {
		status := "存活"
		if !player.Alive {
			status = "死亡"
		}
		m.sendMessage(gen, fmt.Sprintf("  %s: %s (%s)", name, getRoleName(player.Role), status))
	}
	m.sendMessage(gen, "========================================")

	m.logger.LogWinner(winner, m.state.GetAlivePlayers())
}

// getRoleName 获取角色中文名
func getRoleName(role game.Role) string {
	switch role {
	case game.RoleWerewolf:
		return "狼人"
	case game.RoleVillager:
		return "村民"
	case game.RoleSeer:
		return "预言家"
	case game.RoleWitch:
		return "女巫"
	case game.RoleHunter:
		return "猎人"
	default:
		return string(role)
	}
}

// getRoleDisplayName 获取角色显示名称（带 emoji）
func getRoleDisplayName(role game.Role) string {
	switch role {
	case game.RoleWerewolf:
		return "🐺 狼人"
	case game.RoleVillager:
		return "👨‍🌾 村民"
	case game.RoleSeer:
		return "🔮 预言家"
	case game.RoleWitch:
		return "🧙‍♀️ 女巫"
	case game.RoleHunter:
		return "🏹 猎人"
	default:
		return string(role)
	}
}

// Close 关闭资源
func (m *ModeratorAgent) Close() error {
	if m.rag != nil {
		return m.rag.Close()
	}
	return nil
}

// GetRAG 获取 RAG 系统（用于外部访问）
func (m *ModeratorAgent) GetRAG() *memory.RAGSystem {
	return m.rag
}
