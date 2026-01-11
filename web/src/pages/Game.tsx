import React from 'react';
import { Toaster } from 'react-hot-toast';
import { GameBoard } from '../components/GameBoard';
import { useGame } from '../hooks/useGame';

export const Game: React.FC = () => {
  const { gameState, isMyTurn, makeMove } = useGame();

  const getStatusMessage = () => {
    if (!gameState) return "Connecting to server...";
    if (gameState.status === 'waiting') return "Waiting for another player...";
    if (gameState.status === 'in_progress') {
      return isMyTurn ? "Your turn" : "Opponent's turn";
    }
    if (gameState.status === 'completed') {
      if (gameState.winner) {
        const winnerName = gameState.players[gameState.winner];
        return `${winnerName} has won!`;
      }
      return "The game is a draw!";
    }
    return "Game Over";
  };

  return (
    <>
      <Toaster position="top-center" reverseOrder={false} />
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white p-4">
        <h1 className="text-3xl font-bold mb-2">Connect 4</h1>
        <p className="text-xl mb-4 h-6">{getStatusMessage()}</p>
        <GameBoard gameState={gameState} onColumnClick={makeMove} isMyTurn={isMyTurn} />
      </div>
    </>
  );
};
