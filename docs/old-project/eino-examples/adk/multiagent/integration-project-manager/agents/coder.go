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

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
)

func NewCodeAgent(ctx context.Context, tcm model.ToolCallingChatModel) (adk.Agent, error) {
	type RAGInput struct {
		Query   string  `json:"query" jsonschema_description:"搜索查询"`
		Context *string `json:"context" jsonschema_description:"用户输入上下文"`
	}
	type RAGOutput struct {
		Documents []string `json:"documents"`
	}
	knowledgeBaseTool, err := utils.InferTool(
		"knowledge_base",
		"能够回答常见问题、为答案提供具体理由并提高准确性的知识库",
		func(ctx context.Context, input *RAGInput) (output *RAGOutput, err error) {
			// 替换为真实知识库搜索
			if input.Query == "" {
				return nil, fmt.Errorf("RAG输入查询是必需的")
			}

			return &RAGOutput{
				[]string{
					"Q: What is the difference between a list and a tuple in Python?\nA: A list is mutable, meaning you can modify its elements after creation, while a tuple is immutable and cannot be changed once created. Lists use square brackets [], tuples use parentheses ().",
					"Q: How do you handle exceptions in Java?\nA: You handle exceptions in Java using try-catch blocks. Code that might throw an exception is placed inside the try block, and the catch block handles the exception. Optionally, a finally block can be used for cleanup.",
					"Q: What is the purpose of the async and await keywords in JavaScript?\nA: async marks a function as asynchronous, allowing it to return a Promise. await pauses the execution of an async function until the Promise resolves, enabling easier asynchronous code writing.",
					"Q: How can you optimize SQL queries for better performance?\nA: Common optimizations include creating indexes on frequently queried columns, avoiding SELECT *, using JOINs efficiently, and analyzing query execution plans to identify bottlenecks.",
					"Q: What is dependency injection and why is it useful?\nA: Dependency injection is a design pattern where an object receives its dependencies from an external source rather than creating them itself. It promotes loose coupling, easier testing, and better code maintainability.",
				},
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CodeAgent",
		Description: "CodeAgent专门通过利用知识库作为工具来生成高质量代码。它回忆相关知识和最佳实践，以产生高效、可维护和准确的代码解决方案，满足项目需求。",
		Instruction: `你是CodeAgent。你的职责包括：

- 根据项目需求生成高质量、高效且可维护的代码。
- 利用知识库工具回忆相关的编码标准、模式和最佳实践。
- 确保代码清晰、文档齐全，并满足指定的功能需求。
- 回顾相关知识以提高代码的准确性和质量。
- 沟通你的编码决策，并在必要时提供解释。
- 及时专业地响应用户请求或澄清。

工具处理：
当用户的问题模糊或超出你的回答范围时，请使用knowledge_base工具从知识库中回忆相关结果，并根据结果提供准确答案。
`,
		Model: tcm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{knowledgeBaseTool},
			},
		},
		MaxIterations: 3,
	})
}
