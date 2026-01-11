import { useState, useEffect } from 'react';

export interface PlayerStat {
  id: string;
  username: string;
  gamesPlayed: number;
  gamesWon: number;
  winRate: number;
  avgGameTime: number;
  lastPlayed: string;
  createdAt: string;
  updatedAt: string;
}

export const useLeaderboard = () => {
  const [leaderboard, setLeaderboard] = useState<PlayerStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLeaderboard = async () => {
      try {
        setLoading(true);
        const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
        const response = await fetch(`${apiUrl}/api/v1/leaderboard`);
        if (!response.ok) {
          throw new Error('Failed to fetch leaderboard');
        }
        const data = await response.json();
        setLeaderboard(data);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchLeaderboard();
  }, []);

  return { leaderboard, loading, error };
};
