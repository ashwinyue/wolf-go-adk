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

package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GameLogger 游戏日志记录器
type GameLogger struct {
	mu        sync.Mutex
	gameID    string
	startTime time.Time
	fullLog   strings.Builder
	replayLog strings.Builder
}

// NewGameLogger 创建游戏日志记录器
func NewGameLogger() *GameLogger {
	now := time.Now()
	return &GameLogger{
		gameID:    now.Format("20060102_150405"),
		startTime: now,
	}
}

// GetGameID 获取游戏 ID
func (gl *GameLogger) GetGameID() string {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	return gl.gameID
}

// SetPlayers 设置玩家信息
func (gl *GameLogger) SetPlayers(players map[string]Role) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.fullLog.WriteString("# 🐺 狼人杀游戏完整日志\n\n")
	gl.fullLog.WriteString(fmt.Sprintf("**游戏ID**: %s\n\n", gl.gameID))
	gl.fullLog.WriteString(fmt.Sprintf("**开始时间**: %s\n\n", gl.startTime.Format("2006-01-02 15:04:05")))
	gl.fullLog.WriteString("---\n\n")
	gl.fullLog.WriteString("## 📋 角色分配\n\n")
	gl.fullLog.WriteString("| 玩家 | 角色 |\n")
	gl.fullLog.WriteString("|------|------|\n")
	for name, role := range players {
		gl.fullLog.WriteString(fmt.Sprintf("| %s | %s |\n", name, role))
	}
	gl.fullLog.WriteString("\n---\n\n")

	// 回放日志
	gl.replayLog.WriteString("# 🎮 狼人杀游戏回放\n\n")
	gl.replayLog.WriteString(fmt.Sprintf("**游戏ID**: %s\n\n", gl.gameID))
	gl.replayLog.WriteString("## 角色分配\n\n")

	var wolves, villagers, seer, witch, hunter []string
	for name, role := range players {
		switch role {
		case RoleWerewolf:
			wolves = append(wolves, name)
		case RoleVillager:
			villagers = append(villagers, name)
		case RoleSeer:
			seer = append(seer, name)
		case RoleWitch:
			witch = append(witch, name)
		case RoleHunter:
			hunter = append(hunter, name)
		}
	}
	gl.replayLog.WriteString(fmt.Sprintf("- **狼人**: %s\n", strings.Join(wolves, ", ")))
	gl.replayLog.WriteString(fmt.Sprintf("- **村民**: %s\n", strings.Join(villagers, ", ")))
	if len(seer) > 0 {
		gl.replayLog.WriteString(fmt.Sprintf("- **预言家**: %s\n", seer[0]))
	}
	if len(witch) > 0 {
		gl.replayLog.WriteString(fmt.Sprintf("- **女巫**: %s\n", witch[0]))
	}
	if len(hunter) > 0 {
		gl.replayLog.WriteString(fmt.Sprintf("- **猎人**: %s\n", hunter[0]))
	}
	gl.replayLog.WriteString("\n---\n\n")
}

// LogRound 记录回合开始
func (gl *GameLogger) LogRound(round int) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("## 🔄 第 %d 回合\n\n", round))
	gl.replayLog.WriteString(fmt.Sprintf("## 第 %d 回合\n\n", round))
}

// LogPhase 记录阶段
func (gl *GameLogger) LogPhase(phase string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("### %s\n\n", phase))
}

// LogModerator 记录主持人消息
func (gl *GameLogger) LogModerator(message string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("🎭 **主持人**: %s\n\n", message))
}

// LogAction 记录玩家行动（简洁版，只记录回复）
func (gl *GameLogger) LogAction(player, role, action, prompt, response string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	// 简洁格式：只记录玩家发言，不记录提示词
	roleIcon := getRoleIcon(role)
	gl.fullLog.WriteString(fmt.Sprintf("**%s %s**: %s\n\n", roleIcon, player, response))

	if action != "" {
		gl.fullLog.WriteString(fmt.Sprintf("  → %s\n\n", action))
	}
}

// getRoleIcon 获取角色图标
func getRoleIcon(role string) string {
	icons := map[string]string{
		"werewolf": "🐺",
		"villager": "👨‍🌾",
		"seer":     "🔮",
		"witch":    "🧙‍♀️",
		"hunter":   "🎯",
		"狼人":       "🐺",
		"村民":       "👨‍🌾",
		"预言家":      "🔮",
		"女巫":       "🧙‍♀️",
		"猎人":       "🎯",
	}
	if icon, ok := icons[role]; ok {
		return icon
	}
	return "👤"
}

// LogEvent 记录游戏事件（简洁版，用于回放）
func (gl *GameLogger) LogEvent(event string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.replayLog.WriteString(fmt.Sprintf("- %s\n", event))
}

// LogWerewolfDiscussionStart 记录狼人讨论开始
func (gl *GameLogger) LogWerewolfDiscussionStart(wolves []string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString("### 🤝 狼人密谋\n\n")
	gl.replayLog.WriteString(fmt.Sprintf("🤝 狼人密谋 (%s)\n", strings.Join(wolves, ", ")))
}

// LogWerewolfDiscussion 记录狼人讨论
func (gl *GameLogger) LogWerewolfDiscussion(wolf string, round int, message string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	// 统一格式：🐺 **Player1**: 消息内容
	gl.fullLog.WriteString(fmt.Sprintf("🐺 **%s**: %s\n\n", wolf, message))
}

// LogWerewolfIndividualVote 记录单个狼人的投票
func (gl *GameLogger) LogWerewolfIndividualVote(wolf, target string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("- **%s** 投票: %s\n", wolf, target))
}

