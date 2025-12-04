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
	"time"

	"github.com/cloudwego/eino/schema"
)

// EpisodeType 事件类型
type EpisodeType string

const (
	EpisodeSpeech      EpisodeType = "speech"       // 发言
	EpisodeVote        EpisodeType = "vote"         // 投票
	EpisodeKill        EpisodeType = "kill"         // 击杀
	EpisodeSave        EpisodeType = "save"         // 救人
	EpisodePoison      EpisodeType = "poison"       // 毒人
	EpisodeCheck       EpisodeType = "check"        // 查验
	EpisodeDeath       EpisodeType = "death"        // 死亡
	EpisodeAccusation  EpisodeType = "accusation"   // 指控
	EpisodeLastWords   EpisodeType = "last_words"   // 遗言
	EpisodeHunterShoot EpisodeType = "hunter_shoot" // 猎人开枪
)

// Episode 游戏事件
type Episode struct {
	ID        string      `json:"id"`
	GameID    string      `json:"game_id"`
	Round     int         `json:"round"`
	Phase     string      `json:"phase"` // "night" | "day"
	Type      EpisodeType `json:"type"`
	Actor     string      `json:"actor"`   // 行动者
	Target    string      `json:"target"`  // 目标（可选）
	Content   string      `json:"content"` // 发言内容/行动描述
	Timestamp time.Time   `json:"timestamp"`
	Visible   []string    `json:"visible"` // 可见玩家列表，空表示公开
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
		},
	}
}

// NewEpisode 创建新事件
func NewEpisode(gameID string, round int, phase string, episodeType EpisodeType, actor, target, content string) *Episode {
	return &Episode{
		GameID:    gameID,
		Round:     round,
		Phase:     phase,
		Type:      episodeType,
		Actor:     actor,
		Target:    target,
		Content:   content,
		Timestamp: time.Now(),
	}
}
