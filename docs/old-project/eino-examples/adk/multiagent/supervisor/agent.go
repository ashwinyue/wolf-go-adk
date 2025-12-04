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

package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino-examples/adk/common/model"
	"github.com/cloudwego/eino-examples/flow/agent/multiagent/plan_execute/tools"
)

func buildSearchAgent(ctx context.Context) (adk.Agent, error) {
	m := model.NewChatModel()

	type searchReq struct {
		Query string `json:"query"`
	}

	type searchResp struct {
		Result string `json:"result"`
	}

	search := func(ctx context.Context, req *searchReq) (*searchResp, error) {
		return &searchResp{
			Result: "2024年，美国GDP为29.18万亿美元，纽约州GDP为2.297万亿美元",
		}, nil
	}

	searchTool, err := tools.SafeInferTool("search", "在互联网上搜索信息", search)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "research_agent",
		Description: "负责在互联网上搜索信息的智能体",
		Instruction: `
		你是一个研究智能体。


        指令:
        - 仅协助与研究相关的任务，不要进行任何数学计算
        - 完成任务后，直接回复给主管
        - 仅回复你的工作结果，不要包含任何其他文本。`,
		Model: m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{searchTool},
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("未知工具: %s", name), nil
				},
			},
		},
	})
}

func buildMathAgent(ctx context.Context) (adk.Agent, error) {
	m := model.NewChatModel()

	type addReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type addResp struct {
		Result float64
	}

	add := func(ctx context.Context, req *addReq) (*addResp, error) {
		return &addResp{
			Result: req.A + req.B,
		}, nil
	}

	addTool, err := tools.SafeInferTool("add", "加法运算", add)
	if err != nil {
		return nil, err
	}

	type multiplyReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type multiplyResp struct {
		Result float64
	}

	multiply := func(ctx context.Context, req *multiplyReq) (*multiplyResp, error) {
		return &multiplyResp{
			Result: req.A * req.B,
		}, nil
	}

	multiplyTool, err := tools.SafeInferTool("multiply", "乘法运算", multiply)
	if err != nil {
		return nil, err
	}

	type divideReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type divideResp struct {
		Result float64
	}

	divide := func(ctx context.Context, req *divideReq) (*divideResp, error) {
		return &divideResp{
			Result: req.A / req.B,
		}, nil
	}

	divideTool, err := tools.SafeInferTool("divide", "除法运算", divide)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "math_agent",
		Description: "负责进行数学计算的智能体",
		Instruction: `
		你是一个数学智能体。


        指令:
        - 仅协助与数学相关的任务
        - 完成任务后，直接回复给主管
        - 仅回复你的工作结果，不要包含任何其他文本。`,
		Model: m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{addTool, multiplyTool, divideTool},
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("未知工具: %s", name), nil
				},
			},
		},
	})
}

func buildSupervisor(ctx context.Context) (adk.Agent, error) {
	m := model.NewChatModel()

	sv, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "负责监督任务的智能体",
		Instruction: `
		你是一个主管，管理两个智能体：

        - 一个研究智能体。将与研究相关的任务分配给此智能体
        - 一个数学智能体。将与数学相关的任务分配给此智能体
        一次只将工作分配给一个智能体，不要并行调用智能体。
        不要自己做任何工作。`,
		Model: m,
		Exit:  &adk.ExitTool{},
	})
	if err != nil {
		return nil, err
	}

	searchAgent, err := buildSearchAgent(ctx)
	if err != nil {
		return nil, err
	}
	mathAgent, err := buildMathAgent(ctx)
	if err != nil {
		return nil, err
	}

	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  []adk.Agent{searchAgent, mathAgent},
	})
}
