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
	if memCtx == nil || len(memCtx.RelevantEpisodes) == 0 {
		return basePrompt
	}

	var sb strings.Builder

	sb.WriteString("## 📚 相关历史记忆\n\n")
	sb.WriteString("以下是与当前局势相关的历史信息，请参考这些信息做出判断：\n\n")

	// 按类型分组（重要事件优先）
	var keyEvents []*Episode // 关键事件：查验、死亡、救人、毒杀
	var speeches []*Episode  // 发言
	var votes []*Episode     // 投票

	for _, ep := range memCtx.RelevantEpisodes {
		if ep.Round > memCtx.CurrentRound {
			continue
		}
		switch ep.Type {
		case EpisodeCheck, EpisodeDeath, EpisodeSave, EpisodePoison, EpisodeKill:
			keyEvents = append(keyEvents, ep)
		case EpisodeSpeech:
			speeches = append(speeches, ep)
		case EpisodeVote:
			votes = append(votes, ep)
		default:
			speeches = append(speeches, ep)
		}
	}

	// 输出关键事件（最重要）
	if len(keyEvents) > 0 {
		sb.WriteString("### 🔑 关键事件\n")
		for _, ep := range keyEvents {
			sb.WriteString(formatEpisode(ep))
		}
		sb.WriteString("\n")
	}

	// 输出投票记录
	if len(votes) > 0 {
		sb.WriteString("### 🗳️ 投票记录\n")
		for _, ep := range votes {
			sb.WriteString(formatEpisode(ep))
		}
		sb.WriteString("\n")
	}

	// 输出相关发言（限制数量）
	if len(speeches) > 0 {
		sb.WriteString("### 💬 相关发言\n")
		maxSpeeches := 3
		if len(speeches) < maxSpeeches {
			maxSpeeches = len(speeches)
		}
		for i := 0; i < maxSpeeches; i++ {
			sb.WriteString(formatEpisode(speeches[i]))
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
		return fmt.Sprintf("- 💬 [%s] 发言: \"%s\"\n", ep.Actor, truncateContent(ep.Content, 150))
	case EpisodeVote:
		return fmt.Sprintf("- 🗳️ [%s] 投票给 [%s]\n", ep.Actor, ep.Target)
	case EpisodeAccusation:
		return fmt.Sprintf("- ⚠️ [%s] 指控 [%s]: \"%s\"\n", ep.Actor, ep.Target, truncateContent(ep.Content, 100))
	case EpisodeDeath:
		return fmt.Sprintf("- 💀 [%s] 死亡\n", ep.Actor)
	case EpisodeCheck:
		return fmt.Sprintf("- 🔍 查验 [%s]: %s\n", ep.Target, ep.Content)
	case EpisodeKill:
		return fmt.Sprintf("- 🐺 [%s] 被狼人击杀\n", ep.Target)
	case EpisodeSave:
		return fmt.Sprintf("- 💊 [%s] 被救活\n", ep.Target)
	case EpisodePoison:
		return fmt.Sprintf("- ☠️ [%s] 被毒杀\n", ep.Target)
	default:
		return fmt.Sprintf("- [%s] %s\n", ep.Actor, ep.Content)
	}
}

// truncateContent 截断内容
func truncateContent(s string, maxLen int) string {
	// 移除换行符
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// BuildQueryFromContext 根据上下文构建检索查询
func BuildQueryFromContext(playerName string, phase string, round int) string {
	var queries []string

	// 基础查询：包含玩家名，提高相关性
	queries = append(queries, playerName)

	if phase == "day" {
		// 白天阶段：关注发言、怀疑、投票
		queries = append(queries,
			"发言", "怀疑", "投票", "狼人", "可疑",
			"指控", "辩解", "分析", "逻辑",
		)
	} else {
		// 夜晚阶段：关注行动
		queries = append(queries,
			"击杀", "查验", "救人", "毒杀", "目标",
		)
	}

	// 添加回合信息
	if round > 1 {
		queries = append(queries, fmt.Sprintf("第%d轮", round-1))
	}

	return strings.Join(queries, " ")
}
