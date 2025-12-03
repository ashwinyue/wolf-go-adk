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

package supervisor

import (
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// sendMessage 发送消息事件
func (m *ModeratorAgent) sendMessage(gen *adk.AsyncGenerator[*adk.AgentEvent], content string) {
	gen.Send(&adk.AgentEvent{
		AgentName: "Moderator",
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message: &schema.Message{
					Role:    schema.Assistant,
					Content: content,
				},
				Role: schema.Assistant,
			},
		},
	})
}

// broadcastToAll 广播消息给所有玩家
func (m *ModeratorAgent) broadcastToAll(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name := range m.playerMsgs {
		m.playerMsgs[name] = append(m.playerMsgs[name], &schema.Message{Role: schema.User, Content: content})
	}
}
