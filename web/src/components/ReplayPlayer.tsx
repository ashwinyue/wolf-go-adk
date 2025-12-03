'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Play, Pause, SkipBack, SkipForward, FastForward } from 'lucide-react';
import { parseLog, Segment, ROLES } from '@/lib/parser';

interface ReplayPlayerProps {
  markdown: string;
  gameId: string;
  onBack?: () => void;
}

// 头像组件
function Avatar({ role, size = 40 }: { role?: string; size?: number }) {
  const roleInfo = ROLES[role || ''] || ROLES['villager'];
  return (
    <div 
      className="flex items-center justify-center rounded-full text-white font-bold shrink-0"
      style={{ 
        width: size, 
        height: size, 
        backgroundColor: roleInfo.color,
        fontSize: size * 0.5,
      }}
    >
      {roleInfo.icon}
    </div>
  );
}

// 消息气泡组件 - 微信风格
function MessageBubble({ segment, index }: { segment: Segment; index: number }) {
  const roleInfo = ROLES[segment.role || ''] || ROLES['villager'];
  const isSystem = segment.player === 'Moderator' || segment.player === '主持人';
  
  // 主持人消息在左边
  if (isSystem) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
        className="flex gap-3 mb-5"
      >
        <div className="w-10 h-10 rounded-md overflow-hidden flex-shrink-0 bg-blue-500 flex items-center justify-center">
          <span className="text-xl">🎭</span>
        </div>
        <div className="max-w-[70%]">
          <div className="text-xs text-gray-500 mb-1">主持人</div>
          <div className="bg-white rounded-lg px-3 py-2 text-gray-800 text-sm shadow-sm">
            {segment.content}
          </div>
        </div>
      </motion.div>
    );
  }
  
  // 玩家消息在右边
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className="flex gap-3 mb-5 flex-row-reverse"
    >
      <div 
        className="w-10 h-10 rounded-md overflow-hidden flex-shrink-0 flex items-center justify-center"
        style={{ backgroundColor: roleInfo.color }}
      >
        <span className="text-xl">{roleInfo.icon}</span>
      </div>
      <div className="max-w-[70%] flex flex-col items-end">
        <div className="text-xs text-gray-500 mb-1">{segment.player}</div>
        <div className="bg-[#95ec69] rounded-lg px-3 py-2 text-gray-800 text-sm shadow-sm">
          {segment.content}
        </div>
      </div>
    </motion.div>
  );
}

// 系统消息组件
function SystemMessage({ segment }: { segment: Segment }) {
  const isWinner = segment.type === 'winner';
  
  // 胜利消息特殊处理
  if (isWinner) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5 }}
        className="text-center my-8"
      >
        <div className="text-4xl mb-2">🏆</div>
        <div className="text-2xl font-bold text-yellow-400">
          {segment.content.replace('🏆', '').trim()}
        </div>
      </motion.div>
    );
  }
  
  // 普通系统消息 - 简洁文字，不用框
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
      className="text-center my-3 text-sm text-[var(--muted-foreground)]"
    >
      {segment.content}
    </motion.div>
  );
}

// 阶段标题组件 - 居中显示
function PhaseHeader({ segment }: { segment: Segment }) {
  const isNight = segment.content.includes('夜') || segment.content.includes('🌙');
  const isRound = segment.type === 'round';
  
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
      className={`flex justify-center ${isRound ? 'my-6' : 'my-4'}`}
    >
      <div className={`
        inline-flex items-center gap-2 px-4 py-1.5 rounded-full text-xs font-medium
        ${isRound 
          ? 'bg-orange-100 text-orange-600 border border-orange-200' 
          : isNight 
            ? 'bg-indigo-100 text-indigo-600 border border-indigo-200' 
            : 'bg-amber-100 text-amber-600 border border-amber-200'
        }
      `}>
        {segment.content}
      </div>
    </motion.div>
  );
}

