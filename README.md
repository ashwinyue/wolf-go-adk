# 🐺 狼人杀多智能体系统 (ADK 版本)

基于 Eino ADK (Agent Development Kit) 实现的狼人杀游戏，遵循设计文档的最佳实践。

## 📋 项目概述

本项目使用 Eino ADK 的多智能体架构实现了一个完整的狼人杀游戏系统，包含 **10 个独立的 Agent**：
- 9 名玩家 Agent（每个都是独立的 `ChatModelAgent`）
- 1 个游戏主控 Agent（自定义 `Supervisor Agent`）

### 角色配置

| 阵营 | 角色 | 数量 | Agent 类型 | 核心职责 |
|------|------|------|------------|----------|
| 狼人阵营 | 狼人 | 3 | ChatModelAgent | 夜间协作击杀村民，白天隐藏身份 |
| 村民阵营 | 村民 | 3 | ChatModelAgent | 通过推理找出狼人 |
| 村民阵营 | 预言家 | 1 | ChatModelAgent | 每晚查验一名玩家的阵营 |
| 村民阵营 | 女巫 | 1 | ChatModelAgent | 拥有解药和毒药各一瓶 |
| 村民阵营 | 猎人 | 1 | ChatModelAgent | 被淘汰时可开枪带走一人 |
| 系统 | 游戏主控 | 1 | 自定义 Agent (Supervisor) | 编排游戏流程，协调玩家 Agent |

## 🏗️ 架构设计

### 多 Agent 架构

根据设计文档，本系统实现了真正的多 Agent 架构：

```
WerewolfGameAgent (Supervisor - 自定义 Agent)
├── Player1 Agent (ChatModelAgent) - 狼人
├── Player2 Agent (ChatModelAgent) - 狼人  
├── Player3 Agent (ChatModelAgent) - 狼人
├── Player4 Agent (ChatModelAgent) - 村民
├── Player5 Agent (ChatModelAgent) - 村民
├── Player6 Agent (ChatModelAgent) - 村民
├── Player7 Agent (ChatModelAgent) - 预言家
├── Player8 Agent (ChatModelAgent) - 女巫
└── Player9 Agent (ChatModelAgent) - 猎人
```

### 核心 ADK 模式

| 设计要求 | 实现方式 |
|----------|----------|
| **Supervisor 模式** | `WerewolfGameAgent` 作为中央协调器 |
| **LoopAgent 模式** | 游戏主循环在 `Run()` 中实现，使用 `ExitAction` 终止 |
| **SequentialAgent 模式** | 夜晚/白天阶段按顺序执行各玩家 Agent |
| **Transfer Action** | 调用 `playerAgent.Run()` 将控制权传递给玩家 |
| **ChatModelAgent** | 每个玩家都是独立的 Agent，有自己的 ReAct 循环 |
| **Tool Calling** | 7 个工具实现特殊能力 |

### 目录结构

```
werewolves-adk/
├── main.go              # 程序入口
├── agents/
│   ├── game_agent.go    # 游戏主控 Agent (Supervisor)
│   └── players.go       # 玩家 Agent 工厂 (ChatModelAgent)
├── game/
│   └── state.go         # 游戏状态 + 日志记录器
└── tools/
    └── tools.go         # 特殊能力工具
```

## 🛠️ 工具实现

| 工具 | 角色 | 功能 |
|------|------|------|
| `discuss` | 狼人 | 与其他狼人交流 |
| `kill` | 狼人 | 选择击杀目标 |
| `check_identity` | 预言家 | 查验玩家阵营 |
| `save` | 女巫 | 使用解药救人 |
| `poison` | 女巫 | 使用毒药毒人 |
| `shoot` | 猎人 | 开枪射杀玩家 |
| `vote` | 所有玩家 | 投票淘汰玩家 |

## 🎮 游戏流程

### 夜晚阶段 (Sequential Transfer Action)

1. **狼人行动** - 依次调用狼人 Agent 进行讨论和投票
2. **预言家行动** - 调用预言家 Agent 进行查验
3. **女巫行动** - 调用女巫 Agent 决定用药
4. **结算** - 处理死亡

### 白天阶段 (Sequential + Parallel Transfer Action)

1. **死亡公告** - 宣布夜间死亡玩家
2. **讨论阶段** - 依次调用存活玩家 Agent 发言
3. **投票阶段** - 并行调用所有存活玩家 Agent 投票
4. **猎人开枪** - 条件触发猎人 Agent

## 🚀 运行方式

### 后端游戏

```bash
# 设置环境变量（使用 OpenAI 兼容 API）
export OPENAI_API_KEY=your-api-key
export OPENAI_MODEL=qwen-max
export OPENAI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# 运行游戏
go run .
```

### 前端回放

```bash
# 进入前端目录
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

访问 http://localhost:3000 查看游戏回放。

## 🎬 在线演示

GitHub Pages: https://ashwinyue.github.io/wolf-go/

## 📊 与 werewolves-go 版本对比

| 特性 | werewolves-go | werewolves-adk |
|------|---------------|----------------|
| Agent 数量 | 1 (Engine) | 10 (1 Supervisor + 9 Players) |
| 架构模式 | 自定义 Engine | ADK Supervisor + ChatModelAgent |
| 玩家实现 | 直接调用 ChatModel | 独立的 ChatModelAgent |
| 控制流 | 函数调用 | Transfer Action (Agent.Run) |
| 工作流 | goroutine + WaitGroup | Sequential/Parallel Agent 调用 |
| 事件流 | 同步执行 | AsyncIterator |
| 工具绑定 | model.WithTools() | adk.ToolsConfig |
| 游戏终止 | return | ExitAction |

## 🔮 未来扩展

根据设计文档，可以进一步实现：

1. **RAG 记忆系统** - 实现长期语义记忆，让玩家能回忆历史事件
2. **人机协作 (HITL)** - 使用 ADK 的 Interrupt/Resume 允许人类玩家参与
3. **可视化界面** - 通过 WebSocket 消费 AsyncIterator 事件流
4. **AgentWithDeterministicTransferTo** - 确保控制权可靠返回

## 📄 License

Apache License 2.0
