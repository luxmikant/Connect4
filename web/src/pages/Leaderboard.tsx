import React from 'react';
import { useLeaderboard } from '../hooks/useLeaderboard';

export const Leaderboard: React.FC = () => {
  const { leaderboard, loading, error } = useLeaderboard();

  if (loading) {
    return <div className="text-center text-white">Loading...</div>;
  }

  if (error) {
    return <div className="text-center text-red-500">Error: {error}</div>;
  }

  return (
    <div className="flex flex-col items-center min-h-screen bg-gray-900 text-white p-4">
      <h1 className="text-3xl font-bold mb-4">Leaderboard</h1>
      <div className="w-full max-w-2xl">
        <table className="w-full text-left table-auto">
          <thead>
            <tr className="bg-gray-800">
              <th className="px-4 py-2">Rank</th>
              <th className="px-4 py-2">Player</th>
              <th className="px-4 py-2">Games Played</th>
              <th className="px-4 py-2">Games Won</th>
              <th className="px-4 py-2">Win Rate</th>
            </tr>
          </thead>
          <tbody>
            {leaderboard.map((player, index) => (
              <tr key={player.id} className="border-b border-gray-700">
                <td className="px-4 py-2">{index + 1}</td>
                <td className="px-4 py-2">{player.username}</td>
                <td className="px-4 py-2">{player.gamesPlayed}</td>
                <td className="px-4 py-2">{player.gamesWon}</td>
                <td className="px-4 py-2">{(player.winRate * 100).toFixed(2)}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