export default function ReplayPlayer({ markdown, gameId, onBack }: ReplayPlayerProps) {
  const [segments, setSegments] = useState<Segment[]>([]);
  const [visibleCount, setVisibleCount] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const playingRef = useRef(false);

  // 解析日志并自动播放
  useEffect(() => {
    const parsed = parseLog(markdown);
    setSegments(parsed);
    // 自动开始播放
    if (parsed.length > 0) {
      setIsPlaying(true);
    }
  }, [markdown]);

  // 同步 isPlaying 到 ref
  useEffect(() => {
    playingRef.current = isPlaying;
  }, [isPlaying]);

  // 滚动到底部
  const scrollToBottom = useCallback(() => {
    if (containerRef.current) {
      containerRef.current.scrollTo({
        top: containerRef.current.scrollHeight,
        behavior: 'smooth',
      });
    }
  }, []);

  // 显示下一个段落
  const showNext = useCallback(() => {
    setVisibleCount(prev => {
      const next = prev + 1;
      if (next >= segments.length) {
        setIsPlaying(false);
        return segments.length;
      }
      return next;
    });
    setTimeout(scrollToBottom, 50);
  }, [segments.length, scrollToBottom]);

  // 播放控制 - 使用固定间隔
  useEffect(() => {
    if (!isPlaying || visibleCount >= segments.length) {
      return;
    }
    
    const segment = segments[visibleCount];
    const delay = Math.max(200, (segment?.delay || 400) / speed);
    
    const timer = setTimeout(() => {
      if (playingRef.current) {
        showNext();
      }
    }, delay);
    
    return () => clearTimeout(timer);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPlaying, visibleCount, speed]);

  const togglePlay = () => {
    if (visibleCount >= segments.length) {
      setVisibleCount(0);
    }
    setIsPlaying(!isPlaying);
  };

  const reset = () => {
    setIsPlaying(false);
    setVisibleCount(0);
    if (timerRef.current) clearTimeout(timerRef.current);
  };

  const skipToEnd = () => {
    setIsPlaying(false);
    setVisibleCount(segments.length);
    if (timerRef.current) clearTimeout(timerRef.current);
    setTimeout(scrollToBottom, 50);
  };

  const progress = segments.length > 0 
    ? Math.round((visibleCount / segments.length) * 100) 
    : 0;

  // 渲染段落
  const renderSegment = (segment: Segment, index: number) => {
    switch (segment.type) {
      case 'message':
        return <MessageBubble key={index} segment={segment} index={index} />;
      case 'round':
      case 'phase':
        return <PhaseHeader key={index} segment={segment} />;
      case 'result':
      case 'system':
      case 'winner':
        return <SystemMessage key={index} segment={segment} />;
      case 'title':
        return (
          <motion.h1 
            key={index}
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            className="text-3xl font-bold text-center mb-6"
          >
            {segment.content}
          </motion.h1>
        );
      default:
        return (
          <motion.div 
            key={index}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-[var(--muted-foreground)] text-sm my-2"
          >
            {segment.content}
          </motion.div>
        );
    }
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100">
      {/* 标题栏 */}
      <header className="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-200 shadow-sm">
        <div className="flex items-center gap-3">
          {onBack && (
            <button
              onClick={onBack}
              className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition-colors"
            >
              ← 返回
            </button>
          )}
          <span className="text-2xl">🐺</span>
          <h1 className="text-lg font-semibold text-gray-800">狼人杀回放</h1>
        </div>
        <div className="text-sm text-gray-500">
          {gameId}
        </div>
      </header>

      {/* 内容区域 - 白色对话框 */}
      <div className="flex-1 overflow-hidden p-4">
        <div className="max-w-3xl mx-auto h-full flex flex-col bg-white rounded-xl shadow-lg overflow-hidden">
          {/* 对话框标题 */}
          <div className="px-4 py-3 bg-gray-50 border-b border-gray-200 flex items-center gap-2">
            <span className="text-lg">💬</span>
            <span className="font-medium text-gray-700">游戏对话</span>
            <span className="text-xs text-gray-400 ml-auto">{visibleCount} / {segments.length}</span>
          </div>
          
          {/* 消息列表 */}
          <div 
            ref={containerRef}
            className="flex-1 overflow-y-auto p-4 bg-[#ededed]"
          >
            <AnimatePresence>
              {segments.slice(0, visibleCount).map((segment, index) => 
                renderSegment(segment, index)
              )}
            </AnimatePresence>
            {isPlaying && visibleCount < segments.length && (
              <div className="flex gap-3 mb-4 opacity-50">
                <div className="w-10 h-10 rounded-md bg-gray-300 animate-pulse" />
                <div className="flex-1">
                  <div className="h-4 w-24 bg-gray-300 rounded animate-pulse mb-2" />
                  <div className="h-16 bg-white rounded-lg animate-pulse" />
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 控制栏 */}
      <div className="bg-white border-t border-gray-200 px-6 py-3 shadow-sm">
        <div className="max-w-3xl mx-auto flex items-center gap-4">
          {/* 播放控制 */}
          <div className="flex items-center gap-2">
            <button
              onClick={reset}
              className="p-2 rounded-lg hover:bg-gray-100 transition-colors text-gray-600"
              title="重置"
            >
              <SkipBack className="w-5 h-5" />
            </button>
            
            <button
              onClick={togglePlay}
              className="p-3 rounded-full bg-green-500 hover:bg-green-600 transition-colors"
              title={isPlaying ? '暂停' : '播放'}
            >
              {isPlaying ? (
                <Pause className="w-5 h-5 text-white" />
              ) : (
                <Play className="w-5 h-5 text-white" />
              )}
            </button>
            
            <button
              onClick={skipToEnd}
              className="p-2 rounded-lg hover:bg-gray-100 transition-colors text-gray-600"
              title="跳到结尾"
            >
              <SkipForward className="w-5 h-5" />
            </button>
          </div>

          {/* 速度选择 */}
          <div className="flex items-center gap-2">
            <FastForward className="w-4 h-4 text-gray-400" />
            <select
              value={speed}
              onChange={(e) => setSpeed(Number(e.target.value))}
              className="bg-gray-100 border border-gray-200 rounded-lg px-3 py-1.5 text-sm text-gray-700"
            >
              <option value={0.5}>0.5x</option>
              <option value={1}>1x</option>
              <option value={2}>2x</option>
              <option value={4}>4x</option>
              <option value={8}>8x</option>
            </select>
          </div>

          {/* 进度条 */}
          <div className="flex-1 flex items-center gap-3">
            <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
              <motion.div
                className="h-full bg-green-500"
                initial={{ width: 0 }}
                animate={{ width: `${progress}%` }}
                transition={{ duration: 0.3 }}
              />
            </div>
            <span className="text-sm text-gray-500 w-12 text-right">
              {progress}%
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
