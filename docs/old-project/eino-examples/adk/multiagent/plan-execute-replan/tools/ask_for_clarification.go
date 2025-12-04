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

package tools

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type AskForClarificationInput struct {
	Question string `json:"question" jsonschema_description:"你想向用户询问以获取缺失信息的具体问题"`
}

func NewAskForClarificationTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"ask_for_clarification",
		"当用户的请求模糊不清或缺少必要信息以继续进行时，调用此工具。用它来询问后续问题，以获取你需要的详细信息，例如书籍的类型，然后才能有效地使用其他工具。",
		func(ctx context.Context, input *AskForClarificationInput, opts ...tool.Option) (output string, err error) {
			fmt.Printf("\n问题: %s\n", input.Question)
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("\n请在此输入: ")
			scanner.Scan()
			fmt.Println()
			nInput := scanner.Text()
			return nInput, nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}