// LogWerewolfVote 记录狼人投票结果
func (gl *GameLogger) LogWerewolfVote(target, details string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("\n**狼人决定击杀**: %s (%s)\n\n", target, details))
	gl.replayLog.WriteString(fmt.Sprintf("🐺 狼人击杀: %s\n\n", target))
}

// LogSeerCheck 记录预言家查验
func (gl *GameLogger) LogSeerCheck(target, result string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**预言家查验**: %s → %s\n\n", target, result))
	gl.replayLog.WriteString(fmt.Sprintf("🔮 预言家查验 %s: %s\n\n", target, result))
}

// LogWitchSave 记录女巫救人
func (gl *GameLogger) LogWitchSave(target string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**女巫使用解药**: 救活 %s\n\n", target))
	gl.replayLog.WriteString(fmt.Sprintf("💊 女巫救活: %s\n\n", target))
}

// LogWitchPoison 记录女巫毒人
func (gl *GameLogger) LogWitchPoison(target string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**女巫使用毒药**: 毒杀 %s\n\n", target))
	gl.replayLog.WriteString(fmt.Sprintf("☠️ 女巫毒杀: %s\n\n", target))
}

// LogNightSummary 记录夜晚结算
func (gl *GameLogger) LogNightSummary(killed, poisoned, saved, shot string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString("**夜晚结算**:\n")
	if killed != "" {
		if saved != "" {
			gl.fullLog.WriteString(fmt.Sprintf("- 狼人击杀 %s，被女巫救活\n", killed))
		} else {
			gl.fullLog.WriteString(fmt.Sprintf("- 狼人击杀 %s\n", killed))
		}
	}
	if poisoned != "" {
		gl.fullLog.WriteString(fmt.Sprintf("- 女巫毒杀 %s\n", poisoned))
	}
	if shot != "" {
		gl.fullLog.WriteString(fmt.Sprintf("- 猎人射杀 %s\n", shot))
	}
	gl.fullLog.WriteString("\n")
}

// LogDiscussion 记录讨论发言
func (gl *GameLogger) LogDiscussion(player, message string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**[%s]**: %s\n\n", player, message))
}

// LogVote 记录投票
func (gl *GameLogger) LogVote(voter, target string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("- %s → %s\n", voter, target))
}

// LogVoteResult 记录投票结果
func (gl *GameLogger) LogVoteResult(eliminated, details string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if eliminated != "" {
		gl.fullLog.WriteString(fmt.Sprintf("\n**投票结果**: %s 被淘汰 (%s)\n\n", eliminated, details))
		gl.replayLog.WriteString(fmt.Sprintf("🗳️ 投票淘汰: %s\n\n", eliminated))
	} else {
		gl.fullLog.WriteString(fmt.Sprintf("\n**投票结果**: %s\n\n", details))
	}
}

// LogLastWords 记录遗言
func (gl *GameLogger) LogLastWords(player, message string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**[%s 遗言]**: %s\n\n", player, message))
	gl.replayLog.WriteString(fmt.Sprintf("💀 %s 遗言: %s\n\n", player, message))
}

// LogHunterShoot 记录猎人开枪
func (gl *GameLogger) LogHunterShoot(target string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	gl.fullLog.WriteString(fmt.Sprintf("**猎人开枪**: 射杀 %s\n\n", target))
	gl.replayLog.WriteString(fmt.Sprintf("🔫 猎人射杀: %s\n\n", target))
}

// LogWinner 记录胜利者
func (gl *GameLogger) LogWinner(winner Faction, survivors []string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	winnerName := "好人阵营"
	if winner == FactionWerewolf {
		winnerName = "狼人阵营"
	}

	gl.fullLog.WriteString("---\n\n")
	gl.fullLog.WriteString(fmt.Sprintf("## 🏆 游戏结束\n\n**胜利者**: %s\n\n", winnerName))
	gl.fullLog.WriteString(fmt.Sprintf("**存活玩家**: %s\n\n", strings.Join(survivors, ", ")))
	gl.fullLog.WriteString(fmt.Sprintf("**游戏时长**: %s\n\n", time.Since(gl.startTime).Round(time.Second)))

	gl.replayLog.WriteString("---\n\n")
	gl.replayLog.WriteString(fmt.Sprintf("## 🏆 %s 获胜！\n\n", winnerName))
	gl.replayLog.WriteString(fmt.Sprintf("存活: %s\n", strings.Join(survivors, ", ")))
}

// LogReflection 记录玩家反思
func (gl *GameLogger) LogReflection(player, role, message string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	roleIcon := getRoleIcon(role)
	// 移除 LLM 可能添加的 "反思:" 或 "反思：" 前缀
	message = strings.TrimPrefix(message, "反思:")
	message = strings.TrimPrefix(message, "反思：")
	message = strings.TrimSpace(message)
	gl.fullLog.WriteString(fmt.Sprintf("%s **%s**: 💭 %s\n\n", roleIcon, player, message))
}

// Save 保存日志到文件
func (gl *GameLogger) Save() error {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	// 创建日志目录
	logDir := filepath.Join("logs", gl.gameID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 保存完整日志
	fullLogPath := filepath.Join(logDir, "full_log.md")
	if err := os.WriteFile(fullLogPath, []byte(gl.fullLog.String()), 0644); err != nil {
		return fmt.Errorf("保存完整日志失败: %w", err)
	}

	// 保存回放日志
	replayLogPath := filepath.Join(logDir, "replay.md")
	if err := os.WriteFile(replayLogPath, []byte(gl.replayLog.String()), 0644); err != nil {
		return fmt.Errorf("保存回放日志失败: %w", err)
	}

	fmt.Printf("日志已保存到: %s\n", logDir)
	return nil
}
