import { supabase } from '../lib/supabase';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export interface Player {
  id: string;
  username: string;
  auth_user_id?: string;
  is_guest: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface PlayerStats {
  username: string;
  games_played: number;
  wins: number;
  losses: number;
  draws: number;
  win_rate: number;
  total_moves: number;
  avg_moves_per_game: number;
  created_at: string;
  updated_at: string;
}

/**
 * Get or create a player for the authenticated user
 * This will link the Supabase auth user to a player in the game system
 */
export const getOrCreatePlayer = async (): Promise<Player> => {
  const { data: { session } } = await supabase.auth.getSession();
  
  if (!session) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE_URL}/api/v1/auth/player`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${session.access_token}`,
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get or create player');
  }

  return response.json();
};

/**
 * Get player stats by username
 */
export const getPlayerStats = async (username: string): Promise<PlayerStats | null> => {
  const { data: { session } } = await supabase.auth.getSession();
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  
  if (session) {
    headers['Authorization'] = `Bearer ${session.access_token}`;
  }

  const response = await fetch(`${API_BASE_URL}/api/v1/leaderboard/${username}`, {
    method: 'GET',
    headers,
  });

  if (response.status === 404) {
    return null;
  }

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to fetch player stats');
  }

  return response.json();
};

/**
 * Get player by auth user ID
 */
export const getPlayerByAuthUserId = async (authUserId: string): Promise<Player | null> => {
  const { data: { session } } = await supabase.auth.getSession();
  
  if (!session) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE_URL}/api/v1/players/auth/${authUserId}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${session.access_token}`,
    },
  });

  if (response.status === 404) {
    return null;
  }

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to fetch player');
  }

  return response.json();
};
