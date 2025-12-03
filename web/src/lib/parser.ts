// 日志段落类型
export type SegmentType = 
  | 'title'
  | 'info'
  | 'round'
  | 'phase'
  | 'message'      // 玩家消息（对话形式）
  | 'system'       // 系统消息
  | 'result'
  | 'winner';

export interface Segment {
  type: SegmentType;
  content: string;
  delay: number;
  player?: string;
  role?: string;
  isAction?: boolean;  // 是否是行动结果
}

// 角色信息
export const ROLES: Record<string, { name: string; icon: string; color: string }> = {
  'werewolf': { name: '狼人', icon: '🐺', color: '#dc2626' },
  'villager': { name: '村民', icon: '👨‍🌾', color: '#22c55e' },
  'seer': { name: '预言家', icon: '🔮', color: '#a855f7' },
  'witch': { name: '女巫', icon: '🧙‍♀️', color: '#06b6d4' },
  'hunter': { name: '猎人', icon: '🎯', color: '#f59e0b' },
  '狼人': { name: '狼人', icon: '🐺', color: '#dc2626' },
  '村民': { name: '村民', icon: '👨‍🌾', color: '#22c55e' },
  '预言家': { name: '预言家', icon: '🔮', color: '#a855f7' },
  '女巫': { name: '女巫', icon: '🧙‍♀️', color: '#06b6d4' },
  '猎人': { name: '猎人', icon: '🎯', color: '#f59e0b' },
  'moderator': { name: '主持人', icon: '🎭', color: '#6b7280' },
};

