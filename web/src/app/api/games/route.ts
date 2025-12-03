import { NextResponse } from 'next/server';
import fs from 'fs';
import path from 'path';

export async function GET() {
  try {
    // 日志目录在项目根目录的 logs 文件夹
    // process.cwd() 在 web 目录下运行时指向 web 目录
    const possiblePaths = [
      path.join(process.cwd(), '..', 'logs'),
      path.join(process.cwd(), 'logs'),
      '/Users/mervyn/go/src/eino-examples-x/week11-homework/werewolves-adk/logs',
    ];
    
    let logsDir = '';
    for (const p of possiblePaths) {
      if (fs.existsSync(p)) {
        logsDir = p;
        break;
      }
    }
    
    if (!logsDir) {
      return NextResponse.json({ games: [], error: 'Logs directory not found' });
    }

    const entries = fs.readdirSync(logsDir, { withFileTypes: true });
    const games = [];

    for (const entry of entries) {
      if (entry.isDirectory()) {
        const gameId = entry.name;
        const logPath = path.join(logsDir, gameId, 'full_log.md');
        
        if (fs.existsSync(logPath)) {
          const content = fs.readFileSync(logPath, 'utf-8');
          
          // 解析胜利方和回合数
          let winner: string | undefined;
          let rounds: number | undefined;
          
          if (content.includes('狼人阵营获胜') || content.includes('Werewolves Win')) {
            winner = 'werewolf';
          } else if (content.includes('好人阵营获胜') || content.includes('Villagers Win')) {
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
            content,
          });
        }
      }
    }

    // 按时间倒序排列
    games.sort((a, b) => b.id.localeCompare(a.id));

    return NextResponse.json({ games });
  } catch (error) {
    console.error('Failed to read games:', error);
    return NextResponse.json({ games: [], error: 'Failed to read games' }, { status: 500 });
  }
}
