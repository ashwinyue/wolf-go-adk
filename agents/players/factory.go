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

package players

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"github.com/ashwinyue/wolf-go-adk/game"
)

// CreatePlayerAgents 创建所有玩家 Agent
// 每个玩家都是独立的 ChatModelAgent，有自己的 ReAct 循环
func CreatePlayerAgents(ctx context.Context, state *game.GameState) (map[string]adk.Agent, error) {
	playerAgents := make(map[string]adk.Agent)

	for name, player := range state.Players {
		var agent adk.Agent
		var err error

		switch player.Role {
		case game.RoleWerewolf:
			agent, err = NewWerewolfAgent(ctx, name, state)
		case game.RoleVillager:
			agent, err = NewVillagerAgent(ctx, name, state)
		case game.RoleSeer:
			agent, err = NewSeerAgent(ctx, name, state)
		case game.RoleWitch:
			agent, err = NewWitchAgent(ctx, name, state)
		case game.RoleHunter:
			agent, err = NewHunterAgent(ctx, name, state)
		default:
			agent, err = NewVillagerAgent(ctx, name, state)
		}

		if err != nil {
			return nil, fmt.Errorf("创建玩家 Agent 失败: %w", err)
		}

		playerAgents[name] = agent
	}

	return playerAgents, nil
}
