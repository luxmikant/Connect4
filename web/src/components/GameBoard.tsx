import React from 'react';
import type { GameState } from '../types/websocket';

interface GameBoardProps {
  gameState: GameState | null;
  onColumnClick: (column: number) => void;
  isMyTurn: boolean;
}

export const GameBoard: React.FC<GameBoardProps> = ({ gameState, onColumnClick, isMyTurn }) => {
  if (!gameState) {
    return <div style={{ color: 'white', padding: '20px' }}>Loading game...</div>;
  }

  // Handle both array board and object with cells property
  const boardData = gameState?.board;
  const grid: (number | string)[][] = Array.isArray(boardData) 
    ? boardData 
    : (boardData as any)?.cells || Array(6).fill(null).map(() => Array(7).fill(""));

  // Reverse the grid to render from bottom to top (Connect 4 style)
  const reversedGrid = [...grid].reverse();

  const getCellColor = (cell: number | string): string => {
    if (cell === 1 || cell === "red") return "#ef4444"; // red
    if (cell === 2 || cell === "yellow") return "#facc15"; // yellow
    return "#1f2937"; // empty (dark gray)
  };

  return (
    <div 
      style={{
        backgroundColor: '#2563eb',
        padding: '16px',
        borderRadius: '16px',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
        display: 'inline-block',
      }}
    >
      <div 
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(7, 1fr)',
          gap: '8px',
        }}
      >
        {reversedGrid.map((row, rowIndex) => 
          row.map((cell, colIndex) => (
            <div
              key={`${rowIndex}-${colIndex}`}
              onClick={() => isMyTurn && onColumnClick(colIndex)}
              style={{
                width: '60px',
                height: '60px',
                backgroundColor: getCellColor(cell),
                borderRadius: '50%',
                cursor: isMyTurn ? 'pointer' : 'not-allowed',
                border: '3px solid #1e40af',
                transition: 'transform 0.1s, box-shadow 0.1s',
                boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.3)',
              }}
              onMouseEnter={(e) => {
                if (isMyTurn) {
                  e.currentTarget.style.transform = 'scale(1.05)';
                  e.currentTarget.style.boxShadow = '0 0 10px rgba(255,255,255,0.5), inset 0 2px 4px rgba(0,0,0,0.3)';
                }
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'scale(1)';
                e.currentTarget.style.boxShadow = 'inset 0 2px 4px rgba(0,0,0,0.3)';
              }}
            />
          ))
        )}
      </div>
    </div>
  );
};
