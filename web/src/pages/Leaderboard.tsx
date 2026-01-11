import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, Trophy, Medal, Crown } from 'lucide-react';
import { useLeaderboard } from '../hooks/useLeaderboard';
import { cn } from '../lib/utils'; // Ensure you have this utility or use clsx/tailwind-merge directly

export const Leaderboard: React.FC = () => {
    const { leaderboard, loading, error } = useLeaderboard();
    const navigate = useNavigate();

    const getRankIcon = (index: number) => {
        switch (index) {
            case 0: return <Crown className="w-6 h-6 text-yellow-500" />;
            case 1: return <Medal className="w-6 h-6 text-slate-300" />;
            case 2: return <Medal className="w-6 h-6 text-amber-700" />;
            default: return <span className="font-mono text-slate-500">#{index + 1}</span>;
        }
    };

    const getRankRowStyle = (index: number) => {
        switch (index) {
            case 0: return "bg-yellow-500/10 border-yellow-500/50";
            case 1: return "bg-slate-300/10 border-slate-300/50";
            case 2: return "bg-amber-700/10 border-amber-700/50";
            default: return "bg-slate-800/30 border-slate-700 hover:bg-slate-800/50";
        }
    };

    return (
        <div className="min-h-screen bg-game-bg flex flex-col items-center py-12 px-4 relative overflow-hidden">
            {/* Ambient Background */}
            <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
                <div className="absolute top-[-20%] left-[20%] w-[60%] h-[60%] bg-purple-900/20 rounded-full blur-[120px] animate-pulse-glow" />
            </div>

            <div className="w-full max-w-4xl relative z-10">
                {/* Header */}
                <div className="flex items-center justify-between mb-8">
                    <button 
                        onClick={() => navigate('/')}
                        className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
                    >
                        <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                        <span>Back to Menu</span>
                    </button>
                    
                    <div className="flex items-center gap-3">
                        <Trophy className="w-8 h-8 text-yellow-500" />
                        <h1 className="text-3xl font-black text-white tracking-tight">GLOBAL RANKINGS</h1>
                    </div>
                </div>

                {/* Content */}
                {loading ? (
                    <div className="flex flex-col items-center justify-center h-64 gap-4">
                        <div className="w-12 h-12 border-4 border-slate-700 border-t-game-accent rounded-full animate-spin" />
                        <p className="text-slate-400 animate-pulse">Syncing Network Data...</p>
                    </div>
                ) : error ? (
                    <div className="p-8 bg-red-500/10 border border-red-500/50 rounded-xl text-center">
                        <p className="text-red-400">Failed to retrieve data securely.</p>
                        <p className="text-sm text-red-500/70 mt-2">{error}</p>
                    </div>
                ) : (
                    <div className="space-y-4">
                        {/* Table Header - Desktop Only */}
                        <div className="hidden md:grid grid-cols-12 gap-4 px-6 py-3 text-xs font-mono text-slate-500 uppercase tracking-widest">
                            <div className="col-span-2">Rank</div>
                            <div className="col-span-4">Player</div>
                            <div className="col-span-2 text-right">Matches</div>
                            <div className="col-span-2 text-right">Wins</div>
                            <div className="col-span-2 text-right">Win Rate</div>
                        </div>

                        {leaderboard.map((player, index) => (
                            <motion.div
                                initial={{ opacity: 0, x: -20 }}
                                animate={{ opacity: 1, x: 0 }}
                                transition={{ delay: index * 0.05 }}
                                key={player.id}
                                className={cn(
                                    "grid grid-cols-2 md:grid-cols-12 gap-4 items-center p-4 rounded-xl border backdrop-blur-sm transition-all relative overflow-hidden",
                                    getRankRowStyle(index)
                                )}
                            >
                                {/* Rank */}
                                <div className="col-span-1 md:col-span-2 flex items-center gap-3">
                                    {getRankIcon(index)}
                                </div>

                                {/* Player Info */}
                                <div className="col-span-1 md:col-span-4 font-bold text-white truncate text-right md:text-left">
                                    {player.username}
                                </div>

                                {/* Stats - Mobile: Hidden or Compressed */}
                                <div className="col-span-2 hidden md:block text-right text-slate-400 font-mono">
                                    {player.gamesPlayed}
                                </div>
                                <div className="col-span-2 hidden md:block text-right text-game-accent font-mono font-bold">
                                    {player.gamesWon}
                                </div>
                                <div className="col-span-2 hidden md:block text-right text-slate-300 font-mono">
                                    {(player.winRate * 100).toFixed(1)}%
                                </div>

                                {/* Mobile Only Stats View - If needed, could add a breakdown here */}
                            </motion.div>
                        ))}

                        {leaderboard.length === 0 && (
                            <div className="text-center py-20 text-slate-500">
                                <p>No records found in the database.</p>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};
