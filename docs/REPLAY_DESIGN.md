# 🐺 狼人杀游戏回放系统设计文档

## 1. 概述

基于现有的 Markdown 日志文件，实现一个美观的 Web 回放系统。参考 deer-flow 的 UI 设计风格，使用现代化的暗色主题和流畅的动画效果。

## 2. UI 设计参考

参考 deer-flow 的设计语言：
- **暗色主题** - 深色背景 + 高对比度文字
- **卡片布局** - 圆角卡片 + 微妙边框
- **流畅动画** - Framer Motion 风格过渡
- **现代字体** - Geist Sans 字体族

## 3. 技术栈

| 技术 | 用途 |
|------|------|
| **Next.js 14** | React 框架 |
| **TailwindCSS** | 样式系统 |
| **shadcn/ui** | UI 组件库 |
| **Framer Motion** | 动画效果 |
| **marked.js** | Markdown 解析 |
| **Lucide Icons** | 图标库 |

## 4. 配色方案

```css
/* 暗色主题 - 参考 deer-flow */
:root {
  --background: oklch(0.145 0 0);        /* 深色背景 */
  --foreground: oklch(0.985 0 0);        /* 亮色文字 */
  --card: oklch(0.205 0 0);              /* 卡片背景 */
  --card-foreground: oklch(0.985 0 0);   /* 卡片文字 */
  --muted: oklch(0.269 0 0);             /* 次要背景 */
  --muted-foreground: oklch(0.708 0 0);  /* 次要文字 */
  --border: oklch(1 0 0 / 20%);          /* 边框 */
  --brand: rgb(17, 103, 234);            /* 品牌色 */
  
  /* 狼人杀角色配色 */
  --werewolf: #dc2626;    /* 狼人 - 红色 */
  --villager: #22c55e;    /* 村民 - 绿色 */
  --seer: #a855f7;        /* 预言家 - 紫色 */
  --witch: #06b6d4;       /* 女巫 - 青色 */
  --hunter: #f59e0b;      /* 猎人 - 橙色 */
}
```

## 5. 页面布局

