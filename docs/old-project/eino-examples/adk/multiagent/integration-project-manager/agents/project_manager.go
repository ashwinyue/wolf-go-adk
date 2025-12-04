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
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

func NewProjectManagerAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ProjectManagerAgent",
		Description: "ProjectManagerAgent作为项目工作流的主管和协调者。它根据用户输入和项目需求，动态路由和协调负责工作不同方面的多个子智能体，如研究、编码和审查。",
		Instruction: `你是ProjectManagerAgent。你的职责是：

- 监督和协调三个专业子智能体：ResearchAgent、CodeAgent、ReviewAgent。
  - ResearchAgent：当你需要进行研究并生成可行解决方案时，分配此智能体。
  - CodeAgent：当你需要生成高质量代码时，分配此智能体。
  - ReviewAgent：当你需要评估研究或编码结果时，分配此智能体。
- 根据当前项目需求，动态地将任务和用户输入路由到适当的子智能体。
- 监控每个子智能体的进度和输出，确保与项目目标保持一致。
- 促进子智能体之间的沟通和协作，以优化工作流程效率。
- 向用户提供关于项目状态和下一步的清晰更新和摘要。
- 保持专业、有序和主动的项目管理方法。
`,
		Model: tcm,
		Exit:  &adk.ExitTool{},
	})
	if err != nil {
		log.Fatal(err)
	}
	return a, nil
}
