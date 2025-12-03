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

import (
	"fmt"

	"github.com/cloudwego/eino-examples/week11-homework/werewolves-adk/game"
)

// PromptsTemplate 游戏提示词模板
type PromptsTemplate struct {
	BaseSystem string

	// 死亡相关
	ToDeadPlayer string

	// 游戏开始
	ToAllNewGame string

	// 夜晚阶段
	ToAllNight string

	// 狼人相关
	ToWolvesDiscussion string
	ToWolvesVote       string
	ToWolvesRes        string

	// 女巫相关
	ToAllWitchTurn      string
	ToWitchResurrect    string
	ToWitchResurrectNo  string
	ToWitchResurrectYes string
	ToWitchPoison       string

	// 预言家相关
	ToAllSeerTurn string
	ToSeer        string
	ToSeerResult  string

	// 猎人相关
	ToHunter         string
	ToAllHunterShoot string

	// 白天阶段
	ToAllDay     string
	ToAllPeace   string
	ToAllDiscuss string
	ToAllVote    string
	ToAllRes     string

	// 游戏结束
	ToAllWolfWin    string
	ToAllVillageWin string
	ToAllContinue   string
	ToAllReflect    string
}

// Prompts 当前使用的提示词模板
var Prompts = ChinesePrompts

// ChinesePrompts 中文游戏提示词模板
var ChinesePrompts = PromptsTemplate{
	BaseSystem: `你是一个狼人杀游戏玩家，名字是 %s。

# 你的目标
尽可能与队友一起赢得游戏。

# 游戏规则
- 狼人杀游戏中，玩家分为三个狼人、三个村民、一个预言家、一个猎人和一个女巫。
    - 狼人：每晚杀死一名玩家，白天必须隐藏身份。
    - 村民：没有特殊能力的普通玩家，尝试识别并淘汰狼人。
        - 预言家：特殊村民，每晚可以查验一名玩家的身份。
        - 女巫：特殊村民，有两瓶一次性药水：解药可以救活被狼人杀死的玩家，毒药可以毒死一名玩家。
        - 猎人：特殊村民，被淘汰时可以带走一名玩家。
- 游戏在夜晚和白天阶段交替进行，直到一方获胜：
    - 夜晚阶段：狼人选择一名受害者，预言家查验一名玩家身份，女巫决定是否使用药水
    - 白天阶段：所有玩家讨论并投票淘汰一名嫌疑玩家

# 游戏指导
- 尽你所能与队友一起赢得游戏，欺骗、谎言和伪装都是允许的。
- 讨论时不要政治化，直接切入主题。
- 白天投票阶段提供重要线索。

# 你的角色
你是 %s。
%s

# 注意
- [重要] 不要编造主持人或其他玩家未提供的信息。
- 这是一个文字游戏，不要使用或编造任何非文字信息。
- 始终批判性地反思你的证据是否存在，避免做出假设。
- 你的回复应该具体简洁，提供清晰的理由，避免不必要的阐述。
- 生成一行回复。
- 不要重复其他人的发言。`,

	// 死亡相关
	ToDeadPlayer: "%s, 你已被淘汰。现在你可以向所有存活玩家发表最后的遗言。",

	// 游戏开始
	ToAllNewGame: "新的一局游戏开始，参与玩家包括：%s。现在为每位玩家重新随机分配身份，并私下告知各自身份。",

	// 夜晚阶段
	ToAllNight: "天黑了，请所有人闭眼。狼人请睁眼，选择今晚要淘汰的一名玩家...",

	// 狼人相关
	ToWolvesDiscussion: `[仅狼人可见] %s, 你们需要讨论并决定今晚要淘汰的玩家。当前存活玩家有：%s。

讨论要点：
1. 分析哪些玩家可能是特殊角色（预言家、女巫、猎人）
2. 考虑选择中置位玩家以降低怀疑
3. 提出你的建议和具体理由
4. 如果同意队友的建议，说明原因并补充策略

如果达成一致，请将 reach_agreement 设为 True。`,
	ToWolvesVote: "[仅狼人可见] 你投票要杀死哪位玩家？",
	ToWolvesRes:  "[仅狼人可见] 投票结果为 %s，你们选择淘汰 %s。",

	// 女巫相关
	ToAllWitchTurn:      "轮到女巫行动，女巫请睁眼并决定今晚的操作...",
	ToWitchResurrect:    "[仅女巫可见] %s，你是女巫，今晚%s被淘汰。你可以用解药救他/她，注意解药全局只能用一次。你要救%s吗？请给出理由和决定。",
	ToWitchResurrectNo:  "[仅女巫可见] 女巫选择不救该玩家。",
	ToWitchResurrectYes: "[仅女巫可见] 女巫选择救活该玩家。",
	ToWitchPoison:       "[仅女巫可见] %s，你有一瓶一次性毒药，今晚要使用吗？请给出理由和决定。",

	// 预言家相关
	ToAllSeerTurn: "轮到预言家行动，预言家请睁眼并查验一名玩家身份...",
	ToSeer:        "[仅预言家可见] %s, 你是预言家，今晚可以查验一名玩家身份。你要查谁？请给出理由和决定。",
	ToSeerResult:  "[仅预言家可见] 你查验了%s，结果是：%s。",

	// 猎人相关
	ToHunter:         "[仅猎人可见] %s，你是猎人，今晚被淘汰。你可以选择带走一名玩家，也可以选择不带走。请给出理由和决定。",
	ToAllHunterShoot: "猎人选择带走 %s 一起出局。",

	// 白天阶段
	ToAllDay:     "天亮了，请所有玩家睁眼。昨晚被淘汰的玩家有：%s。",
	ToAllPeace:   "天亮了，请所有玩家睁眼。昨晚平安夜，无人被淘汰。",
	ToAllDiscuss: "现在存活玩家有：%s。游戏继续，大家开始讨论并投票淘汰一名玩家。请按顺序（%s）依次发言。",
	ToAllVote:    "讨论结束。请大家从存活玩家中投票淘汰一人：%s。",
	ToAllRes:     "投票结果为 %s，%s 被淘汰。",

	// 游戏结束
	ToAllWolfWin:    "当前存活玩家共%d人，其中%d人为狼人。游戏结束，狼人获胜🐺🎉！本局所有玩家真实身份为：%s",
	ToAllVillageWin: "所有狼人已被淘汰。游戏结束，村民获胜🏘️🎉！本局所有玩家真实身份为：%s",
	ToAllContinue:   "游戏继续。",
	ToAllReflect:    "游戏结束。现在每位玩家可以对自己的表现进行反思。注意每位玩家只有一次发言机会，且反思内容仅自己可见。",
}