```
┌─────────────────────────────────────────────────────────────────────────┐
│  🐺 狼人杀回放                                              [🌙 暗色模式] │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  📂 选择对局                                                     │   │
│  │  ┌─────────────────────────────────────────────────────────────┐│   │
│  │  │ 🐺 20251204_001037  │  狼人胜利  │  4回合  │  15分钟        ││   │
│  │  │ 🏘️ 20251203_222741  │  村民胜利  │  6回合  │  22分钟        ││   │
│  │  └─────────────────────────────────────────────────────────────┘│   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                         回放内容区域                             │   │
│  │  ┌───────────────────────────────────────────────────────────┐ │   │
│  │  │  # 🐺 狼人杀游戏完整日志                                   │ │   │
│  │  │                                                           │ │   │
│  │  │  **游戏ID**: 20251204_001037                              │ │   │
│  │  │  **开始时间**: 2025-12-04 00:10:37                        │ │   │
│  │  │                                                           │ │   │
│  │  │  ## 📋 角色分配                                            │ │   │
│  │  │  ┌──────────┬──────────┐                                  │ │   │
│  │  │  │  玩家    │  角色    │                                  │ │   │
│  │  │  ├──────────┼──────────┤                                  │ │   │
│  │  │  │ Player1  │ 🐺 狼人  │                                  │ │   │
│  │  │  │ Player2  │ 🔮 预言家│                                  │ │   │
│  │  │  └──────────┴──────────┘                                  │ │   │
│  │  │                                                           │ │   │
│  │  │  ## 🔄 第 1 回合                                           │ │   │
│  │  │  ### 🌙 夜间阶段                                           │ │   │
│  │  │                                                           │ │   │
│  │  │  #### 🎭 Player1 (werewolf)                               │ │   │
│  │  │  ┌─────────────────────────────────────────────────────┐  │ │   │
│  │  │  │ **提示词**                                          │  │ │   │
│  │  │  │ [WEREWOLVES ONLY] Discuss with your fellow...       │  │ │   │
│  │  │  └─────────────────────────────────────────────────────┘  │ │   │
│  │  │  ┌─────────────────────────────────────────────────────┐  │ │   │
│  │  │  │ **回复**                                            │  │ │   │
│  │  │  │ 我认为我们应该杀 Player2，他可能是预言家... █       │  │ │   │
│  │  │  └─────────────────────────────────────────────────────┘  │ │   │
│  │  │                                                           │ │   │
│  │  └───────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  [⏮️] [⏪] [▶️ 播放] [⏩] [⏭️]   速度: [1x ▼]   进度: ████░░ 65%  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## 6. 组件设计

### 6.1 对局选择卡片

```tsx
function GameCard({ game, isSelected, onClick }) {
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      className={cn(
        "p-4 rounded-xl border cursor-pointer transition-colors",
        "bg-card hover:bg-accent",
        isSelected && "border-brand ring-2 ring-brand/20"
      )}
      onClick={onClick}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-2xl">
            {game.winner === 'werewolf' ? '🐺' : '🏘️'}
          </span>
          <div>
            <div className="font-medium">{game.id}</div>
            <div className="text-sm text-muted-foreground">
              {game.rounds} 回合 · {game.duration}
            </div>
          </div>
        </div>
        <Badge variant={game.winner === 'werewolf' ? 'destructive' : 'success'}>
          {game.winner === 'werewolf' ? '狼人胜利' : '村民胜利'}
        </Badge>
      </div>
    </motion.div>
  );
}
```

### 6.2 回放内容区域

```tsx
function ReplayContent({ segments, currentIndex }) {
  return (
    <ScrollContainer className="flex-1 p-6">
      <div className="prose prose-invert max-w-none">
        {segments.slice(0, currentIndex + 1).map((segment, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
          >
            <SegmentRenderer segment={segment} isLatest={i === currentIndex} />
          </motion.div>
        ))}
      </div>
    </ScrollContainer>
  );
}
```

### 6.3 段落渲染器

```tsx
function SegmentRenderer({ segment, isLatest }) {
  // 根据段落类型渲染不同样式
  if (segment.type === 'round') {
    return (
      <h2 className="text-xl font-bold flex items-center gap-2 mt-8 mb-4">
        <span className="text-2xl">🔄</span>
        {segment.content}
      </h2>
    );
  }
  
  if (segment.type === 'phase') {
    const isNight = segment.content.includes('夜');
    return (
      <h3 className={cn(
        "text-lg font-semibold flex items-center gap-2 mt-6 mb-3",
        isNight ? "text-indigo-400" : "text-amber-400"
      )}>
        <span className="text-xl">{isNight ? '🌙' : '☀️'}</span>
        {segment.content}
      </h3>
    );
  }
  
  if (segment.type === 'player_action') {
    return (
      <Card className="my-4 overflow-hidden">
        <CardHeader className="pb-2">
          <div className="flex items-center gap-2">
            <RoleIcon role={segment.role} />
            <span className="font-medium">{segment.player}</span>
            <Badge variant="outline">{segment.role}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="bg-muted/50 rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">提示词</div>
            <code className="text-sm">{segment.prompt}</code>
          </div>
          <div className="bg-muted/50 rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">回复</div>
            <TypeWriter text={segment.response} isActive={isLatest} />
          </div>
        </CardContent>
      </Card>
    );
  }
  
  // 默认 Markdown 渲染
  return <div dangerouslySetInnerHTML={{ __html: marked(segment.content) }} />;
}
```

### 6.4 回放控制栏

```tsx
function ReplayControls({ 
  isPlaying, 
  speed, 
  progress,
  onPlay, 
  onPause, 
  onSpeedChange,
  onSeek 
}) {
  return (
    <div className="flex items-center gap-4 p-4 bg-card/50 backdrop-blur border-t">
      {/* 播放控制按钮 */}
      <div className="flex items-center gap-1">
        <Button variant="ghost" size="icon" onClick={() => onSeek(0)}>
          <SkipBack className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="icon" onClick={() => onSeek(progress - 10)}>
          <Rewind className="h-4 w-4" />
        </Button>
        <Button 
          variant="default" 
          size="icon" 
          className="h-10 w-10"
          onClick={isPlaying ? onPause : onPlay}
        >
          {isPlaying ? <Pause className="h-5 w-5" /> : <Play className="h-5 w-5" />}
        </Button>
        <Button variant="ghost" size="icon" onClick={() => onSeek(progress + 10)}>
          <FastForward className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="icon" onClick={() => onSeek(100)}>
          <SkipForward className="h-4 w-4" />
        </Button>
      </div>
      
      {/* 速度选择 */}
      <Select value={speed} onValueChange={onSpeedChange}>
        <SelectTrigger className="w-20">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="0.5">0.5x</SelectItem>
          <SelectItem value="1">1x</SelectItem>
          <SelectItem value="2">2x</SelectItem>
          <SelectItem value="4">4x</SelectItem>
        </SelectContent>
      </Select>
      
      {/* 进度条 */}
      <div className="flex-1">
        <Slider
          value={[progress]}
          max={100}
          step={1}
          onValueChange={([v]) => onSeek(v)}
          className="cursor-pointer"
        />
      </div>
      
      {/* 进度百分比 */}
      <span className="text-sm text-muted-foreground w-12 text-right">
        {progress}%
      </span>
    </div>
  );
}
```

### 6.5 打字机效果

```tsx
function TypeWriter({ text, isActive, speed = 30 }) {
  const [displayText, setDisplayText] = useState('');
  const [isComplete, setIsComplete] = useState(false);
  
  useEffect(() => {
    if (!isActive) {
      setDisplayText(text);
      setIsComplete(true);
      return;
    }
    
    let index = 0;
    setDisplayText('');
    setIsComplete(false);
    
    const timer = setInterval(() => {
      if (index < text.length) {
        setDisplayText(text.slice(0, index + 1));
        index++;
      } else {
        setIsComplete(true);
        clearInterval(timer);
      }
    }, speed);
    
    return () => clearInterval(timer);
  }, [text, isActive, speed]);
  
  return (
    <span>
      {displayText}
      {!isComplete && (
        <motion.span
          animate={{ opacity: [1, 0] }}
          transition={{ duration: 0.5, repeat: Infinity }}
          className="inline-block w-2 h-4 bg-brand ml-0.5"
        />
      )}
    </span>
  );
}
```

## 7. 动画效果

### 7.1 段落进入动画

```tsx
const segmentVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { 
    opacity: 1, 
    y: 0,
    transition: { duration: 0.3, ease: "easeOut" }
  }
};
```

### 7.2 阶段切换动画

```tsx
const phaseTransition = {
  night: {
    background: "linear-gradient(to bottom, #1e1b4b, #0f172a)",
    transition: { duration: 0.5 }
  },
  day: {
    background: "linear-gradient(to bottom, #fef3c7, #fcd34d)",
    transition: { duration: 0.5 }
  }
};
```

### 7.3 彩虹文字效果 (胜利时)

```css
.rainbow-text {
  background: linear-gradient(
    to right,
    rgba(255, 255, 255, 0.3) 15%,
    rgba(255, 255, 255, 0.75) 35%,
    rgba(255, 255, 255, 0.75) 65%,
    rgba(255, 255, 255, 0.3) 85%
  );
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  background-size: 500% auto;
  animation: textShine 2s ease-in-out infinite alternate;
}

