import React from 'react';
import { useNavigate } from 'react-router-dom';

export const Home: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white">
      <h1 className="text-4xl font-bold mb-8">Connect 4 Multiplayer</h1>
      <div className="space-y-4">
        <button
          onClick={() => navigate('/lobby')}
          className="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
        >
          Play Game
        </button>
        <button
          onClick={() => navigate('/leaderboard')}
          className="px-6 py-3 bg-gray-700 hover:bg-gray-800 rounded-lg font-semibold transition-colors"
        >
          View Leaderboard
        </button>
      </div>
    </div>
  );
};