// EnglishPrompts 英文游戏提示词模板
var EnglishPrompts = PromptsTemplate{
	BaseSystem: `You're a werewolf game player named %s.

# YOUR TARGET
Your target is to win the game with your teammates as much as possible.

# GAME RULES
- In werewolf game, players are divided into three werewolves, three villagers, one seer, one hunter and one witch.
    - Werewolves: kill one player each night, and must hide identity during the day.
    - Villagers: ordinary players without special abilities, try to identify and eliminate werewolves.
        - Seer: A special villager who can check one player's identity each night.
        - Witch: A special villager with two one-time-use potions: a healing potion to save a player from being killed at night, and a poison to eliminate one player at night.
        - Hunter: A special villager who can take one player down with them when they are eliminated.
- The game alternates between night and day phases until one side wins:
    - Night Phase: Werewolves choose one victim, Seer checks one player's identity, Witch decides whether to use potions
    - Day Phase: All players discuss and vote to eliminate one suspected player

# GAME GUIDANCE
- Try your best to win the game with your teammates, tricks, lies, and deception are all allowed.
- During discussion, don't be political, be direct and to the point.
- The day phase voting provides important clues.

# YOUR ROLE
You are a %s.
%s

# NOTE
- [IMPORTANT] DO NOT make up any information that is not provided by the moderator or other players.
- This is a TEXT-based game, so DO NOT use or make up any non-textual information.
- Always critically reflect on whether your evidence exist, and avoid making assumptions.
- Your response should be specific and concise, provide clear reason and avoid unnecessary elaboration.
- Generate a one-line response.
- Don't repeat the others' speeches.`,

	// 死亡相关
	ToDeadPlayer: "%s, you're eliminated now. Now you can make a final statement to all alive players before you leave the game.",

	// 游戏开始
	ToAllNewGame: "A new game is starting, the players are: %s. Now we randomly reassign the roles to each player and inform them of their roles privately.",

	// 夜晚阶段
	ToAllNight: "Night has fallen, everyone close your eyes. Werewolves open your eyes and choose a player to eliminate tonight.",

	// 狼人相关
	ToWolvesDiscussion: `[WEREWOLVES ONLY] %s, you need to discuss and decide on a player to eliminate tonight. Current alive players are %s.

Discussion points:
1. Analyze which players might be special roles (Seer, Witch, Hunter)
2. Consider choosing mid-position players to reduce suspicion
3. Propose your suggestion with specific reasons
4. If you agree with teammates, explain why and add strategy tips

Set reach_agreement to True when you reach consensus.`,
	ToWolvesVote: "[WEREWOLVES ONLY] Which player do you vote to kill?",
	ToWolvesRes:  "[WEREWOLVES ONLY] The voting result is %s. So you have chosen to eliminate %s.",

	// 女巫相关
	ToAllWitchTurn:      "Witch's turn, witch open your eyes and decide your action tonight...",
	ToWitchResurrect:    "[WITCH ONLY] %s, you're the witch, and tonight %s is eliminated. You can resurrect him/her by using your healing potion, and note you can only use it once in the whole game. Do you want to resurrect %s? Give me your reason and decision.",
	ToWitchResurrectNo:  "[WITCH ONLY] The witch has chosen not to resurrect the player.",
	ToWitchResurrectYes: "[WITCH ONLY] The witch has chosen to resurrect the player.",
	ToWitchPoison:       "[WITCH ONLY] %s, as a witch, you have a one-time-use poison potion, do you want to use it tonight? Give me your reason and decision.",

	// 预言家相关
	ToAllSeerTurn: "Seer's turn, seer open your eyes and check one player's identity tonight...",
	ToSeer:        "[SEER ONLY] %s, as the seer you can check one player's identity tonight. Who do you want to check? Give me your reason and decision.",
	ToSeerResult:  "[SEER ONLY] You've checked %s, and the result is: %s.",

	// 猎人相关
	ToHunter:         "[HUNTER ONLY] %s, as the hunter you're eliminated tonight. You can choose one player to take down with you. Also, you can choose not to use this ability. Give me your reason and decision.",
	ToAllHunterShoot: "The hunter has chosen to shoot %s down with him/herself.",

	// 白天阶段
	ToAllDay:     "The day is coming, all players open your eyes. Last night, the following player(s) has been eliminated: %s.",
	ToAllPeace:   "The day is coming, all the players open your eyes. Last night is peaceful, no player is eliminated.",
	ToAllDiscuss: "Now the alive players are %s. The game goes on, it's time to discuss and vote a player to be eliminated. Now you each take turns to speak once in the order of %s.",
	ToAllVote:    "Now the discussion is over. Everyone, please vote to eliminate one player from the alive players: %s.",
	ToAllRes:     "The voting result is %s. So %s has been voted out.",

	// 游戏结束
	ToAllWolfWin:    "There are %d players alive, and %d of them are werewolves. The game is over and werewolves win🐺🎉!In this game, the true roles of all players are: %s",
	ToAllVillageWin: "All the werewolves have been eliminated.The game is over and villagers win🏘️🎉!In this game, the true roles of all players are: %s",
	ToAllContinue:   "The game goes on.",
	ToAllReflect:    "The game is over. Now each player can reflect on their performance. Note each player only has one chance to speak and the reflection is only visible to themselves.",
}

