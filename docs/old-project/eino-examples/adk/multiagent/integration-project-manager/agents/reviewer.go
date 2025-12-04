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

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

func NewReviewAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	// these sub-agents don't need description because they'll be set in a fixed workflow.
	questionAnalysisAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "question_analysis_agent",
		Description: "问题分析智能体",
		Instruction: `你是问题分析智能体。你的职责包括：

- 分析给定的研究或编码结果，以识别关键问题和评估标准。
- 将复杂问题分解为清晰、可管理的组成部分。
- 突出潜在问题或关注领域。
- 准备结构化框架以指导后续的审查生成。
- 在传递内容之前确保全面理解。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	generateReviewAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "generate_review_agent",
		Description: "生成审查智能体",
		Instruction: `你是生成审查智能体。你的职责是：

- 基于问题分析生成全面且平衡的审查。
- 突出优点、缺点和改进领域。
- 提供建设性和可操作的反馈。
- 在评估中保持客观性和清晰性。
- 为下一步的验证准备审查内容。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	reviewValidationAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "review_validation_agent",
		Description: "审查验证智能体",
		Instruction: `你是审查验证智能体。你的任务是：

- 验证生成的审查的准确性、连贯性和公平性。
- 检查逻辑一致性和完整性。
- 识别任何偏见或错误并提出修正建议。
- 确认审查与原始分析和项目目标一致。
- 批准审查以供最终展示，或在必要时请求修订。`,
		Model: tcm,
	})
	if err != nil {
		return nil, err
	}

	return adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        "ReviewAgent",
		Description: "ReviewAgent负责通过顺序工作流评估研究和编码结果。它协调三个关键步骤——问题分析、审查生成和审查验证——以提供支持项目管理决策的合理评估。",
		SubAgents:   []adk.Agent{questionAnalysisAgent, generateReviewAgent, reviewValidationAgent},
	})
}
