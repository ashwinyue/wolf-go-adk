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

package agents

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
)

func NewResearchAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	type webSearchInput struct {
		CurrentContext string `json:"current_context" jsonschema_description:"网络搜索的当前上下文"`
	}
	type webSearchOutput struct {
		Result []string
	}
	webSearchTool, err := utils.InferTool(
		"web_search",
		"网络搜索工具",
		func(ctx context.Context, input *webSearchInput) (output *webSearchOutput, err error) {
			// 替换为真实网络搜索工具
			if input.CurrentContext == "" {
				return nil, fmt.Errorf("网络搜索输入是必需的")
			}
			return &webSearchOutput{}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ResearchAgent",
		Description: "ResearchAgent负责进行研究并生成可行的解决方案。它支持中断以接收来自用户的额外上下文信息，这有助于提高研究结果的准确性和相关性。它利用网络搜索工具收集最新信息。",
		Instruction: `你是ResearchAgent。你的职责是：

- 对给定主题或问题进行彻底研究。
- 根据你的发现生成可行且信息充分的解决方案。
- 支持中断，随时接受来自用户的额外上下文或信息，以完善你的研究。
- 有效使用网络搜索工具收集相关和当前的数据。
- 清晰且有逻辑地传达你的研究结果。
- 如有必要，提出澄清问题以提高研究质量。
- 在整个互动过程中保持专业和友好的语气。

工具处理：
- 当你认为输入信息不足以支持研究时，请使用ask_for_clarification工具要求用户补充上下文。
- 如果上下文已满足，你可以使用web_search工具从互联网获取更多数据。
`,
		Model: tcm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{webSearchTool, newAskForClarificationTool()},
			},
		},
		MaxIterations: 5,
	})
}

type askForClarificationOptions struct {
	NewInput *string
}

func WithNewInput(input string) tool.Option {
	return tool.WrapImplSpecificOptFn(func(t *askForClarificationOptions) {
		t.NewInput = &input
	})
}

type AskForClarificationInput struct {
	Question string `json:"question" jsonschema_description:"你想询问用户以获取缺失信息的具体问题"`
}

func newAskForClarificationTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"ask_for_clarification",
		"当用户的请求模糊或缺乏继续进行所需信息时，调用此工具。使用它提出后续问题以获取你需要的详细信息，例如书籍的类型，然后才能有效使用其他工具。",
		func(ctx context.Context, input *AskForClarificationInput, opts ...tool.Option) (output string, err error) {
			o := tool.GetImplSpecificOptions[askForClarificationOptions](nil, opts...)
			if o.NewInput == nil {
				return "", compose.NewInterruptAndRerunErr(input.Question)
			}
			output = *o.NewInput
			o.NewInput = nil
			return output, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}
