import React, { useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '../lib/utils';
import type { GameState } from '../types/websocket';
import { useGameSound } from '../hooks/useGameSound';

interface GameBoardProps {
  gameState: GameState | null;
  onColumnClick: (column: number) => void;
  isMyTurn: boolean;
}

export const GameBoard: React.FC<GameBoardProps> = ({ gameState, onColumnClick, isMyTurn }) => {
  const { playSound } = useGameSound();

  // Safe grid parsing
  const grid = useMemo(() => {
    if (!gameState) return Array(6).fill(Array(7).fill(0));
    const boardData = gameState.board;
    // Ensure we always have a valid grid
    if (Array.isArray(boardData)) return boardData;
    return (boardData as any)?.cells || Array(6).fill(Array(7).fill(0));
  }, [gameState]);

  // Visual grid needs row 0 at TOP.
  const visualGrid = useMemo(() => [...grid].reverse(), [grid]);

  const handleColumnClick = (colIndex: number) => {
    if (isMyTurn) {
      playSound('drop');
      onColumnClick(colIndex);
    }
  };

  return (
    <div className="relative p-4 md:p-8 rounded-3xl bg-white/10 backdrop-blur-md border border-white/20 shadow-2xl">
      {/* Game Board Container */}
      <div className="relative bg-game-board rounded-lg p-3 shadow-inner border-4 border-slate-700">
        
        {/* The Grid of Slots */}
        <div className="grid grid-cols-7 gap-2 md:gap-3 relative z-10">
          {/* Generate columns for click handling */}
          {Array.from({ length: 7 }).map((_, colIndex) => (
            <div
              key={`col-${colIndex}`}
              className="flex flex-col gap-2 md:gap-3 group cursor-pointer"
              onClick={() => handleColumnClick(colIndex)}
            >
              {/* Hover Indicator (The "Ghost" Disc) */}
              <div className={cn(
                "h-2 w-full rounded-full transition-all duration-300 opacity-0 transform translate-y-2",
                isMyTurn && "group-hover:opacity-100 group-hover:translate-y-0",
                isMyTurn && "bg-white/50 shadow-[0_0_10px_rgba(255,255,255,0.5)]"
              )} />

              {/* Rows within the column */}
              {visualGrid.map((row, rowIndex) => {
                const cellValue = row[colIndex]; // Get value at this specific row/col
                const isRed = cellValue === 1 || cellValue === "red";
                const isYellow = cellValue === 2 || cellValue === "yellow";

                return (
                  <div 
                    key={`${rowIndex}-${colIndex}`} 
                    className="relative w-8 h-8 md:w-12 md:h-12 lg:w-16 lg:h-16 rounded-full bg-game-slot shadow-inner overflow-hidden"
                  >
                    <AnimatePresence mode='popLayout'>
                      {(isRed || isYellow) && (
                        <motion.div
                          initial={{ y: -300, opacity: 0 }}
                          animate={{ y: 0, opacity: 1 }}
                          exit={{ scale: 0, opacity: 0 }}
                          transition={{ 
                            type: "spring", 
                            stiffness: 400, 
                            damping: 25,
                            mass: 1 
                          }}
                          className={cn(
                            "w-full h-full rounded-full shadow-[inset_-2px_-2px_6px_rgba(0,0,0,0.3)]",
                            isRed ? "bg-game-red" : "bg-game-yellow",
                            // Add a shine effect
                            "after:content-[''] after:absolute after:top-2 after:left-2 after:w-1/3 after:h-1/3 after:bg-white/30 after:rounded-full after:blur-[1px]"
                          )}
                        />
                      )}
                    </AnimatePresence>
                  </div>
                );
              })}
            </div>
          ))}
        </div>

        {/* Board Legs (Purely decorative) */}
        <div className="absolute -bottom-6 -left-2 w-4 h-24 bg-slate-800 rounded-b-lg -z-10 transform -rotate-6" />
        <div className="absolute -bottom-6 -right-2 w-4 h-24 bg-slate-800 rounded-b-lg -z-10 transform rotate-6" />
      </div>

      {/* Status Glow */}
      {isMyTurn && (
        <div className="absolute -inset-1 bg-gradient-to-r from-green-400 to-blue-500 rounded-3xl blur opacity-20 animate-pulse -z-10 pointer-events-none" />
      )}
    </div>
  );
};
