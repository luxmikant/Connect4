import { useState, useEffect } from 'react';

export interface PlayerStat {
  rank: number;
  player_id: string;
  username: string;
  games_played: number;
  games_won: number;
  win_rate: number;
}

export const useLeaderboard = () => {
  const [leaderboard, setLeaderboard] = useState<PlayerStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLeaderboard = async () => {
      try {
        setLoading(true);
        const response = await fetch('http://localhost:8080/api/v1/leaderboard');
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
