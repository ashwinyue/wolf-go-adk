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

package memory

import (
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	// CollectionName Milvus collection 名称
	CollectionName = "wolf_episodes"
	// VectorDim Ark embedding 向量维度（doubao-embedding 默认 2560）
	VectorDim = 2560
)

// GetEinoFields 获取符合 eino 格式的 Fields（使用 FloatVector）
func GetEinoFields() []*entity.Field {
	return []*entity.Field{
		entity.NewField().
			WithName("id").
			WithDescription("the unique id of the document").
			WithIsPrimaryKey(true).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(255),
		entity.NewField().
			WithName("vector").
			WithDescription("the vector of the document").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(VectorDim),
		entity.NewField().
			WithName("content").
			WithDescription("the content of the document").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(4096),
		entity.NewField().
			WithName("metadata").
			WithDescription("the metadata of the document").
			WithIsPrimaryKey(false).
			WithDataType(entity.FieldTypeJSON),
	}
}
