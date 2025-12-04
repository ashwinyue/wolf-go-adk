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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/cloudwego/eino/adk"

	"github.com/ashwinyue/wolf-go-adk/agents/supervisor"
	"github.com/ashwinyue/wolf-go-adk/params"
	"github.com/cloudwego/eino-examples/adk/common/prints"
	"github.com/cloudwego/eino-examples/adk/common/trace"
)

func main() {
	// 命令行参数
	humanPlayer := flag.String("human", "", "人类玩家名字 (Player1-Player9)，为空则全部 AI")
	flag.Parse()

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 语言设置：GAME_LANG=zh 使用中文，默认英文
	if strings.ToLower(os.Getenv("GAME_LANG")) == "zh" {
		params.UseChinese()
		log.Println("使用中文模式")
	}

	ctx := context.Background()

	// 初始化追踪（可选）
	traceCloseFn, startSpanFn := trace.AppendCozeLoopCallbackIfConfigured(ctx)
	defer traceCloseFn(ctx)

	// 创建主持人 Agent（Supervisor 模式）
	var moderator *supervisor.ModeratorAgent
	var err error
	if *humanPlayer != "" {
		fmt.Printf("🎮 人类玩家模式: %s\n", *humanPlayer)
		moderator, err = supervisor.NewModeratorAgentWithHuman(ctx, *humanPlayer)
	} else {
		fmt.Println("🤖 全 AI 模式")
		moderator, err = supervisor.NewModeratorAgent(ctx)
	}
	if err != nil {
		log.Fatalf("创建主持人 Agent 失败: %v", err)
	}

	query := "开始一局狼人杀游戏"
	ctx, endSpanFn := startSpanFn(ctx, "werewolf-game", query)

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           moderator,
	})

	// 运行游戏
	iter := runner.Query(ctx, query)

	var lastMessage adk.Message

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			fmt.Printf("错误: %v\n", event.Err)
			break
		}

		prints.Event(event)
		if event.Output != nil {
			lastMessage, _, _ = adk.GetMessage(event)
		}
	}

	endSpanFn(ctx, lastMessage)
}