@keyframes textShine {
  0% { background-position: 0% 50%; }
  100% { background-position: 100% 50%; }
}
```

## 8. 文件结构

```
werewolves-adk/
├── web/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx              # 首页/回放列表
│   │   └── replay/
│   │       └── [id]/
│   │           └── page.tsx      # 回放播放页
│   ├── components/
│   │   ├── ui/                   # shadcn/ui 组件
│   │   └── replay/
│   │       ├── game-card.tsx
│   │       ├── replay-content.tsx
│   │       ├── replay-controls.tsx
│   │       ├── segment-renderer.tsx
│   │       └── type-writer.tsx
│   ├── lib/
│   │   ├── parser.ts             # Markdown 解析
│   │   └── utils.ts
│   ├── styles/
│   │   └── globals.css
│   ├── public/
│   │   └── logs/                 # 日志文件 (符号链接)
│   ├── package.json
│   ├── tailwind.config.ts
│   └── next.config.js
└── logs/
    └── {gameID}/
        └── full_log.md
```

## 9. 实现计划

### Phase 1: 项目搭建 (1小时)

- [ ] 创建 Next.js 项目
- [ ] 配置 TailwindCSS
- [ ] 安装 shadcn/ui 组件
- [ ] 配置暗色主题

### Phase 2: 核心功能 (2小时)

- [ ] Markdown 解析器
- [ ] 段落类型识别
- [ ] 回放状态管理
- [ ] 基础 UI 布局

### Phase 3: UI 组件 (2小时)

- [ ] 对局选择卡片
- [ ] 回放内容区域
- [ ] 段落渲染器
- [ ] 回放控制栏

### Phase 4: 动画效果 (1小时)

- [ ] 段落进入动画
- [ ] 打字机效果
- [ ] 阶段切换效果
- [ ] 胜利庆祝动画

### Phase 5: 完善优化 (1小时)

- [ ] 响应式设计
- [ ] 键盘快捷键
- [ ] 性能优化
- [ ] 错误处理

**总计约 7 小时**

## 10. 使用方式

```bash
# 1. 安装依赖
cd werewolves-adk/web
pnpm install

# 2. 启动开发服务器
pnpm dev

# 3. 访问
open http://localhost:3000
```
