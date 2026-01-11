import React from 'react';
import { motion } from 'framer-motion';
import { clsx } from 'clsx';

interface DiscProps {
  color: number; // 0 = empty, 1 = red, 2 = yellow
  rowIndex: number;
}

export const Disc: React.FC<DiscProps> = ({ color, rowIndex }) => {
  if (color === 0) return null;

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
        color === 1 ? "bg-red-500" : "bg-yellow-400"
      )}
    />
  );
};
