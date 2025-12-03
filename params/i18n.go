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

package params

// I18nStrings 国际化字符串
type I18nStrings struct {
	// 游戏流程
	GameStarted    string
	Players        string
	RoleAssignment string
	Round          string
	NightPhase     string
	DayPhase       string
	GameEnded      string

	// 狼人
	WerewolvesDiscussing  string
	WerewolfRound         string
	WerewolvesAgreed      string
	WerewolvesNoAgreement string
	WerewolvesVoteTo      string
	WerewolvesDecided     string

	// 女巫
	WitchDeciding string
	WitchSaved    string
	WitchPoisoned string
	WitchNoAction string

	// 预言家
	SeerChecking string
	SeerResult   string

	// 夜间结算
	NightSummary  string
	PeacefulNight string
	NightDeaths   string

	// 白天
	AlivePlayers    string
	DiscussionPhase string
	PlayerSpeaks    string
	VotingPhase     string
	PlayerVotes     string
	NoValidVotes    string
	VoteResult      string

	// 遗言
	LastWords string

	// 猎人
	HunterShot string

	// 胜利
	WerewolvesWin string
	VillagersWin  string
	Roles         string

	// 反思
	PlayerReflections string
	Reflection        string

	// 错误
	Error string
}

// I18n 当前使用的国际化字符串
var I18n = ChineseI18n

// ChineseI18n 中文国际化
var ChineseI18n = I18nStrings{
	GameStarted:    "=== 🐺 狼人杀游戏开始 🐺 ===",
	Players:        "玩家",
	RoleAssignment: "=== 角色分配 ===",
	Round:          "========== 第 %d 回合 ==========",
	NightPhase:     "--- 🌙 夜晚阶段 ---",
	DayPhase:       "--- ☀️ 白天阶段 ---",
	GameEnded:      "游戏结束：达到最大回合数",

	WerewolvesDiscussing:  "狼人 (%s) 正在讨论...",
	WerewolfRound:         "[%s] (狼人第 %d 轮): %s",
	WerewolvesAgreed:      "✓ 狼人达成一致！",
	WerewolvesNoAgreement: "⚠️ 狼人未达成明确一致，进入投票...",
	WerewolvesVoteTo:      "[%s] 投票杀: %s",
	WerewolvesDecided:     "➡️ 狼人决定杀: %s (%s)",

	WitchDeciding: "女巫 (%s) 正在决定...",
	WitchSaved:    "➡️ 女巫救了 %s！",
	WitchPoisoned: "➡️ 女巫毒了 %s！",
	WitchNoAction: "女巫选择不使用药水",

	SeerChecking: "预言家 (%s) 正在查验...",
	SeerResult:   "➡️ 预言家查验 %s: %s",

	NightSummary:  "夜晚结算",
	PeacefulNight: "✨ 平安夜，无人死亡。",
	NightDeaths:   "☠️ 夜晚结算，死亡: %s",

	AlivePlayers:    "📢 存活玩家: %s",
	DiscussionPhase: "💬 讨论阶段:",
	PlayerSpeaks:    "[%s]: %s",
	VotingPhase:     "🗳️ 投票阶段:",
	PlayerVotes:     "[%s] 投票: %s",
	NoValidVotes:    "无有效投票。",
	VoteResult:      "➡️ 投票结果: %s 被淘汰 (%s)",

	LastWords: "[%s] (遗言): %s",

	HunterShot: "🔫 猎人射杀了 %s！",

	WerewolvesWin: "🐺🐺🐺 狼人获胜！🐺🐺🐺",
	VillagersWin:  "🏘️🏘️🏘️ 村民获胜！🏘️🏘️🏘️",
	Roles:         "角色",

	PlayerReflections: "=== 玩家反思 ===",
	Reflection:        "[%s] 反思: %s",

	Error: "[%s] 错误: %v",
}

// EnglishI18n 英文国际化
var EnglishI18n = I18nStrings{
	GameStarted:    "=== 🐺 Werewolf Game Started 🐺 ===",
	Players:        "Players",
	RoleAssignment: "=== Role Assignment ===",
	Round:          "========== Round %d ==========",
	NightPhase:     "--- 🌙 Night Phase ---",
	DayPhase:       "--- ☀️ Day Phase ---",
	GameEnded:      "Game ended: Maximum rounds reached",

	WerewolvesDiscussing:  "Werewolves (%s) are discussing...",
	WerewolfRound:         "[%s] (Wolf round %d): %s",
	WerewolvesAgreed:      "✓ Werewolves reached agreement!",
	WerewolvesNoAgreement: "⚠️ Werewolves did not reach clear agreement, proceeding to vote...",
	WerewolvesVoteTo:      "[%s] votes to kill: %s",
	WerewolvesDecided:     "➡️ Werewolves decided to kill: %s (%s)",

	WitchDeciding: "Witch (%s) is deciding...",
	WitchSaved:    "➡️ Witch saved %s!",
	WitchPoisoned: "➡️ Witch poisoned %s!",
	WitchNoAction: "Witch chose not to use potions",

	SeerChecking: "Seer (%s) is checking...",
	SeerResult:   "➡️ Seer checked %s: %s",

	NightSummary:  "Night Summary",
	PeacefulNight: "✨ Peaceful night, no one died.",
	NightDeaths:   "☠️ Night summary, deaths: %s",

	AlivePlayers:    "📢 Alive players: %s",
	DiscussionPhase: "💬 Discussion phase:",
	PlayerSpeaks:    "[%s]: %s",
	VotingPhase:     "🗳️ Voting phase:",
	PlayerVotes:     "[%s] votes: %s",
	NoValidVotes:    "No valid votes.",
	VoteResult:      "➡️ Vote result: %s eliminated (%s)",

	LastWords: "[%s] (Last words): %s",

	HunterShot: "🔫 Hunter shot %s!",

	WerewolvesWin: "🐺🐺🐺 Werewolves Win! 🐺🐺🐺",
	VillagersWin:  "🏘️🏘️🏘️ Villagers Win! 🏘️🏘️🏘️",
	Roles:         "Roles",

	PlayerReflections: "=== Player Reflections ===",
	Reflection:        "[%s] Reflection: %s",

	Error: "[%s] Error: %v",
}

// UseChinese 切换到中文模式
func UseChinese() {
	Prompts = ChinesePrompts
	I18n = ChineseI18n
}

// UseEnglish 使用英文
func UseEnglish() {
	I18n = EnglishI18n
}
