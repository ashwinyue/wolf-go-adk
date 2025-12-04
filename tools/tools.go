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
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"github.com/ashwinyue/wolf-go-adk/game"
)

// ========== 狼人工具 ==========

// DiscussInput 狼人讨论输入
type DiscussInput struct {
	Message        string `json:"message" jsonschema:"description=你想对其他狼人说的话"`
	ReachAgreement bool   `json:"reach_agreement" jsonschema:"description=是否已达成一致意见"`
}

// DiscussOutput 狼人讨论输出
type DiscussOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewDiscussTool 创建狼人讨论工具
func NewDiscussTool() tool.BaseTool {
	fn := func(ctx context.Context, input *DiscussInput) (*DiscussOutput, error) {
		return &DiscussOutput{
			Success: true,
			Message: input.Message,
		}, nil
	}

	t, err := utils.InferTool("discuss", "狼人内部讨论工具，用于与其他狼人交流并决定击杀目标", fn)
	if err != nil {
		panic(fmt.Errorf("create discuss tool failed: %w", err))
	}
	return t
}

// KillInput 狼人击杀输入
type KillInput struct {
	Target string `json:"target" jsonschema:"description=要击杀的玩家名"`
}

// KillOutput 狼人击杀输出
type KillOutput struct {
	Success bool   `json:"success"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// NewKillTool 创建狼人击杀工具
func NewKillTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *KillInput) (*KillOutput, error) {
		if !state.IsAlive(input.Target) {
			return &KillOutput{
				Success: false,
				Message: fmt.Sprintf("目标 %s 已死亡", input.Target),
			}, nil
		}
		if state.GetPlayerRole(input.Target) == game.RoleWerewolf {
			return &KillOutput{
				Success: false,
				Message: "不能击杀同伴狼人",
			}, nil
		}

		state.SetNightKilled(input.Target)
		return &KillOutput{
			Success: true,
			Target:  input.Target,
			Message: fmt.Sprintf("决定击杀 %s", input.Target),
		}, nil
	}

	t, err := utils.InferTool("kill", "狼人击杀工具，用于选择今晚要击杀的玩家", fn)
	if err != nil {
		panic(fmt.Errorf("create kill tool failed: %w", err))
	}
	return t
}

// ========== 预言家工具 ==========

// CheckInput 预言家查验输入
type CheckInput struct {
	Target string `json:"target" jsonschema:"description=要查验的玩家名"`
}

// CheckOutput 预言家查验输出
type CheckOutput struct {
	Target  string `json:"target"`
	IsWolf  bool   `json:"is_wolf"`
	Message string `json:"message"`
}

// NewCheckTool 创建预言家查验工具
func NewCheckTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *CheckInput) (*CheckOutput, error) {
		if !state.IsAlive(input.Target) {
			return &CheckOutput{
				Target:  input.Target,
				Message: fmt.Sprintf("目标 %s 已死亡，无法查验", input.Target),
			}, nil
		}

		role := state.GetPlayerRole(input.Target)
		isWolf := role == game.RoleWerewolf

		result := "好人"
		if isWolf {
			result = "狼人"
		}

		return &CheckOutput{
			Target:  input.Target,
			IsWolf:  isWolf,
			Message: fmt.Sprintf("查验结果：%s 是 %s", input.Target, result),
		}, nil
	}

	t, err := utils.InferTool("check_identity", "预言家查验工具，用于查验一名玩家的身份", fn)
	if err != nil {
		panic(fmt.Errorf("create check tool failed: %w", err))
	}
	return t
}

// ========== 女巫工具 ==========

// SaveInput 女巫救人输入
type SaveInput struct {
	Save bool `json:"save" jsonschema:"description=是否使用解药救人"`
}

// SaveOutput 女巫救人输出
type SaveOutput struct {
	Success bool   `json:"success"`
	Saved   string `json:"saved"`
	Message string `json:"message"`
}

// NewSaveTool 创建女巫救人工具
func NewSaveTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *SaveInput) (*SaveOutput, error) {
		if !state.CanUseHealingPotion() {
			return &SaveOutput{
				Success: false,
				Message: "解药已用完",
			}, nil
		}

		killed := state.GetNightKilled()
		if killed == "" {
			return &SaveOutput{
				Success: false,
				Message: "今晚没有人被狼人击杀",
			}, nil
		}

		// 女巫不能自救
		if killed == state.Witch {
			return &SaveOutput{
				Success: false,
				Message: "女巫不能自救",
			}, nil
		}

		if input.Save {
			state.SetNightSaved(true)
			return &SaveOutput{
				Success: true,
				Saved:   killed,
				Message: fmt.Sprintf("使用解药救活了 %s", killed),
			}, nil
		}

		return &SaveOutput{
			Success: true,
			Message: "选择不使用解药",
		}, nil
	}

	t, err := utils.InferTool("save", "女巫救人工具，用于使用解药救活被狼人击杀的玩家", fn)
	if err != nil {
		panic(fmt.Errorf("create save tool failed: %w", err))
	}
	return t
}

// PoisonInput 女巫毒人输入
type PoisonInput struct {
	Poison bool   `json:"poison" jsonschema:"description=是否使用毒药"`
	Target string `json:"target" jsonschema:"description=要毒杀的玩家名（如果使用毒药）"`
}

// PoisonOutput 女巫毒人输出
type PoisonOutput struct {
	Success  bool   `json:"success"`
	Poisoned string `json:"poisoned"`
	Message  string `json:"message"`
}

// NewPoisonTool 创建女巫毒人工具
func NewPoisonTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *PoisonInput) (*PoisonOutput, error) {
		if !state.CanUsePoisonPotion() {
			return &PoisonOutput{
				Success: false,
				Message: "毒药已用完",
			}, nil
		}

		if !input.Poison {
			return &PoisonOutput{
				Success: true,
				Message: "选择不使用毒药",
			}, nil
		}

		if input.Target == "" {
			return &PoisonOutput{
				Success: false,
				Message: "请指定毒杀目标",
			}, nil
		}

		if !state.IsAlive(input.Target) {
			return &PoisonOutput{
				Success: false,
				Message: fmt.Sprintf("目标 %s 已死亡", input.Target),
			}, nil
		}

		// 女巫不能毒自己
		if input.Target == state.Witch {
			return &PoisonOutput{
				Success: false,
				Message: "女巫不能毒自己",
			}, nil
		}

		state.SetNightPoisoned(input.Target)
		return &PoisonOutput{
			Success:  true,
			Poisoned: input.Target,
			Message:  fmt.Sprintf("使用毒药毒杀了 %s", input.Target),
		}, nil
	}

	t, err := utils.InferTool("poison", "女巫毒人工具，用于使用毒药毒杀一名玩家", fn)
	if err != nil {
		panic(fmt.Errorf("create poison tool failed: %w", err))
	}
	return t
}

// ========== 猎人工具 ==========

// ShootInput 猎人开枪输入
type ShootInput struct {
	Shoot  bool   `json:"shoot" jsonschema:"description=是否开枪"`
	Target string `json:"target" jsonschema:"description=要射杀的玩家名（如果开枪）"`
}

// ShootOutput 猎人开枪输出
type ShootOutput struct {
	Success bool   `json:"success"`
	Shot    string `json:"shot"`
	Message string `json:"message"`
}

// NewShootTool 创建猎人开枪工具
func NewShootTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *ShootInput) (*ShootOutput, error) {
		if !input.Shoot {
			return &ShootOutput{
				Success: true,
				Message: "选择不开枪",
			}, nil
		}

		if input.Target == "" {
			return &ShootOutput{
				Success: false,
				Message: "请指定射杀目标",
			}, nil
		}

		if !state.IsAlive(input.Target) {
			return &ShootOutput{
				Success: false,
				Message: fmt.Sprintf("目标 %s 已死亡", input.Target),
			}, nil
		}

		state.KillPlayer(input.Target)
		return &ShootOutput{
			Success: true,
			Shot:    input.Target,
			Message: fmt.Sprintf("猎人开枪射杀了 %s", input.Target),
		}, nil
	}

	t, err := utils.InferTool("shoot", "猎人开枪工具，被淘汰时可以开枪带走一名玩家", fn)
	if err != nil {
		panic(fmt.Errorf("create shoot tool failed: %w", err))
	}
	return t
}

// ========== 投票工具 ==========

// VoteInput 投票输入
type VoteInput struct {
	Target string `json:"target" jsonschema:"description=投票淘汰的玩家名"`
}

// VoteOutput 投票输出
type VoteOutput struct {
	Success bool   `json:"success"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// NewVoteTool 创建投票工具
func NewVoteTool(state *game.GameState) tool.BaseTool {
	fn := func(ctx context.Context, input *VoteInput) (*VoteOutput, error) {
		if !state.IsAlive(input.Target) {
			return &VoteOutput{
				Success: false,
				Message: fmt.Sprintf("目标 %s 已死亡，无法投票", input.Target),
			}, nil
		}

		return &VoteOutput{
			Success: true,
			Target:  input.Target,
			Message: fmt.Sprintf("投票淘汰 %s", input.Target),
		}, nil
	}

	t, err := utils.InferTool("vote", "投票工具，用于在白天投票淘汰玩家", fn)
	if err != nil {
		panic(fmt.Errorf("create vote tool failed: %w", err))
	}
	return t
}
