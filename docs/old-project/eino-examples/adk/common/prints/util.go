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

package prints

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-examples/internal/logs"
)

func Event(event *adk.AgentEvent) {
	fmt.Printf("名称: %s\n路径: %s", event.AgentName, event.RunPath)
	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				if m.Role == schema.Tool {
					fmt.Printf("\n工具响应: %s", m.Content)
				} else {
					fmt.Printf("\n回答: %s", m.Content)
				}
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\n工具名称: %s", tc.Function.Name)
					fmt.Printf("\n参数: %s", tc.Function.Arguments)
				}
			}
		} else if s := event.Output.MessageOutput.MessageStream; s != nil {
			toolMap := map[int][]*schema.Message{}
			var contentStart bool
			charNumOfOneRow := 0
			maxCharNumOfOneRow := 120
			for {
				chunk, err := s.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("错误: %v", err)
					return
				}
				if chunk.Content != "" {
					if !contentStart {
						contentStart = true
						if chunk.Role == schema.Tool {
							fmt.Printf("\n工具响应: ")
						} else {
							fmt.Printf("\n回答: ")
						}
					}

					charNumOfOneRow += len(chunk.Content)
					if strings.Contains(chunk.Content, "\n") {
						charNumOfOneRow = 0
					} else if charNumOfOneRow >= maxCharNumOfOneRow {
						fmt.Printf("\n")
						charNumOfOneRow = 0
					}
					fmt.Printf("%v", chunk.Content)
				}

				if len(chunk.ToolCalls) > 0 {
					for _, tc := range chunk.ToolCalls {
						index := tc.Index
						if index == nil {
							logs.Fatalf("索引为空")
						}
						toolMap[*index] = append(toolMap[*index], &schema.Message{
							Role: chunk.Role,
							ToolCalls: []schema.ToolCall{
								{
									ID:    tc.ID,
									Type:  tc.Type,
									Index: tc.Index,
									Function: schema.FunctionCall{
										Name:      tc.Function.Name,
										Arguments: tc.Function.Arguments,
									},
								},
							},
						})
					}
				}
			}

			for _, msgs := range toolMap {
				m, err := schema.ConcatMessages(msgs)
				if err != nil {
					log.Fatalf("连接消息失败: %v", err)
					return
				}
				fmt.Printf("\n工具名称: %s", m.ToolCalls[0].Function.Name)
				fmt.Printf("\n参数: %s", m.ToolCalls[0].Function.Arguments)
			}
		}
	}
	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fmt.Printf("\n操作: 转移到 %v", event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			ii, _ := json.MarshalIndent(event.Action.Interrupted.Data, "  ", "  ")
			fmt.Printf("\n操作: 中断")
			fmt.Printf("\n中断快照: %v", string(ii))
		}
		if event.Action.Exit {
			fmt.Printf("\n操作: 退出")
		}
	}
	if event.Err != nil {
		fmt.Printf("\n错误: %v", event.Err)
	}
	fmt.Println()
	fmt.Println()
}