// 解析 Markdown 日志为段落数组
export function parseLog(markdown: string): Segment[] {
  const segments: Segment[] = [];
  const paragraphs = markdown.split(/\n\n+/);
  
  // 用于存储角色分配
  const playerRoles: Record<string, string> = {};
  // 用于去重
  const seenMessages = new Set<string>();
  
  for (const para of paragraphs) {
    const trimmed = para.trim();
    if (!trimmed || trimmed === '---') continue;
    
    // 跳过代码块、分隔线和提示词相关内容
    if (trimmed.startsWith('```') || 
        trimmed.startsWith('---') ||
        trimmed.startsWith('#### ') ||
        trimmed.includes('**提示词**') ||
        trimmed.includes('**回复**') ||
        trimmed.startsWith('[仅狼人可见]') ||
        trimmed.startsWith('[WEREWOLVES') ||
        trimmed.startsWith('[Previous') ||
        trimmed.includes('Moderator:') ||
        trimmed.includes('讨论要点') ||
        trimmed.includes('reach_agreement')) {
      continue;
    }
    
    // 解析角色分配表格
    if (trimmed.includes('| 玩家 | 角色 |') || trimmed.includes('|------|------|')) {
      continue;
    }
    const tableMatch = trimmed.match(/\| (Player\d+) \| (\w+) \|/);
    if (tableMatch) {
      playerRoles[tableMatch[1]] = tableMatch[2];
      continue;
    }
    
    // 跳过标题和游戏信息
    if (trimmed.startsWith('# 🐺') || 
        trimmed.startsWith('**游戏ID**') || 
        trimmed.startsWith('**开始时间**')) {
      continue;
    }
    
    // 回合
    if (trimmed.startsWith('## 🔄') || trimmed.startsWith('## 第')) {
      segments.push({ type: 'round', content: trimmed.replace('## ', ''), delay: 400 });
      continue;
    }
    
    // 阶段
    if (trimmed.startsWith('### ')) {
      segments.push({ type: 'phase', content: trimmed.replace('### ', ''), delay: 300 });
      continue;
    }
    
    // 主持人消息: 🎭 **主持人**: 消息
    const moderatorMatch = trimmed.match(/🎭\s*\*\*主持人\*\*:\s*(.+)/);
    if (moderatorMatch) {
      segments.push({
        type: 'message',
        content: moderatorMatch[1],
        delay: 300,
        player: '主持人',
        role: 'moderator',
      });
      continue;
    }
    
    // 反思消息: 🐺 **Player1**: 💭 消息 (必须在玩家消息之前匹配)
    const reflectIconMatch = trimmed.match(/^(🐺|🔮|🧙‍♀️|🎯|👨‍🌾|🎭)\s*\*\*(\w+)\*\*:\s*💭\s*(.+)$/);
    if (reflectIconMatch) {
      const icon = reflectIconMatch[1];
      const player = reflectIconMatch[2];
      // 根据图标确定角色
      const iconToRole: Record<string, string> = {
        '🐺': 'werewolf',
        '🔮': 'seer',
        '🧙‍♀️': 'witch',
        '🎯': 'hunter',
        '👨‍🌾': 'villager',
        '🎭': 'moderator',
      };
      const role = iconToRole[icon] || playerRoles[player] || 'villager';
      const content = reflectIconMatch[3].replace(/\[仅\w+可见\]\s*/, '').trim();
      if (content) {
        segments.push({
          type: 'message',
          content: `💭 ${content}`,
          delay: 500,
          player,
          role,
        });
      }
      continue;
    }
    
    // 玩家消息: 🐺 **Player1**: 消息 (支持各种角色图标)
    const playerMsgMatch = trimmed.match(/^(🐺|🔮|🧙‍♀️|🎯|👨‍🌾|🎭)\s*\*\*(\w+)\*\*:\s*(.+)$/);
    if (playerMsgMatch) {
      const icon = playerMsgMatch[1];
      const player = playerMsgMatch[2];
      // 根据图标确定角色
      const iconToRole: Record<string, string> = {
        '🐺': 'werewolf',
        '🔮': 'seer',
        '🧙‍♀️': 'witch',
        '🎯': 'hunter',
        '👨‍🌾': 'villager',
        '🎭': 'moderator',
      };
      const role = iconToRole[icon] || playerRoles[player] || 'villager';
      // 移除可能的引号
      const content = playerMsgMatch[3].replace(/^[""\u201c]|[""\u201d]$/g, '');
      const msgKey = `${player}:${content.substring(0, 50)}`;
      if (!seenMessages.has(msgKey)) {
        seenMessages.add(msgKey);
        segments.push({
          type: 'message',
          content,
          delay: 400,
          player,
          role,
        });
      }
      continue;
    }
    
    // 玩家发言: **[Player1]**: 消息 或 **[Player1]** (第X轮): 消息
    const discussMatch = trimmed.match(/\*\*\[(\w+)\]\*\*(?:\s*\(第\d+轮\))?:\s*(.+)/);
    if (discussMatch) {
      const player = discussMatch[1];
      const content = discussMatch[2];
      const msgKey = `${player}:${content.substring(0, 50)}`;
      if (!seenMessages.has(msgKey)) {
        seenMessages.add(msgKey);
        segments.push({
          type: 'message',
          content,
          delay: 400,
          player,
          role: playerRoles[player] || 'villager',
        });
      }
      continue;
    }
    
    // 玩家消息: **🐺 Player1**: 消息
    const msgMatch = trimmed.match(/\*\*(?:🐺|🔮|🧙‍♀️|🎯|👨‍🌾)\s*(\w+)\*\*:\s*(.+)/);
    if (msgMatch) {
      const player = msgMatch[1];
      const content = msgMatch[2];
      const msgKey = `${player}:${content.substring(0, 50)}`;
      if (!seenMessages.has(msgKey)) {
        seenMessages.add(msgKey);
        segments.push({
          type: 'message',
          content,
          delay: 400,
          player,
          role: playerRoles[player] || 'villager',
        });
      }
      continue;
    }
    
    // 投票行: - Player1 → Player2 或 - **Player1** 投票: Player2
    const voteMatch = trimmed.match(/^-\s*\*?\*?(\w+)\*?\*?\s*(?:→|投票[:：])\s*(\w+)/);
    if (voteMatch) {
      segments.push({
        type: 'system',
        content: `${voteMatch[1]} 投票给 ${voteMatch[2]}`,
        delay: 200,
      });
      continue;
    }
    
    // 行动结果
    if (trimmed.includes('决定击杀') || trimmed.includes('查验') || 
        trimmed.includes('投票结果') || trimmed.includes('使用解药') ||
        trimmed.includes('使用毒药') || trimmed.includes('射杀')) {
      segments.push({
        type: 'result',
        content: trimmed.replace(/\*\*/g, ''),
        delay: 500,
        isAction: true,
      });
      continue;
    }
    
    // 遗言: **[Player1 遗言]**: 消息
    const lastWordsMatch = trimmed.match(/\*\*\[(\w+)\s*遗言\]\*\*:\s*(.+)/);
    if (lastWordsMatch) {
      segments.push({
        type: 'message',
        content: `💀 遗言: ${lastWordsMatch[2]}`,
        delay: 500,
        player: lastWordsMatch[1],
        role: playerRoles[lastWordsMatch[1]] || 'villager',
      });
      continue;
    }
    
    // 反思旧格式: **[Player1 反思]**: 消息
    const reflectMatch = trimmed.match(/\*?\*?\[(\w+)\s*反思\]\*?\*?[:：]\s*(.*)/);
    if (reflectMatch) {
      // 移除 [仅PlayerX可见] 前缀和"反思："前缀
      let content = reflectMatch[2]
        .replace(/\[仅\w+可见\]\s*/, '')
        .replace(/^反思[:：]\s*/, '')
        .trim();
      // 跳过空白反思
      if (!content) continue;
      segments.push({
        type: 'message',
        content: `💭 ${content}`,
        delay: 500,
        player: reflectMatch[1],
        role: playerRoles[reflectMatch[1]] || 'villager',
      });
      continue;
    }
    
    // 胜利
    if (trimmed.includes('🏆') || trimmed.includes('获胜')) {
      segments.push({ type: 'winner', content: trimmed.replace(/[#*]/g, '').trim(), delay: 800 });
      continue;
    }
    
    // 夜晚结算
    if (trimmed.includes('夜晚结算')) {
      const lines = trimmed.split('\n').filter(l => l.trim());
      for (const line of lines) {
        segments.push({
          type: 'result',
          content: line.replace(/^-\s*/, '').replace(/\*\*/g, ''),
          delay: 300,
        });
      }
      continue;
    }
    
    // 其他系统消息（排除反思和遗言）
    if (trimmed.startsWith('**') && !trimmed.includes('提示词') && !trimmed.includes('反思') && !trimmed.includes('遗言')) {
      segments.push({
        type: 'system',
        content: trimmed.replace(/\*\*/g, '').trim(),
        delay: 200,
      });
      continue;
    }
  }
  
  return segments;
}

// 获取角色颜色
export function getRoleColor(role?: string): string {
  if (!role) return '#ededed';
  
  const colors: Record<string, string> = {
    werewolf: '#dc2626',
    villager: '#22c55e',
    seer: '#a855f7',
    witch: '#06b6d4',
    hunter: '#f59e0b',
    '狼人': '#dc2626',
    '村民': '#22c55e',
    '预言家': '#a855f7',
    '女巫': '#06b6d4',
    '猎人': '#f59e0b',
  };
  
  return colors[role.toLowerCase()] || '#ededed';
}

// 获取角色图标
export function getRoleIcon(role?: string): string {
  if (!role) return '👤';
  
  const icons: Record<string, string> = {
    werewolf: '🐺',
    villager: '👨‍🌾',
    seer: '🔮',
    witch: '🧙‍♀️',
    hunter: '🎯',
    '狼人': '🐺',
    '村民': '👨‍🌾',
    '预言家': '🔮',
    '女巫': '🧙‍♀️',
    '猎人': '🎯',
  };
  
  return icons[role.toLowerCase()] || '👤';
}
