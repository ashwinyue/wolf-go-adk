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

func NewPlanAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "PlannerAgent",
		Description: "根据主题生成研究计划。",
		Instruction: `
你是一个专业的研究规划师。
你的目标是为给定主题创建一个全面的、分步骤的研究计划。
计划应该逻辑清晰、易于理解。
用户将提供研究主题。你的输出必须仅包含研究计划本身，不包含任何对话文本、介绍或总结。`,
		Model:     model.NewChatModel(),
		OutputKey: "Plan",
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewWriterAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "WriterAgent",
		Description: "根据研究计划撰写报告。",
		Instruction: `
你是一个专业的学术写作专家。
你将获得一个详细的研究计划：
{Plan}

你的任务是扩展这个计划，撰写一份全面的、结构良好的、深入的报告。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}
