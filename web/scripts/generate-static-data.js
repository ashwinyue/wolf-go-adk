const fs = require('fs');
const path = require('path');

// 日志目录
const logsDir = path.join(__dirname, '..', '..', 'logs');
const publicDir = path.join(__dirname, '..', 'public');
const dataDir = path.join(publicDir, 'data');

// 确保目录存在
if (!fs.existsSync(dataDir)) {
  fs.mkdirSync(dataDir, { recursive: true });
}

// 读取所有游戏日志
const games = [];

if (fs.existsSync(logsDir)) {
  const entries = fs.readdirSync(logsDir, { withFileTypes: true });
  
  for (const entry of entries) {
    if (entry.isDirectory()) {
      const gameId = entry.name;
      const logPath = path.join(logsDir, gameId, 'full_log.md');
      
      if (fs.existsSync(logPath)) {
        const content = fs.readFileSync(logPath, 'utf-8');
        
        // 解析胜利方和回合数
        let winner;
        let rounds;
        
        if (content.includes('狼人阵营') && content.includes('胜利者')) {
          winner = 'werewolf';
        } else if (content.includes('好人阵营') && content.includes('胜利者')) {
          winner = 'villager';
        }
        
        const roundMatches = content.match(/## 🔄 第 (\d+) 回合/g);
        if (roundMatches) {
          rounds = roundMatches.length;
        }
        
        games.push({
          id: gameId,
          winner,
          rounds,
        });
        
        // 保存单个游戏日志
        fs.writeFileSync(
          path.join(dataDir, `${gameId}.json`),
          JSON.stringify({ id: gameId, content })
        );
      }
    }
  }
}

// 按时间倒序排列
games.sort((a, b) => b.id.localeCompare(a.id));

// 保存游戏列表
fs.writeFileSync(
  path.join(dataDir, 'games.json'),
  JSON.stringify({ games })
);

console.log(`Generated static data for ${games.length} games`);
