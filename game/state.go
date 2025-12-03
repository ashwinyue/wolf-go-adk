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
	"strings"
	"sync"
)

// Role 角色类型
type Role string

const (
	RoleWerewolf Role = "werewolf" // 狼人
	RoleVillager Role = "villager" // 村民
	RoleSeer     Role = "seer"     // 预言家
	RoleWitch    Role = "witch"    // 女巫
	RoleHunter   Role = "hunter"   // 猎人
)

// Faction 阵营
type Faction string

const (
	FactionWerewolf Faction = "werewolf" // 狼人阵营
	FactionVillager Faction = "villager" // 村民阵营
)

// Player 玩家信息
type Player struct {
	Name  string
	Role  Role
	Alive bool
}

// GameState 游戏状态（对应设计文档的 SessionValues）
type GameState struct {
	mu sync.RWMutex

	// 玩家信息
	Players      map[string]*Player
	AlivePlayers []string

	// 特殊角色
	Seer   string
	Witch  string
	Hunter string

	// 女巫药水状态
	HealingPotion bool // 解药是否可用
	PoisonPotion  bool // 毒药是否可用

	// 夜间状态
	NightKilled   string // 狼人击杀目标
	NightSaved    bool   // 是否被女巫救活
	NightPoisoned string // 女巫毒杀目标
	NightShot     string // 猎人射杀目标

	// 游戏状态
	Round    int
	Phase    string // "night" or "day"
	FirstDay bool
	Winner   Faction
}

// NewGameState 创建游戏状态
func NewGameState() *GameState {
	return &GameState{
		Players:       make(map[string]*Player),
		AlivePlayers:  []string{},
		HealingPotion: true,
		PoisonPotion:  true,
		Round:         0,
		Phase:         "night",
		FirstDay:      true,
	}
}

// InitPlayers 初始化玩家
func (gs *GameState) InitPlayers(names []string, roles []Role) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.AlivePlayers = make([]string, len(names))
	copy(gs.AlivePlayers, names)

	for i, name := range names {
		role := roles[i]
		gs.Players[name] = &Player{
			Name:  name,
			Role:  role,
			Alive: true,
		}

		switch role {
		case RoleSeer:
			gs.Seer = name
		case RoleWitch:
			gs.Witch = name
		case RoleHunter:
			gs.Hunter = name
		}
	}
}

// GetAlivePlayers 获取存活玩家列表
func (gs *GameState) GetAlivePlayers() []string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var alive []string
	for _, name := range gs.AlivePlayers {
		if player, ok := gs.Players[name]; ok && player.Alive {
			alive = append(alive, name)
		}
	}
	return alive
}

// GetAliveWerewolves 获取存活狼人列表
func (gs *GameState) GetAliveWerewolves() []string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var wolves []string
	for name, player := range gs.Players {
		if player.Alive && player.Role == RoleWerewolf {
			wolves = append(wolves, name)
		}
	}
	return wolves
}

// GetAliveVillagers 获取存活村民阵营列表
func (gs *GameState) GetAliveVillagers() []string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var villagers []string
	for name, player := range gs.Players {
		if player.Alive && player.Role != RoleWerewolf {
			villagers = append(villagers, name)
		}
	}
	return villagers
}

// IsAlive 检查玩家是否存活
func (gs *GameState) IsAlive(name string) bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if player, ok := gs.Players[name]; ok {
		return player.Alive
	}
	return false
}

// GetPlayerRole 获取玩家角色
func (gs *GameState) GetPlayerRole(name string) Role {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if player, ok := gs.Players[name]; ok {
		return player.Role
	}
	return ""
}

// KillPlayer 杀死玩家
func (gs *GameState) KillPlayer(name string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if player, ok := gs.Players[name]; ok {
		player.Alive = false
	}

	// 更新存活玩家列表
	var alive []string
	for _, n := range gs.AlivePlayers {
		if n != name {
			alive = append(alive, n)
		}
	}
	gs.AlivePlayers = alive
}

// ResetNightState 重置夜间状态
func (gs *GameState) ResetNightState() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.NightKilled = ""
	gs.NightSaved = false
	gs.NightPoisoned = ""
	gs.NightShot = ""
}

// CheckWinner 检查胜利条件
func (gs *GameState) CheckWinner() Faction {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	aliveWolves := 0
	aliveVillagers := 0

	for _, player := range gs.Players {
		if player.Alive {
			if player.Role == RoleWerewolf {
				aliveWolves++
			} else {
				aliveVillagers++
			}
		}
	}

	// 狼人全灭，村民获胜
	if aliveWolves == 0 {
		return FactionVillager
	}
	// 狼人数量 >= 村民数量，狼人获胜
	if aliveWolves >= aliveVillagers {
		return FactionWerewolf
	}
	return ""
}

// GetRolesString 获取角色分配字符串
func (gs *GameState) GetRolesString() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var parts []string
	for name, player := range gs.Players {
		parts = append(parts, fmt.Sprintf("%s=%s", name, player.Role))
	}
	return strings.Join(parts, ", ")
}

// SetNightKilled 设置狼人击杀目标
func (gs *GameState) SetNightKilled(target string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.NightKilled = target
}

// GetNightKilled 获取狼人击杀目标
func (gs *GameState) GetNightKilled() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.NightKilled
}

// SetNightSaved 设置女巫救人状态
func (gs *GameState) SetNightSaved(saved bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.NightSaved = saved
	if saved {
		gs.HealingPotion = false
	}
}

// SetNightPoisoned 设置女巫毒人目标
func (gs *GameState) SetNightPoisoned(target string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.NightPoisoned = target
	if target != "" {
		gs.PoisonPotion = false
	}
}

// CanUseHealingPotion 检查解药是否可用
func (gs *GameState) CanUseHealingPotion() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.HealingPotion
}

// CanUsePoisonPotion 检查毒药是否可用
func (gs *GameState) CanUsePoisonPotion() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.PoisonPotion
}
