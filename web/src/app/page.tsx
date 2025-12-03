'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Play, Clock, Users, Trophy } from 'lucide-react';
import ReplayPlayer from '@/components/ReplayPlayer';

interface GameInfo {
  id: string;
  winner?: string;
  rounds?: number;
  content?: string;
}

export default function Home() {
  const [games, setGames] = useState<GameInfo[]>([]);
  const [selectedGame, setSelectedGame] = useState<GameInfo | null>(null);
  const [gameContent, setGameContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [loadingContent, setLoadingContent] = useState(false);

  // 加载游戏列表 - 支持 API 和静态文件两种模式
  useEffect(() => {
    async function loadGames() {
      try {
        // 优先尝试 API（开发模式）
        let res = await fetch('/api/games');
        if (!res.ok) {
          // 回退到静态文件（生产模式）
          res = await fetch(`${process.env.NEXT_PUBLIC_BASE_PATH || ''}/data/games.json`);
        }
        const data = await res.json();
        setGames(data.games || []);
      } catch (error) {
        console.error('Failed to load games:', error);
      } finally {
        setLoading(false);
      }
    }
    loadGames();
  }, []);

  // 加载游戏内容
  const handleSelectGame = async (game: GameInfo) => {
    setSelectedGame(game);
    
    // 如果已有内容，直接使用
    if (game.content) {
      setGameContent(game.content);
      return;
    }
    
    // 否则从静态文件加载
    setLoadingContent(true);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_BASE_PATH || ''}/data/${game.id}.json`);
      const data = await res.json();
      setGameContent(data.content || '');
    } catch (error) {
      console.error('Failed to load game content:', error);
    } finally {
      setLoadingContent(false);
    }
  };

  // 如果选择了游戏，显示回放
  if (selectedGame) {
    if (loadingContent) {
      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin w-8 h-8 border-2 border-green-500 border-t-transparent rounded-full" />
        </div>
      );
    }
    
    return (
      <ReplayPlayer 
        markdown={gameContent || selectedGame.content || ''} 
        gameId={selectedGame.id}
        onBack={() => { setSelectedGame(null); setGameContent(''); }}
      />
    );
  }

  return (
    <div className="min-h-screen bg-[var(--background)] p-8">
      <div className="max-w-4xl mx-auto">
        {/* 标题 */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            <span className="text-5xl">🐺</span>
            狼人杀游戏回放
          </h1>
          <p className="text-[var(--muted-foreground)]">
            选择一局游戏，观看完整的对局过程
          </p>
        </div>

        {/* 游戏列表 */}
        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin w-8 h-8 border-2 border-[var(--brand)] border-t-transparent rounded-full mx-auto mb-4" />
            <p className="text-[var(--muted-foreground)]">加载中...</p>
          </div>
        ) : games.length === 0 ? (
          <div className="text-center py-12 bg-[var(--card)] rounded-xl border border-[var(--border)]">
            <p className="text-[var(--muted-foreground)] mb-4">暂无游戏记录</p>
            <p className="text-sm text-[var(--muted-foreground)]">
              运行 <code className="bg-[var(--muted)] px-2 py-1 rounded">go run main.go</code> 开始一局游戏
            </p>
          </div>
        ) : (
          <div className="grid gap-4">
            {games.map((game, index) => (
              <motion.div
                key={game.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                onClick={() => handleSelectGame(game)}
                className="p-6 bg-[var(--card)] rounded-xl border border-[var(--border)] cursor-pointer hover:border-[var(--brand)] transition-colors group"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <span className="text-3xl">
                      {game.winner === 'werewolf' ? '🐺' : game.winner === 'villager' ? '🏘️' : '🎮'}
                    </span>
                    <div>
                      <h3 className="font-semibold text-lg">{game.id}</h3>
                      <div className="flex items-center gap-4 text-sm text-[var(--muted-foreground)] mt-1">
                        {game.winner && (
                          <span className="flex items-center gap-1">
                            <Trophy className="w-4 h-4" />
                            {game.winner === 'werewolf' ? '狼人胜利' : '村民胜利'}
                          </span>
                        )}
                        {game.rounds && (
                          <span className="flex items-center gap-1">
                            <Clock className="w-4 h-4" />
                            {game.rounds} 回合
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 text-[var(--brand)] opacity-0 group-hover:opacity-100 transition-opacity">
                    <Play className="w-5 h-5" />
                    <span>播放</span>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
