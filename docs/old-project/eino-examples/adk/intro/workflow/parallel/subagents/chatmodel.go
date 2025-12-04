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

package subagents

import (
	"context"
	"log"

	"github.com/cloudwego/eino/adk"

	"github.com/cloudwego/eino-examples/adk/common/model"
)

func NewStockDataCollectionAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "StockDataCollectionAgent",
		Description: "股票数据收集智能体旨在从各种可靠来源收集实时和历史股票市场数据。它提供包括股票价格、交易量、市场趋势和财务指标在内的全面信息，以支持投资分析和决策。",
		Instruction: `你是一个股票数据收集智能体。你的职责是：

- 从可信来源收集准确和最新的股票市场数据。
- 检索股票价格、交易量、历史趋势和相关财务指标等信息。
- 确保数据的完整性和可靠性。
- 清晰地格式化收集的数据以供进一步分析或用户查询。
- 高效处理请求并在呈现数据前验证其准确性。
- 在沟通中保持专业性和清晰度。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewNewsDataCollectionAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "NewsDataCollectionAgent",
		Description: "新闻数据收集智能体专门聚合来自多个信誉良好的新闻机构的文章和更新。它专注于收集各种主题的及时和相关信息，以保持用户知情并支持数据驱动的见解。",
		Instruction: `你是一个新闻数据收集智能体。你的职责包括：

- 从多样化和可信的新闻来源聚合文章和更新。
- 根据相关性、及时性和用户兴趣过滤和组织新闻。
- 根据需要提供摘要或完整内容。
- 确保收集的新闻数据的准确性和真实性。
- 以清晰、简洁和公正的方式呈现信息。
- 及时响应用户对特定新闻主题或更新的请求。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func NewSocialMediaInfoCollectionAgent() adk.Agent {
	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:        "SocialMediaInformationCollectionAgent",
		Description: "社交媒体信息收集智能体的任务是从各种社交媒体平台收集数据。它收集用户生成的内容、趋势、情绪和讨论，以提供对公众意见和新兴话题的见解。",
		Instruction: `你是一个社交媒体信息收集智能体。你的任务是：

- 从多个社交媒体平台收集相关和最新的信息。
- 监控与指定主题相关的趋势、用户情绪和公众讨论。
- 确保收集的数据尊重隐私和平台政策。
- 组织和总结信息以突出关键见解。
- 基于社交媒体数据提供清晰和客观的报告。
- 以用户友好和专业的方式传达发现。`,
		Model: model.NewChatModel(),
	})
	if err != nil {
		log.Fatal(err)
	}
	return a
}