// RoleGuidance 角色指导
var RoleGuidance = map[game.Role]string{
	game.RoleWerewolf: `## 狼人游戏指导
- 预言家是你最大的威胁，他每晚可以查验一名玩家的身份。分析玩家发言，找出预言家并淘汰他将大大增加你获胜的机会。
- 第一晚，由于没有信息，狼人随机选择是常见的。
- 假装成其他角色（预言家、女巫或村民）是隐藏身份和在白天误导其他村民的常见策略。
- 夜晚阶段的结果提供重要线索。例如，女巫是否使用了解药或毒药，死者是否是猎人等。利用这些信息调整你的策略。`,

	game.RoleVillager: `## 村民游戏指导
- 保护特殊村民，尤其是预言家，对你方的成功至关重要。
- 狼人可能假装成预言家。保持警惕，不要轻易相信任何人。
- 夜晚阶段的结果提供重要线索。利用这些信息识别狼人。`,

	game.RoleSeer: `## 预言家游戏指导
- 预言家对村民非常重要，过早暴露自己可能导致被狼人盯上。
- 你查验一名玩家身份的能力对村民至关重要。
- 考虑何时揭示你的身份并与其他村民分享你的发现。`,

	game.RoleWitch: `## 女巫游戏指导
- 女巫有两瓶强力药水，明智地使用它们来保护关键村民或淘汰嫌疑狼人。
- 如果你被狼人杀死，你不能救自己。
- 在使用药水之前考虑游戏局势。`,

	game.RoleHunter: `## 猎人游戏指导
- 在白天使用你的能力会暴露你的角色（因为只有猎人可以带走一名玩家）。
- 你的开枪能力在你被淘汰时激活（被女巫毒死除外）。
- 在讨论中表现得像普通村民，避免被盯上。`,
}

// BuildPlayerInstruction 构建玩家系统提示
func BuildPlayerInstruction(name string, role game.Role) string {
	guidance := RoleGuidance[role]
	return fmt.Sprintf(Prompts.BaseSystem, name, role, guidance)
}
