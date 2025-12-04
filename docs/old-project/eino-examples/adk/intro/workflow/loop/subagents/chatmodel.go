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
	"log"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/eino-examples/adk/common/model"
)

func NewMainAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "MainAgent",
		Description: "负责尝试解决用户任务的主智能体。",
		Instruction: `你是负责解决用户任务的主智能体。
请始终使用中文回答。根据给定的要求提供全面的解决方案。
专注于提供准确和完整的结果。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewCritiqueAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "CritiqueAgent",
		Description: "审查主智能体工作并提供反馈的评判智能体。",
		Instruction: `你是负责审查主智能体工作的评判智能体。
请始终使用中文回答。分析提供的解决方案的准确性、完整性和质量。
如果你发现问题或需要改进的地方，请提供具体的反馈。
如果工作令人满意，请调用'exit'工具并提供最终的总结回复。`,
		Model: model.NewChatModel(),
		// Exit:  nil, // use default exit tool
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}
