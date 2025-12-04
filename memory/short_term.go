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
	"strings"
	"sync"
)

// ShortTermMemory 短期情景记忆
// 存储最近的游戏事件，按时间顺序组织
type ShortTermMemory struct {
	mu       sync.RWMutex
	episodes []*Episode
	maxSize  int

	// 怀疑关系图：被怀疑者 -> 怀疑者列表
	accusations map[string][]string
}

// NewShortTermMemory 创建短期记忆
func NewShortTermMemory(maxSize int) *ShortTermMemory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &ShortTermMemory{
		episodes:    make([]*Episode, 0),
		maxSize:     maxSize,
		accusations: make(map[string][]string),
	}
}

// Add 添加事件
func (m *ShortTermMemory) Add(ep *Episode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.episodes = append(m.episodes, ep)
	if len(m.episodes) > m.maxSize {
		m.episodes = m.episodes[1:]
	}

	// 如果是怀疑事件，更新怀疑关系图
	if ep.Type == EpisodeAccusation && ep.Target != "" {
		m.accusations[ep.Target] = appendUnique(m.accusations[ep.Target], ep.Actor)
	}
}

// GetRecent 获取最近 N 条事件
func (m *ShortTermMemory) GetRecent(n int) []*Episode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n > len(m.episodes) {
		n = len(m.episodes)
	}
	if n <= 0 {
		return nil
	}

	result := make([]*Episode, n)
	copy(result, m.episodes[len(m.episodes)-n:])
	return result
}

// GetByRound 获取指定轮次的事件
func (m *ShortTermMemory) GetByRound(round int) []*Episode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Episode
	for _, ep := range m.episodes {
		if ep.Round == round {
			result = append(result, ep)
		}
	}
	return result
}

// GetByPlayer 获取指定玩家相关的事件
func (m *ShortTermMemory) GetByPlayer(player string) []*Episode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Episode
	for _, ep := range m.episodes {
		if ep.Actor == player || ep.Target == player {
			result = append(result, ep)
		}
	}
	return result
}

// GetAccusers 获取怀疑某玩家的人
func (m *ShortTermMemory) GetAccusers(player string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if accusers, ok := m.accusations[player]; ok {
		result := make([]string, len(accusers))
		copy(result, accusers)
		return result
	}
	return nil
}

// GetAccused 获取某玩家怀疑的人
func (m *ShortTermMemory) GetAccused(player string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []string
	for target, accusers := range m.accusations {
		for _, accuser := range accusers {
			if accuser == player {
				result = appendUnique(result, target)
				break
			}
		}
	}
	return result
}

// GetCurrentRoundSummary 获取当前轮次摘要
func (m *ShortTermMemory) GetCurrentRoundSummary(round int) string {
	events := m.GetByRound(round)
	if len(events) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("### 📋 本轮事件\n")

	for _, ep := range events {
		sb.WriteString(formatEpisode(ep))
	}

	return sb.String()
}

// Clear 清空记忆
func (m *ShortTermMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.episodes = make([]*Episode, 0)
	m.accusations = make(map[string][]string)
}

// appendUnique 添加唯一元素
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// AccusationKeywords 怀疑关键词
var AccusationKeywords = []string{
	"怀疑", "可疑", "狼人", "有问题", "不对劲",
	"撒谎", "投票淘汰", "投他", "投她", "跳狼",
	"suspect", "suspicious", "werewolf", "vote",
}

// DetectAccusations 检测发言中的怀疑关系
func DetectAccusations(speaker, content string, players []string) []string {
	var accused []string
	contentLower := strings.ToLower(content)

	for _, player := range players {
		if player == speaker {
			continue
		}

		// 检查是否提到该玩家
		if !strings.Contains(content, player) {
			continue
		}

		// 检查是否包含怀疑关键词
		for _, kw := range AccusationKeywords {
			if strings.Contains(contentLower, strings.ToLower(kw)) {
				accused = appendUnique(accused, player)
				break
			}
		}
	}

	return accused
}
