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

package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// MustNewChatModel 创建聊天模型，失败时 panic
func MustNewChatModel(ctx context.Context) model.ToolCallingChatModel {
	cm, err := NewChatModel(ctx)
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}
	return cm
}

// NewChatModel 创建聊天模型
// 支持多种模型后端，通过 MODEL_TYPE 环境变量切换：
//   - "dashscope": 使用阿里云 DashScope (DASHSCOPE_API_KEY, MODEL_NAME)
//   - 默认: 使用 OpenAI 兼容接口 (OPENAI_API_KEY, OPENAI_MODEL, OPENAI_BASE_URL)
//
// 与 eino-examples 最佳实践一致，支持通过环境变量灵活配置模型
func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	modelType := strings.ToLower(os.Getenv("MODEL_TYPE"))

	switch modelType {
	case "dashscope":
		// 阿里云 DashScope (通过 OpenAI 兼容接口)
		modelName := os.Getenv("MODEL_NAME")
		if modelName == "" {
			modelName = "qwen-max"
		}
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			APIKey:  os.Getenv("DASHSCOPE_API_KEY"),
			Model:   modelName,
		})

	default:
		// OpenAI 兼容接口（默认）
		apiKey := os.Getenv("OPENAI_API_KEY")
		modelName := os.Getenv("OPENAI_MODEL")
		baseURL := os.Getenv("OPENAI_BASE_URL")

		// 兼容旧的 DASHSCOPE 环境变量（向后兼容）
		if apiKey == "" && os.Getenv("DASHSCOPE_API_KEY") != "" {
			apiKey = os.Getenv("DASHSCOPE_API_KEY")
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
			modelName = os.Getenv("MODEL_NAME")
			if modelName == "" {
				modelName = "qwen-max"
			}
		}

		if apiKey == "" {
			return nil, fmt.Errorf("no API key configured, set OPENAI_API_KEY or DASHSCOPE_API_KEY")
		}

		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL: baseURL,
			APIKey:  apiKey,
			Model:   modelName,
			ByAzure: os.Getenv("OPENAI_BY_AZURE") == "true",
		})
	}
}
