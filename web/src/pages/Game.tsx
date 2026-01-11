import React, { useEffect, useRef } from 'react';
import { Toaster } from 'react-hot-toast';
import { GameBoard } from '../components/GameBoard';
import { useGame } from '../hooks/useGame';
import { useNavigate } from 'react-router-dom';
import { wsService } from '../services/websocket';
import { MessageType } from '../types/websocket';
import { usePlayer } from '../hooks/usePlayer';

export const Game: React.FC = () => {
  const { gameState, queueStatus, isMyTurn, myColor, makeMove, leaveQueue } = useGame();
  const { username } = usePlayer();
  const navigate = useNavigate();
  const hasJoinedRef = useRef(false);

  // Get game mode from localStorage
  const gameMode = localStorage.getItem('connect4_gameMode') || 'matchmaking';

  // Join queue/bot game once WebSocket is connected
  useEffect(() => {
    if (username && !hasJoinedRef.current) {
      const checkAndJoin = () => {
        if (wsService.isConnected() && !hasJoinedRef.current) {
          hasJoinedRef.current = true;
          if (gameMode === 'bot') {
            console.log("Requesting bot game for:", username);
            wsService.send(MessageType.PlayWithBot, { username });
          } else {
            console.log("Joining matchmaking queue for:", username);
            wsService.send(MessageType.JoinQueue, { username });
          }
        } else if (!wsService.isConnected()) {
          // Retry after a short delay
          setTimeout(checkAndJoin, 200);
        }
      };
      checkAndJoin();
    }
  }, [username, gameMode]);

  // Cleanup when leaving
  useEffect(() => {
    return () => {
      if (queueStatus?.inQueue) {
        leaveQueue();
      }
    };
  }, [queueStatus, leaveQueue]);

  const getStatusMessage = () => {
    // Show queue status while waiting
    if (queueStatus?.inQueue) {
      return gameMode === 'bot' 
        ? 'Starting bot game...'
        : `Waiting for opponent... (${queueStatus.timeRemaining} remaining)`;
    }
    
    if (!gameState) {
      return gameMode === 'bot' ? "Starting bot game..." : "Connecting to server...";
    }
    
    if (gameState.status === 'waiting') return "Waiting for another player...";
    
    if (gameState.status === 'in_progress') {
      return isMyTurn ? "🎯 Your turn!" : "⏳ Opponent's turn...";
    }
    
    if (gameState.status === 'completed') {
      if (gameState.winner) {
        const winnerName = gameState.players[gameState.winner];
        return winnerName === username ? "🎉 You won!" : `${winnerName} has won!`;
      }
      return "🤝 The game is a draw!";
    }
    
    return "Game Over";
  };

  const getColorIndicator = () => {
    if (!myColor) return null;
    const colorClass = myColor === 'red' ? 'bg-red-500' : 'bg-yellow-400';
    return (
      <div className="flex items-center gap-2 mb-2">
        <span className="text-gray-400">You are:</span>
        <div className={`w-6 h-6 rounded-full ${colorClass} shadow-lg`}></div>
        <span className="capitalize font-semibold">{myColor}</span>
      </div>
    );
  };

  const handlePlayAgain = () => {
    hasJoinedRef.current = false;
    navigate('/lobby');
  };

  const handleBackToLobby = () => {
    if (queueStatus?.inQueue) {
      leaveQueue();
    }
    hasJoinedRef.current = false;
    navigate('/lobby');
  };

  return (
    <>
      <Toaster position="top-center" reverseOrder={false} />
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white p-4">
        <h1 className="text-3xl font-bold mb-2">Connect 4</h1>
        
        {/* Show opponent info */}
        {gameState && gameState.isBot && (
          <p className="text-sm text-green-400 mb-1">🤖 Playing against Bot</p>
        )}
        
        {/* Show username */}
        {username && (
          <p className="text-sm text-gray-400 mb-1">Playing as: <span className="text-white font-semibold">{username}</span></p>
        )}
        
        {getColorIndicator()}
        
        <p className={`text-xl mb-4 h-6 ${isMyTurn ? 'text-green-400 font-bold' : ''}`}>
          {getStatusMessage()}
        </p>
        
        {/* Show queue countdown */}
        {queueStatus?.inQueue && gameMode !== 'bot' && (
          <div className="mb-4 text-center">
            <div className="w-64 bg-gray-700 rounded-full h-2 mb-2">
              <div 
                className="bg-blue-500 h-2 rounded-full transition-all duration-1000"
                style={{ 
                  width: `${Math.max(0, (parseInt(queueStatus.timeRemaining) / 10) * 100)}%` 
                }}
              ></div>
            </div>
            <p className="text-sm text-gray-400">
              If no player joins, you'll play against a bot
            </p>
            <button
              onClick={handleBackToLobby}
              className="mt-2 px-4 py-1 text-sm bg-gray-700 hover:bg-gray-600 rounded transition-colors"
            >
              Cancel
            </button>
          </div>
        )}

        {/* Loading spinner while waiting for game */}
        {!gameState && !queueStatus?.inQueue && (
          <div className="mb-4">
            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500"></div>
          </div>
        )}
        
        {/* Game board */}
        {gameState && !queueStatus?.inQueue && (
          <GameBoard 
            gameState={gameState} 
            onColumnClick={makeMove} 
            isMyTurn={isMyTurn} 
          />
        )}
        
        {/* Play again button */}
        {gameState?.status === 'completed' && (
          <div className="flex gap-4 mt-6">
            <button
              onClick={() => navigate('/')}
              className="px-6 py-3 bg-green-600 hover:bg-green-700 rounded-lg font-semibold transition-colors"
            >
              🏠 Home
            </button>
            <button
              onClick={handlePlayAgain}
              className="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
            >
              Play Again
            </button>
            <button
              onClick={() => navigate('/leaderboard')}
              className="px-6 py-3 bg-gray-700 hover:bg-gray-600 rounded-lg font-semibold transition-colors"
            >
              🏆 Leaderboard
            </button>
          </div>
        )}
      </div>
    </>
  );
};
