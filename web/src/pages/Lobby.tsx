import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { usePlayer } from '../hooks/usePlayer';

export const Lobby: React.FC = () => {
  const [name, setName] = useState('');
  const { setUsername } = usePlayer();
  const navigate = useNavigate();

  const handleJoin = () => {
    if (name.trim()) {
      setUsername(name.trim());
      navigate('/game');
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900 text-white">
      <h1 className="text-4xl font-bold mb-8">Connect 4</h1>
      <div className="w-full max-w-xs">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter your name"
          className="w-full px-4 py-2 mb-4 text-lg text-gray-900 bg-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          onClick={handleJoin}
          className="w-full px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
        >
          Join Game
        </button>
      </div>
    </div>
  );
};
