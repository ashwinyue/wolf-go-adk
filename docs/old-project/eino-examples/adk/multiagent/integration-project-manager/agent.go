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

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino-examples/adk/common/prints"
	"github.com/cloudwego/eino-examples/adk/multiagent/integration-project-manager/agents"
)

func main() {
	ctx := context.Background()

	// 为智能体初始化聊天模型
	tcm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		ByAzure: func() bool {
			return os.Getenv("OPENAI_BY_AZURE") == "true"
		}(),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 初始化研究智能体
	researchAgent, err := agents.NewResearchAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化代码智能体
	codeAgent, err := agents.NewCodeAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化技术审查智能体
	reviewAgent, err := agents.NewReviewAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化项目经理智能体
	s, err := agents.NewProjectManagerAgent(ctx, tcm)
	if err != nil {
		log.Fatal(err)
	}

	// 组合成supervisor模式的智能体
	// Supervisor: 项目经理
	// Sub-agents: 研究员 / 编码员 / 审查员
	supervisorAgent, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: s,
		SubAgents:  []adk.Agent{researchAgent, codeAgent, reviewAgent},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 配置运行器
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           supervisorAgent,
		EnableStreaming: true,
		// 你可以在这里禁用流式传输
		CheckPointStore: newInMemoryStore(),
	})

	// 替换为你自己的查询
	// 当使用以下查询时，researchAgent将中断并提示用户通过stdin输入具体的研究主题。
	query := "请给我一份关于的优势报告"
	checkpointID := "1"

	// researchAgent可能需要用户多次输入信息
	// 因此，使用以下标志"interrupted"和"finished"来支持多次中断和恢复。
	interrupted := false
	finished := false

	for !finished {
		var iter *adk.AsyncIterator[*adk.AgentEvent]

		if !interrupted {
			iter = runner.Query(ctx, query, adk.WithCheckPointID(checkpointID))
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("\n请输入网络搜索的额外上下文: ")
			scanner.Scan()
			fmt.Println()
			nInput := scanner.Text()

			iter, err = runner.Resume(ctx, checkpointID, adk.WithToolOptions([]tool.Option{agents.WithNewInput(nInput)}))
			if err != nil {
				log.Fatal(err)
			}
		}

		interrupted = false

		for {
			event, ok := iter.Next()
			if !ok {
				if !interrupted {
					finished = true
				}
				break
			}
			if event.Err != nil {
				log.Fatal(event.Err)
			}
			if event.Action != nil {
				if event.Action.Interrupted != nil {
					interrupted = true
				}
				if event.Action.Exit {
					finished = true
				}
			}
			prints.Event(event)
		}
	}
}

func newInMemoryStore() compose.CheckPointStore {
	return &inMemoryStore{
		mem: map[string][]byte{},
	}
}

type inMemoryStore struct {
	mem map[string][]byte
}

func (i *inMemoryStore) Set(ctx context.Context, key string, value []byte) error {
	i.mem[key] = value
	return nil
}

func (i *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	v, ok := i.mem[key]
	return v, ok, nil
}
