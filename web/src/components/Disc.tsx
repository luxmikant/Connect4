import React from 'react';
import { motion } from 'framer-motion';
import { clsx } from 'clsx';

interface DiscProps {
  color: number | string; // 0 = empty, 1 or "red" = red, 2 or "yellow" = yellow
  rowIndex: number;
}

export const Disc: React.FC<DiscProps> = ({ color, rowIndex }) => {
  // Handle both numeric and string color values
  if (color === 0 || color === "" || color === null || color === undefined) return null;

  const isRed = color === 1 || color === "red";
  const isYellow = color === 2 || color === "yellow";

  if (!isRed && !isYellow) return null;

  return (
    <motion.div
      initial={{ y: -300, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ 
        type: "spring", 
        damping: 15, 
        stiffness: 100,
        delay: 0.05 * (5 - rowIndex) // Slight delay based on row for effect if multiple load
      }}
      className={clsx(
        "w-full h-full rounded-full shadow-inner",
        isRed ? "bg-red-500" : "bg-yellow-400"
      )}
    />
  );
};
