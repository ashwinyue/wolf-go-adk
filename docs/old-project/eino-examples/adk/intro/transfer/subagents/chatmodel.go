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

package subagents

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino-examples/adk/common/model"
)

type GetWeatherInput struct {
	City string `json:"city"`
}

func NewWeatherAgent() adk.Agent {
	weatherTool, err := utils.InferTool(
		"get_weather",
		"获取指定城市的当前天气。", // 中文描述
		func(ctx context.Context, input *GetWeatherInput) (string, error) {
			return fmt.Sprintf(`The temperature in %s is 25°C`, input.City), nil
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "WeatherAgent",
		Description: "这个智能体可以获取给定城市的当前天气。",
		Instruction: `你的唯一目的是使用'get_weather'工具获取给定城市的当前天气。
请始终使用中文回答。调用工具后，直接向用户报告结果。`,
		Model: model.NewChatModel(),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{weatherTool},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewChatAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "ChatAgent",
		Description: "一个用于处理对话聊天通用目的的智能体。", // 中文描述
		Instruction: `你是一个友好的对话助手。
请始终使用中文回答。你的角色是处理一般聊天并回答与任何特定基于工具的任务无关的问题。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewRouterAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "RouterAgent",
		Description: "一个手动路由器，将任务转移到其他专家智能体。",
		Instruction: `你是一个智能任务路由器。
请始终使用中文回答。你的责任是分析用户的请求并将其委托给最合适的专家智能体。
如果没有智能体可以处理该任务，只需告知用户该任务无法处理。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}
