import React from 'react';
import { Disc } from './Disc';
import type { GameState } from '../types/websocket';
import { motion } from 'framer-motion';

interface GameBoardProps {
  gameState: GameState | null;
  onColumnClick: (column: number) => void;
  isMyTurn: boolean;
}

export const GameBoard: React.FC<GameBoardProps> = ({ gameState, onColumnClick, isMyTurn }) => {
  const grid: number[][] = gameState?.board || Array(6).fill(null).map(() => Array(7).fill(0));

  return (
    <div className="relative bg-blue-800 p-2 rounded-lg shadow-lg">
      {/* Board Cells */}
      <div className="grid grid-cols-7 gap-2">
        {Array.from({ length: 42 }).map((_, i) => (
          <div key={i} className="w-16 h-16 md:w-20 md:h-20 bg-gray-900 rounded-full flex items-center justify-center">
          </div>
        ))}
      </div>
      
      {/* Discs */}
      <div className="absolute top-0 left-0 w-full h-full p-2 grid grid-cols-7 gap-2">
        {grid.slice().reverse().map((row, rowIndex) => 
          row.map((cell, colIndex) => (
            <div key={`${5-rowIndex}-${colIndex}`} className="w-full h-full flex items-center justify-center">
              <Disc color={cell} rowIndex={5-rowIndex} />
            </div>
          ))
        )}
      </div>

      {/* Clickable Columns */}
      <div className="absolute top-0 left-0 w-full h-full grid grid-cols-7">
        {Array.from({ length: 7 }).map((_, colIndex) => (
          <motion.div
            key={colIndex}
            onClick={() => onColumnClick(colIndex)}
            className="h-full cursor-pointer"
            whileHover={isMyTurn ? { backgroundColor: "rgba(255, 255, 255, 0.1)" } : {}}
          />
        ))}
      </div>
    </div>
  );
};
