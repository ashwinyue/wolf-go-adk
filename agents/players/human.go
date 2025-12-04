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

package players

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/wolf-go-adk/game"
)

// HumanAgent 人类玩家 Agent
// 实现 adk.Agent 接口，通过终端获取用户输入
type HumanAgent struct {
	name        string
	role        game.Role
	description string
	reader      *bufio.Reader
}

// NewHumanAgent 创建人类玩家 Agent
func NewHumanAgent(name string, role game.Role) *HumanAgent {
	return &HumanAgent{
		name:        name,
		role:        role,
		description: fmt.Sprintf("人类玩家 %s，角色：%s", name, getRoleDisplayName(role)),
		reader:      bufio.NewReader(os.Stdin),
	}
}

// Name 返回 Agent 名称
func (h *HumanAgent) Name(ctx context.Context) string {
	return h.name
}

// Description 返回 Agent 描述
func (h *HumanAgent) Description(ctx context.Context) string {
	return h.description
}

// Run 运行 Agent（等待用户输入）
func (h *HumanAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer func() {
			if e := recover(); e != nil {
				gen.Send(&adk.AgentEvent{
					Err: fmt.Errorf("recover from panic: %v", e),
				})
			}
			gen.Close()
		}()

		// 显示提示信息
		prompt := h.extractPrompt(input)
		h.displayPrompt(prompt)

		// 等待用户输入
		response, err := h.waitForInput(ctx)
		if err != nil {
			gen.Send(&adk.AgentEvent{
				Err: fmt.Errorf("读取用户输入失败: %w", err),
			})
			return
		}

		// 发送用户响应
		gen.Send(&adk.AgentEvent{
			AgentName: h.name,
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{
					IsStreaming: false,
					Message: &schema.Message{
						Role:    schema.Assistant,
						Content: response,
					},
					Role: schema.Assistant,
				},
			},
		})
	}()

	return iter
}

// extractPrompt 从输入中提取提示信息
func (h *HumanAgent) extractPrompt(input *adk.AgentInput) string {
	if input == nil || len(input.Messages) == 0 {
		return ""
	}

	// 获取最后一条消息作为提示
	lastMsg := input.Messages[len(input.Messages)-1]
	return lastMsg.Content
}

// displayPrompt 显示提示信息给用户
func (h *HumanAgent) displayPrompt(prompt string) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  🎮 轮到你了 [%s - %s]\n", h.name, getRoleDisplayName(h.role))
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	if prompt != "" {
		// 分行显示提示
		lines := strings.Split(prompt, "\n")
		for _, line := range lines {
			if len(line) > 60 {
				// 长行换行显示
				for len(line) > 60 {
					fmt.Printf("║  %s\n", line[:60])
					line = line[60:]
				}
				if len(line) > 0 {
					fmt.Printf("║  %s\n", line)
				}
			} else {
				fmt.Printf("║  %s\n", line)
			}
		}
	}
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  请输入你的回复（按 Enter 确认）:")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Print(">>> ")
}

// waitForInput 等待用户输入
func (h *HumanAgent) waitForInput(ctx context.Context) (string, error) {
	// 检查是否是真正的终端
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		fmt.Println("\n⚠️ 检测到非交互式终端，使用默认响应")
		return "我选择跳过这轮发言。", nil
	}

	// 创建一个 channel 来接收输入
	inputCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		// 使用 bufio.Reader 读取输入
		line, err := h.reader.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		inputCh <- strings.TrimSpace(line)
	}()

	// 等待输入或上下文取消（增加 5 分钟超时）
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(5 * time.Minute):
		return "我选择跳过这轮发言。", nil
	case err := <-errCh:
		return "", err
	case input := <-inputCh:
		if input == "" {
			// 如果用户没有输入，返回默认响应
			return "我选择跳过这轮发言。", nil
		}
		return input, nil
	}
}

// getRoleDisplayName 获取角色显示名称
func getRoleDisplayName(role game.Role) string {
	switch role {
	case game.RoleWerewolf:
		return "🐺 狼人"
	case game.RoleVillager:
		return "👨‍🌾 村民"
	case game.RoleSeer:
		return "🔮 预言家"
	case game.RoleWitch:
		return "🧙‍♀️ 女巫"
	case game.RoleHunter:
		return "🏹 猎人"
	default:
		return string(role)
	}
}

// IsHuman 标识这是人类玩家
func (h *HumanAgent) IsHuman() bool {
	return true
}

// GetRole 获取角色
func (h *HumanAgent) GetRole() game.Role {
	return h.role
}
