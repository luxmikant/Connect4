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
    <div className="relative rounded-[28px] p-[1px] bg-gradient-to-br from-fuchsia-400/40 via-violet-400/20 to-emerald-400/35 shadow-[0_25px_70px_rgba(2,6,23,0.55)]">
      <div className="relative rounded-[28px] bg-slate-950/85 border border-white/10 backdrop-blur-xl p-4 md:p-6 overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-20 -left-10 w-44 h-44 bg-fuchsia-500/20 blur-3xl rounded-full" />
          <div className="absolute -bottom-20 -right-10 w-44 h-44 bg-purple-500/20 blur-3xl rounded-full" />
        </div>

        <div className="relative z-10 bg-gradient-to-b from-slate-700/90 via-slate-800/95 to-slate-900 rounded-2xl p-3 md:p-4 border border-slate-600/60">
          <div className="grid grid-cols-7 gap-2 md:gap-3 mb-2 md:mb-3">
            {Array.from({ length: 7 }).map((_, colIndex) => (
              <button
                key={`top-${colIndex}`}
                type="button"
                onClick={() => handleColumnClick(colIndex)}
                className={cn(
                  "h-7 md:h-8 rounded-md text-xs font-bold transition-all duration-200",
                  isMyTurn
                    ? "bg-fuchsia-500/15 text-fuchsia-200 hover:bg-fuchsia-500/30 hover:-translate-y-0.5"
                    : "bg-slate-800/60 text-slate-500 cursor-not-allowed"
                )}
                aria-label={`Drop disc in column ${colIndex + 1}`}
              >
                {colIndex + 1}
              </button>
            ))}
          </div>

          <div className="grid grid-cols-7 gap-2 md:gap-3 relative">
            {Array.from({ length: 7 }).map((_, colIndex) => (
              <div
                key={`col-${colIndex}`}
                className={cn(
                  "flex flex-col gap-2 md:gap-3 rounded-xl p-1 transition-colors",
                  isMyTurn ? "hover:bg-white/5 cursor-pointer" : "cursor-not-allowed"
                )}
                onClick={() => handleColumnClick(colIndex)}
              >
                {visualGrid.map((row, rowIndex) => {
                  const cellValue = row[colIndex];
                  const isRed = cellValue === 1 || cellValue === 'red';
                  const isYellow = cellValue === 2 || cellValue === 'yellow';

                  return (
                    <div
                      key={`${rowIndex}-${colIndex}`}
                      className="relative w-9 h-9 md:w-12 md:h-12 lg:w-16 lg:h-16 rounded-full bg-gradient-to-br from-slate-950 via-slate-900 to-slate-800 border border-slate-700/80 shadow-[inset_0_4px_10px_rgba(0,0,0,0.45)] overflow-hidden"
                    >
                      <AnimatePresence mode="popLayout">
                        {(isRed || isYellow) && (
                          <motion.div
                            initial={{ y: -320, opacity: 0 }}
                            animate={{ y: 0, opacity: 1 }}
                            exit={{ scale: 0, opacity: 0 }}
                            transition={{ type: 'spring', stiffness: 420, damping: 24, mass: 1 }}
                            className={cn(
                              "relative w-full h-full rounded-full border shadow-[inset_-2px_-2px_8px_rgba(0,0,0,0.35),0_6px_10px_rgba(0,0,0,0.35)]",
                              isRed
                                ? "bg-gradient-to-br from-rose-300 via-rose-500 to-rose-700 border-rose-200/60"
                                : "bg-gradient-to-br from-amber-100 via-amber-300 to-amber-500 border-amber-200/70"
                            )}
                          >
                            <div className="absolute top-[16%] left-[18%] w-[35%] h-[35%] rounded-full bg-white/35 blur-[1px]" />
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </div>

        {isMyTurn && (
          <div className="absolute -inset-1 rounded-[28px] bg-gradient-to-r from-fuchsia-500/25 via-violet-400/20 to-emerald-400/25 blur-xl -z-10" />
        )}
      </div>
    </div>
  );
};
