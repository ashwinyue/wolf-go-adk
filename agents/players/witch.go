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
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino-examples/week11-homework/werewolves-adk/game"
	"github.com/cloudwego/eino-examples/week11-homework/werewolves-adk/params"
	"github.com/cloudwego/eino-examples/week11-homework/werewolves-adk/tools"
	"github.com/cloudwego/eino-examples/week11-homework/werewolves-adk/utils"
)

// NewWitchAgent 创建女巫 Agent
func NewWitchAgent(ctx context.Context, name string, state *game.GameState) (adk.Agent, error) {
	instruction := params.BuildPlayerInstruction(name, game.RoleWitch)

	// 女巫工具：救人、毒人、投票
	playerTools := []tool.BaseTool{
		tools.NewSaveTool(state),
		tools.NewPoisonTool(state),
		tools.NewVoteTool(state),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        name,
		Description: fmt.Sprintf("玩家 %s，角色：女巫", name),
		Instruction: instruction,
		Model:       utils.MustNewChatModel(ctx),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: playerTools,
			},
		},
		MaxIterations: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("创建女巫 Agent %s 失败: %w", name, err)
	}

	return agent, nil
}
