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
	"log"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/eino-examples/adk/common/prints"
	"github.com/cloudwego/eino-examples/adk/intro/transfer/subagents"
)

func main() {
	weatherAgent := subagents.NewWeatherAgent()
	chatAgent := subagents.NewChatAgent()
	routerAgent := subagents.NewRouterAgent()

	ctx := context.Background()
	a, err := adk.SetSubAgents(ctx, routerAgent, []adk.Agent{chatAgent, weatherAgent})
	if err != nil {
		log.Fatal(err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true, // 你可以在这里禁用流式输出
		Agent:           a,
	})

	// 查询天气
	println("\n\n>>>>>>>>>查询天气<<<<<<<<<")
	iter := runner.Query(ctx, "北京天气怎么样？")
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(event.Err)
		}

		prints.Event(event)
	}

	// 路由失败
	println("\n\n>>>>>>>>>路由失败<<<<<<<<<")
	iter = runner.Query(ctx, "帮我预订明天从纽约到伦敦的航班。")
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(event.Err)
		}
		prints.Event(event)
	}
}
