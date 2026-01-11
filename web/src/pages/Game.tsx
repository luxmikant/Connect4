import React, { useEffect, useRef } from 'react';
import { Toaster } from 'react-hot-toast';
import { GameBoard } from '../components/GameBoard';
import { useGame } from '../hooks/useGame';
import { useNavigate } from 'react-router-dom';
import { wsService } from '../services/websocket';
import { MessageType } from '../types/websocket';
import { usePlayer } from '../hooks/usePlayer';
import { useGameSound } from '../hooks/useGameSound';
import confetti from 'canvas-confetti';
import { Loader2, ArrowLeft, Users, Cpu } from 'lucide-react';
import { cn } from '../lib/utils';
import { motion } from 'framer-motion';

export const Game: React.FC = () => {
  const { gameState, queueStatus, isMyTurn, myColor, makeMove, leaveQueue } = useGame();
  const { username } = usePlayer();
  const { playSound } = useGameSound();
  const navigate = useNavigate();
  const hasJoinedRef = useRef(false);

  // Get game mode from localStorage
  const gameMode = localStorage.getItem('connect4_gameMode') || 'matchmaking';

  // Join queue/bot game once WebSocket is connected
  useEffect(() => {
    // wait for connection before checking
    const checkInterval = setInterval(() => {
       if (wsService.isConnected() && username && !hasJoinedRef.current) {
          hasJoinedRef.current = true;
          clearInterval(checkInterval);
          if (gameMode === 'bot') {
            console.log("Requesting bot game for:", username);
            wsService.send(MessageType.PlayWithBot, { username });
          } else {
            console.log("Joining matchmaking queue for:", username);
            wsService.send(MessageType.JoinQueue, { username });
          }
       }
    }, 500);
    
    return () => clearInterval(checkInterval);
  }, [username, gameMode]);

  // Cleanup when leaving
  useEffect(() => {
    return () => {
      // If we navigate away, we should probably leave queue if still in it
      if (queueStatus?.inQueue) {
        leaveQueue();
      }
    };
  }, [queueStatus, leaveQueue]);

  // Confetti and Sound effect on win/end
  useEffect(() => {
    if (gameState?.status === 'completed' && gameState.winner) {
      const winnerName = gameState.players[gameState.winner];
      if (winnerName === username) {
        playSound('win');
        confetti({
          particleCount: 150,
          spread: 70,
          origin: { y: 0.6 },
          colors: ['#f43f5e', '#fbbf24', '#3b82f6']
        });
      } else {
        playSound('lose');
      }
    } else if (gameState?.status === 'draw') {
        playSound('draw');
    }
  }, [gameState?.status, gameState?.winner, username, playSound]);

  const handlePlayAgain = () => {
    playSound('click');
    hasJoinedRef.current = false;
    // For bot, we can just re-send join, but nav to lobby is safer
    navigate('/lobby');
  };

  const handleBackToLobby = () => {
    playSound('click');
    if (queueStatus?.inQueue) {
      leaveQueue();
    }
    hasJoinedRef.current = false;
    navigate('/lobby');
  };

  const getStatusContent = () => {
    if (queueStatus?.inQueue) {
      return (
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="w-8 h-8 animate-spin text-game-accent" />
          <p className="text-xl font-medium text-slate-300">
            {gameMode === 'bot' 
              ? 'Initializing Bot...' 
              : `Searching for opponent... (${queueStatus.timeRemaining || '...'})`}
          </p>
        </div>
      );
    }
    
    if (!gameState) {
      return (
        <div className="flex flex-col items-center gap-2">
           <Loader2 className="w-8 h-8 animate-spin text-game-accent" />
           <p className="text-slate-400">Connecting to server...</p>
        </div>
      );
    }
    
    if (gameState.status === 'in_progress') {
       return (
         <div className={cn(
           "px-6 py-3 rounded-full backdrop-blur-md border shadow-lg transition-all duration-300 flex items-center gap-3",
           isMyTurn 
             ? "bg-gradient-to-r from-game-accent/20 to-purple-500/20 border-game-accent/50 text-white" 
             : "bg-slate-800/50 border-slate-700 text-slate-400"
         )}>
           {isMyTurn ? (
              <>
                <span className="relative flex h-3 w-3">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-game-accent opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-3 w-3 bg-game-accent"></span>
                </span>
                <span className="font-bold tracking-wide">YOUR TURN</span>
              </>
           ) : (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>OPPONENT THINKING...</span>
              </>
           )}
         </div>
       );
    }
    
    if (gameState.status === 'completed') {
      const winnerName = gameState.players[gameState.winner || 0];
      const isWinner = winnerName === username;
      const isDraw = !gameState.winner;

      return (
        <motion.div 
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="text-center space-y-4"
        >
          <div className="text-4xl md:text-5xl font-black mb-2 filter drop-shadow-lg">
             {isDraw ? "🤝 DRAW!" : isWinner ? "🎉 VICTORY!" : "💀 DEFEAT"}
          </div>
          <div className="flex gap-4 justify-center">
            <button 
              onClick={handlePlayAgain}
              className="px-6 py-2 bg-game-accent hover:bg-blue-600 text-white rounded-lg font-bold shadow-lg transition-all"
            >
              Play Again
            </button>
            <button 
               onClick={handleBackToLobby}
               className="px-6 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-bold shadow-lg transition-all"
            >
              Lobby
            </button>
          </div>
        </motion.div>
      );
    }
    
    return null;
  };

  return (
    <div className="min-h-screen bg-game-bg text-white font-heading overflow-hidden relative">
      <Toaster position="top-center" />
      
      {/* Ambient Background Elements */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-purple-600/20 rounded-full blur-[100px]" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-600/20 rounded-full blur-[100px]" />
      </div>

      <div className="relative z-10 container mx-auto px-4 py-6 md:py-10 flex flex-col items-center h-full min-h-screen">
        
        {/* Header Bar */}
        <div className="w-full max-w-4xl flex justify-between items-center mb-8 md:mb-12">
          <button 
            onClick={handleBackToLobby}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
          >
            <div className="p-2 rounded-full bg-slate-800 group-hover:bg-slate-700 transition-colors">
              <ArrowLeft size={20} />
            </div>
            <span className="hidden md:inline font-medium">Leave Game</span>
          </button>

          <div className="flex items-center gap-4 bg-slate-800/50 backdrop-blur rounded-full px-6 py-2 border border-slate-700">
            {gameMode === 'bot' ? <Cpu size={18} className="text-purple-400" /> : <Users size={18} className="text-blue-400" />}
            <span className="font-mono text-sm tracking-wider text-slate-300">
              {gameMode === 'bot' ? 'TRAINING MODE' : 'RANKED MATCH'}
            </span>
          </div>

          <div className="flex items-center gap-3">
             <div className="hidden md:flex flex-col items-end">
                <span className="text-xs text-slate-400 uppercase tracking-widest">Player</span>
                <span className="font-bold text-game-accent">{username}</span>
             </div>
             <div className={cn(
               "w-10 h-10 rounded-full border-2 shadow-lg flex items-center justify-center font-bold text-slate-900",
               myColor === 'red' ? "bg-game-red border-red-400" : myColor === 'yellow' ? "bg-game-yellow border-yellow-400" : "bg-slate-700 border-slate-500"
             )}>
               {username?.charAt(0).toUpperCase()}
             </div>
          </div>
        </div>

        {/* Main Game Area */}
        <div className="flex-1 flex flex-col items-center justify-center gap-8 md:gap-12 w-full">
           {/* Game Status/Notification Area */}
           <div className="h-20 flex items-center justify-center w-full">
              {getStatusContent()}
           </div>

           {/* The Board */}
           <div className="transform transition-all duration-500 hover:scale-[1.01]">
              <GameBoard 
                gameState={gameState} 
                onColumnClick={makeMove} 
                isMyTurn={isMyTurn} 
              />
           </div>

           {/* Mobile Turn Indicator (Bottom) */}
           <div className="md:hidden w-full px-4 text-center text-sm text-slate-500 pb-4">
              {isMyTurn ? "Tap a column to drop your piece" : "Waiting for opponent..."}
           </div>
        </div>
      </div>
    </div>
  );
};
