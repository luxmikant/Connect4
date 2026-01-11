import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { usePlayer } from '../hooks/usePlayer';

export const Lobby: React.FC = () => {
  const [name, setName] = useState('');
  const [gameMode, setGameMode] = useState<'matchmaking' | 'bot'>('matchmaking');
  const { setUsername } = usePlayer();
  const navigate = useNavigate();

  const handleJoin = () => {
    if (name.trim()) {
      setUsername(name.trim());
      // Store game mode preference
      localStorage.setItem('connect4_gameMode', gameMode);
      navigate('/game');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && name.trim()) {
      handleJoin();
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
          onKeyPress={handleKeyPress}
          placeholder="Enter your name"
          className="w-full px-4 py-2 mb-4 text-lg text-gray-900 bg-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        
        {/* Game Mode Selection */}
        <div className="mb-4 space-y-2">
          <label className="text-sm text-gray-400 block mb-2">Select Game Mode:</label>
          <div className="flex gap-2">
            <button
              onClick={() => setGameMode('matchmaking')}
              className={`flex-1 px-4 py-2 rounded-lg font-semibold transition-colors ${
                gameMode === 'matchmaking'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              🎮 Find Player
            </button>
            <button
              onClick={() => setGameMode('bot')}
              className={`flex-1 px-4 py-2 rounded-lg font-semibold transition-colors ${
                gameMode === 'bot'
                  ? 'bg-green-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              🤖 Play vs Bot
            </button>
          </div>
        </div>
        
        <p className="text-sm text-gray-400 mb-4 text-center">
          {gameMode === 'matchmaking' 
            ? 'Wait for another player (auto-matches with bot after 10s)'
            : 'Play against our AI bot immediately'}
        </p>
        
        <button
          onClick={handleJoin}
          disabled={!name.trim()}
          className="w-full px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded-lg font-semibold transition-colors"
        >
          {gameMode === 'matchmaking' ? 'Join Game' : 'Play vs Bot'}
        </button>
        
        <Link 
          to="/leaderboard" 
          className="block w-full text-center mt-4 px-6 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors"
        >
          🏆 View Leaderboard
        </Link>
      </div>
    </div>
  );
};
