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

	"github.com/ashwinyue/wolf-go-adk/game"
	"github.com/ashwinyue/wolf-go-adk/params"
	"github.com/ashwinyue/wolf-go-adk/tools"
	"github.com/ashwinyue/wolf-go-adk/utils"
)

// NewHunterAgent 创建猎人 Agent
func NewHunterAgent(ctx context.Context, name string, state *game.GameState) (adk.Agent, error) {
	instruction := params.BuildPlayerInstruction(name, game.RoleHunter)

	// 猎人工具：开枪、投票
	playerTools := []tool.BaseTool{
		tools.NewShootTool(state),
		tools.NewVoteTool(state),
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        name,
		Description: fmt.Sprintf("玩家 %s，角色：猎人", name),
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
		return nil, fmt.Errorf("创建猎人 Agent %s 失败: %w", name, err)
	}

	return agent, nil
}
