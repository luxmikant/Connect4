import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { User, Mail, Calendar, Trophy, TrendingUp, Save, LogOut, ArrowLeft } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { getPlayerStats, type PlayerStats } from '../services/playerService';

export const ProfilePage: React.FC = () => {
  const navigate = useNavigate();
  const { user, profile, updateProfile, signOut, loading } = useAuth();
  
  const [username, setUsername] = useState('');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState({ type: '', text: '' });
  const [stats, setStats] = useState<PlayerStats | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);

  useEffect(() => {
    if (!user && !loading) {
      navigate('/login');
    }
  }, [user, loading, navigate]);

  useEffect(() => {
    if (profile) {
      setUsername(profile.username || '');
      setAvatarUrl(profile.avatar_url || '');
      
      // Load player stats
      if (profile.username) {
        loadStats(profile.username);
      }
    }
  }, [profile]);

  const loadStats = async (username: string) => {
    setLoadingStats(true);
    try {
      const playerStats = await getPlayerStats(username);
      setStats(playerStats);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoadingStats(false);
    }
  };

  const handleSave = async () => {
    if (!username.trim()) {
      setMessage({ type: 'error', text: 'Username cannot be empty' });
      return;
    }

    setIsSaving(true);
    setMessage({ type: '', text: '' });

    const { error } = await updateProfile({
      username: username.trim(),
      avatar_url: avatarUrl.trim() || null,
    });

    setIsSaving(false);

    if (error) {
      setMessage({ type: 'error', text: error.message });
    } else {
      setMessage({ type: 'success', text: 'Profile updated successfully!' });
      setIsEditing(false);
      setTimeout(() => setMessage({ type: '', text: '' }), 3000);
    }
  };

  const handleSignOut = async () => {
    await signOut();
    navigate('/');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-game-bg flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-game-bg">
      {/* Background Effects */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-600/10 rounded-full blur-[100px] animate-pulse-glow" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-600/10 rounded-full blur-[100px] animate-pulse-glow delay-700" />
      </div>

      <div className="relative z-10 container mx-auto px-4 py-8 max-w-4xl">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <button
            onClick={() => navigate('/')}
            className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
            <span>Back to Home</span>
          </button>
          <button
            onClick={handleSignOut}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-slate-700 text-white hover:bg-slate-600 transition-colors"
          >
            <LogOut className="w-4 h-4" />
            Sign Out
          </button>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-6"
        >
          {/* Profile Card */}
          <div className="bg-slate-800/50 backdrop-blur-xl border border-slate-700 rounded-3xl p-8">
            <div className="flex items-start justify-between mb-6">
              <h1 className="text-3xl font-bold text-white">My Profile</h1>
              {!isEditing && (
                <button
                  onClick={() => setIsEditing(true)}
                  className="px-4 py-2 rounded-lg bg-game-accent text-white hover:bg-blue-600 transition-colors"
                >
                  Edit Profile
                </button>
              )}
            </div>

            {/* Avatar */}
            <div className="flex items-center gap-6 mb-8">
              <div className="w-24 h-24 rounded-full bg-gradient-to-br from-game-accent to-purple-600 flex items-center justify-center text-white text-3xl font-bold">
                {username.charAt(0).toUpperCase() || 'U'}
              </div>
              <div>
                <h2 className="text-2xl font-bold text-white">{username || 'Anonymous'}</h2>
                <p className="text-slate-400 flex items-center gap-2 mt-1">
                  <Mail className="w-4 h-4" />
                  {user.email}
                </p>
              </div>
            </div>

            {/* Edit Form */}
            {isEditing && (
              <div className="space-y-4 mb-6 p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                <div>
                  <label className="block text-sm text-slate-400 mb-2">Username</label>
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 text-white px-4 py-3 rounded-lg focus:outline-none focus:border-game-accent"
                    placeholder="Enter username"
                  />
                </div>
                <div>
                  <label className="block text-sm text-slate-400 mb-2">Avatar URL (optional)</label>
                  <input
                    type="text"
                    value={avatarUrl}
                    onChange={(e) => setAvatarUrl(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 text-white px-4 py-3 rounded-lg focus:outline-none focus:border-game-accent"
                    placeholder="https://example.com/avatar.png"
                  />
                </div>

                {message.text && (
                  <div
                    className={`p-3 rounded-lg ${
                      message.type === 'success'
                        ? 'bg-green-500/20 text-green-400 border border-green-500/50'
                        : 'bg-red-500/20 text-red-400 border border-red-500/50'
                    }`}
                  >
                    {message.text}
                  </div>
                )}

                <div className="flex gap-3">
                  <button
                    onClick={handleSave}
                    disabled={isSaving}
                    className="flex items-center gap-2 px-6 py-3 rounded-lg bg-game-accent text-white hover:bg-blue-600 transition-colors disabled:opacity-50"
                  >
                    <Save className="w-4 h-4" />
                    {isSaving ? 'Saving...' : 'Save Changes'}
                  </button>
                  <button
                    onClick={() => {
                      setIsEditing(false);
                      setUsername(profile?.username || '');
                      setAvatarUrl(profile?.avatar_url || '');
                      setMessage({ type: '', text: '' });
                    }}
                    className="px-6 py-3 rounded-lg bg-slate-700 text-white hover:bg-slate-600 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}

            {/* Account Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-center gap-3 p-4 bg-slate-900/50 rounded-xl border border-slate-700">
                <Calendar className="w-5 h-5 text-game-accent" />
                <div>
                  <p className="text-xs text-slate-400">Member Since</p>
                  <p className="text-white font-medium">
                    {new Date(profile?.created_at || Date.now()).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3 p-4 bg-slate-900/50 rounded-xl border border-slate-700">
                <User className="w-5 h-5 text-game-accent" />
                <div>
                  <p className="text-xs text-slate-400">User ID</p>
                  <p className="text-white font-medium text-sm">
                    {user.id.substring(0, 8)}...
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Stats Card */}
          <div className="bg-slate-800/50 backdrop-blur-xl border border-slate-700 rounded-3xl p-8">
            <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-2">
              <Trophy className="w-6 h-6 text-yellow-500" />
              Game Statistics
            </h2>
            
            {loadingStats ? (
              <div className="text-center text-slate-400 py-8">
                Loading stats...
              </div>
            ) : stats ? (
              <>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-white mb-2">{stats.games_played}</div>
                    <div className="text-sm text-slate-400">Games Played</div>
                  </div>
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-green-400 mb-2">{stats.wins}</div>
                    <div className="text-sm text-slate-400">Wins</div>
                  </div>
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-game-accent mb-2 flex items-center justify-center gap-2">
                      <TrendingUp className="w-6 h-6" />
                      {stats.win_rate.toFixed(1)}%
                    </div>
                    <div className="text-sm text-slate-400">Win Rate</div>
                  </div>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="text-center p-4 bg-slate-900/30 rounded-xl border border-slate-700/50">
                    <div className="text-2xl font-bold text-red-400 mb-1">{stats.losses}</div>
                    <div className="text-xs text-slate-500">Losses</div>
                  </div>
                  <div className="text-center p-4 bg-slate-900/30 rounded-xl border border-slate-700/50">
                    <div className="text-2xl font-bold text-yellow-400 mb-1">{stats.draws}</div>
                    <div className="text-xs text-slate-500">Draws</div>
                  </div>
                </div>
              </>
            ) : (
              <div className="text-center">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-white mb-2">0</div>
                    <div className="text-sm text-slate-400">Games Played</div>
                  </div>
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-green-400 mb-2">0</div>
                    <div className="text-sm text-slate-400">Wins</div>
                  </div>
                  <div className="text-center p-6 bg-slate-900/50 rounded-xl border border-slate-700">
                    <div className="text-3xl font-bold text-game-accent mb-2 flex items-center justify-center gap-2">
                      <TrendingUp className="w-6 h-6" />
                      0%
                    </div>
                    <div className="text-sm text-slate-400">Win Rate</div>
                  </div>
                </div>
                <p className="text-slate-500 text-sm">
                  Play your first game to see stats!
                </p>
              </div>
            )}
          </div>

          {/* Quick Actions */}
          <div className="flex gap-4">
            <button
              onClick={() => navigate('/lobby')}
              className="flex-1 py-4 rounded-xl bg-gradient-to-r from-game-accent to-purple-600 text-white font-bold hover:from-blue-600 hover:to-purple-700 transition-all"
            >
              Play Now
            </button>
            <button
              onClick={() => navigate('/leaderboard')}
              className="flex-1 py-4 rounded-xl bg-slate-700 text-white font-bold hover:bg-slate-600 transition-colors"
            >
              View Leaderboard
            </button>
          </div>
        </motion.div>
      </div>
    </div>
  );
};
