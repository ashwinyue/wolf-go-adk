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

package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-examples/adk/common/model"
	"github.com/cloudwego/eino-examples/adk/multiagent/plan-execute-replan/tools"
)

func NewPlanner(ctx context.Context) (adk.Agent, error) {

	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: model.NewChatModel(),
	})
}

var executorPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`你是一个勤奋细致的旅行研究执行者，遵循给定的计划，仔细而彻底地执行你的任务。
使用可用的工具执行每个规划步骤。
对于天气查询，使用get_weather工具。
对于航班搜索，使用search_flights工具。
对于酒店搜索，使用search_hotels工具。
对于景点研究，使用search_attractions工具。
对于用户澄清，使用ask_for_clarification工具。在总结中，重复问题和结果以确认用户，尽量避免打扰用户。
为每个任务提供详细结果。
可以调用多个工具来获取最终结果。`),
	schema.UserMessage(`## 目标
{input}
## 根据以下计划：
{plan}
## 已完成步骤和结果
{executed_steps}
## 你的任务是执行第一步，即：
{step}`))

func formatInput(in []adk.Message) string {
	return in[0].Content
}

func formatExecutedSteps(in []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for idx, m := range in {
		sb.WriteString(fmt.Sprintf("## %d. Step: %v\n  Result: %v\n\n", idx+1, m.Step, m.Result))
	}
	return sb.String()
}

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	// Get travel tools for the executor
	travelTools, err := tools.GetAllTravelTools(ctx)
	if err != nil {
		return nil, err
	}

	return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: model.NewChatModel(),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: travelTools,
			},
		},

		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			planContent, err_ := in.Plan.MarshalJSON()
			if err_ != nil {
				return nil, err_
			}

			firstStep := in.Plan.FirstStep()

			msgs, err_ := executorPrompt.Format(ctx, map[string]any{
				"input":          formatInput(in.UserInput),
				"plan":           string(planContent),
				"executed_steps": formatExecutedSteps(in.ExecutedSteps),
				"step":           firstStep,
			})
			if err_ != nil {
				return nil, err_
			}

			return msgs, nil
		},
	})
}

func NewReplanAgent(ctx context.Context) (adk.Agent, error) {
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: model.NewChatModel(),
	})
}
