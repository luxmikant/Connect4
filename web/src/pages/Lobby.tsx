import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { User, Cpu, Globe, ArrowRight, Trophy, Users, Link as LinkIcon } from 'lucide-react';
import { usePlayer } from '../hooks/usePlayer';
import { useAuth } from '../contexts/AuthContext';
import { getOrCreatePlayer } from '../services/playerService';
import { cn } from '../lib/utils';

export const Lobby: React.FC = () => {
    const [name, setName] = useState('');
    const [gameMode, setGameMode] = useState<'matchmaking' | 'bot' | 'custom'>('matchmaking');
    const [customRoomAction, setCustomRoomAction] = useState<'create' | 'join'>('create');
    const [roomCode, setRoomCode] = useState('');
    const [isCreatingPlayer, setIsCreatingPlayer] = useState(false);
    const { setUsername } = usePlayer();
    const { user, profile } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        if (profile?.username) {
            setName(profile.username);
        }
    }, [profile]);

    const handleJoin = async () => {
        if (!name.trim()) return;

        // Validate room code if joining custom room
        if (gameMode === 'custom' && customRoomAction === 'join' && !roomCode.trim()) {
            return;
        }

        setIsCreatingPlayer(true);
        
        try {
            // If user is authenticated, link their player account
            if (user && profile?.username) {
                await getOrCreatePlayer();
            }
            
            setUsername(name.trim());
            localStorage.setItem('connect4_gameMode', gameMode);
            
            // For custom room join, navigate with room code as query param
            if (gameMode === 'custom' && customRoomAction === 'join' && roomCode.trim()) {
                localStorage.setItem('connect4_customRoomAction', 'join');
                navigate(`/game?room=${roomCode.toUpperCase().trim()}`);
            } else if (gameMode === 'custom' && customRoomAction === 'create') {
                localStorage.setItem('connect4_customRoomAction', 'create');
                navigate('/game');
            } else {
                navigate('/game');
            }
        } catch (error) {
            console.error('Failed to create player:', error);
            // Continue anyway for guest users
            setUsername(name.trim());
            localStorage.setItem('connect4_gameMode', gameMode);
            
            if (gameMode === 'custom' && customRoomAction === 'join' && roomCode.trim()) {
                localStorage.setItem('connect4_customRoomAction', 'join');
                navigate(`/game?room=${roomCode.toUpperCase().trim()}`);
            } else if (gameMode === 'custom' && customRoomAction === 'create') {
                localStorage.setItem('connect4_customRoomAction', 'create');
                navigate('/game');
            } else {
                navigate('/game');
            }
        } finally {
            setIsCreatingPlayer(false);
        }
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && name.trim()) {
            handleJoin();
        }
    };

    return (
        <div className="min-h-screen bg-game-bg flex items-center justify-center relative overflow-hidden p-4">
            {/* Background Effects */}
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-600/10 rounded-full blur-[100px] animate-pulse-glow" />
                <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-600/10 rounded-full blur-[100px] animate-pulse-glow delay-700" />
            </div>

            <motion.div 
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                className="w-full max-w-md relative z-10"
            >
                {/* Card Container */}
                <div className="bg-slate-900/50 backdrop-blur-xl border border-slate-700/50 rounded-3xl p-8 shadow-2xl">
                    
                    <div className="text-center mb-8">
                        <h2 className="text-3xl font-black text-white mb-2 tracking-tight">PLAYER SETUP</h2>
                        <p className="text-slate-400 text-sm">Configure your battle identity</p>
                    </div>

                    <div className="space-y-6">
                        
                        {/* Name Input */}
                        <div className="space-y-2">
                            <label className="text-xs font-mono text-cyan-400 uppercase tracking-widest pl-1">Codename</label>
                            <div className="relative group">
                                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-cyan-400 transition-colors" />
                                <input
                                    type="text"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    onKeyPress={handleKeyPress}
                                    placeholder="ENTER USERNAME..."
                                    className={cn(
                                        "w-full bg-slate-800/50 border border-slate-700 text-white pl-12 pr-4 py-4 rounded-xl focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition-all font-mono placeholder:text-slate-600",
                                        profile?.username && "opacity-75 cursor-not-allowed bg-slate-800/80"
                                    )}
                                    autoFocus
                                    readOnly={!!profile?.username}
                                />
                            </div>
                            {profile?.username && (
                                <p className="text-xs text-slate-400 pl-1">Logged in as {profile.username}</p>
                            )}
                        </div>

                        {/* Mode Selection */}
                        <div className="space-y-2">
                            <label className="text-xs font-mono text-cyan-400 uppercase tracking-widest pl-1">Battle Mode</label>
                            <div className="grid grid-cols-3 gap-3">
                                <button
                                    onClick={() => setGameMode('matchmaking')}
                                    className={cn(
                                        "relative flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-300 gap-2 overflow-hidden",
                                        gameMode === 'matchmaking' 
                                            ? "bg-blue-600/20 border-blue-500 text-white shadow-[0_0_20px_rgba(37,99,235,0.2)]" 
                                            : "bg-slate-800/30 border-slate-700 text-slate-400 hover:bg-slate-800"
                                    )}
                                >
                                    <Globe className={cn("w-6 h-6", gameMode === 'matchmaking' ? "text-blue-400" : "text-slate-500")} />
                                    <span className="text-xs font-bold">ONLINE</span>
                                    {gameMode === 'matchmaking' && (
                                        <motion.div layoutId="active-ring" className="absolute inset-0 border-2 border-blue-500 rounded-xl" />
                                    )}
                                </button>

                                <button
                                    onClick={() => setGameMode('bot')}
                                    className={cn(
                                        "relative flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-300 gap-2 overflow-hidden",
                                        gameMode === 'bot' 
                                            ? "bg-emerald-600/20 border-emerald-500 text-white shadow-[0_0_20px_rgba(16,185,129,0.2)]" 
                                            : "bg-slate-800/30 border-slate-700 text-slate-400 hover:bg-slate-800"
                                    )}
                                >
                                    <Cpu className={cn("w-6 h-6", gameMode === 'bot' ? "text-emerald-400" : "text-slate-500")} />
                                    <span className="text-xs font-bold">VS AI</span>
                                    {gameMode === 'bot' && (
                                        <motion.div layoutId="active-ring" className="absolute inset-0 border-2 border-emerald-500 rounded-xl" />
                                    )}
                                </button>

                                <button
                                    onClick={() => setGameMode('custom')}
                                    className={cn(
                                        "relative flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-300 gap-2 overflow-hidden",
                                        gameMode === 'custom' 
                                            ? "bg-purple-600/20 border-purple-500 text-white shadow-[0_0_20px_rgba(168,85,247,0.2)]" 
                                            : "bg-slate-800/30 border-slate-700 text-slate-400 hover:bg-slate-800"
                                    )}
                                >
                                    <Users className={cn("w-6 h-6", gameMode === 'custom' ? "text-purple-400" : "text-slate-500")} />
                                    <span className="text-xs font-bold">FRIEND</span>
                                    {gameMode === 'custom' && (
                                        <motion.div layoutId="active-ring" className="absolute inset-0 border-2 border-purple-500 rounded-xl" />
                                    )}
                                </button>
                            </div>
                        </div>

                        {/* Custom Room Options */}
                        <AnimatePresence>
                            {gameMode === 'custom' && (
                                <motion.div
                                    initial={{ opacity: 0, height: 0 }}
                                    animate={{ opacity: 1, height: 'auto' }}
                                    exit={{ opacity: 0, height: 0 }}
                                    className="space-y-3 overflow-hidden"
                                >
                                    <div className="grid grid-cols-2 gap-2">
                                        <button
                                            onClick={() => setCustomRoomAction('create')}
                                            className={cn(
                                                "py-2 px-3 rounded-lg text-sm font-medium transition-all",
                                                customRoomAction === 'create'
                                                    ? "bg-purple-600 text-white"
                                                    : "bg-slate-800 text-slate-400 hover:bg-slate-700"
                                            )}
                                        >
                                            Create Room
                                        </button>
                                        <button
                                            onClick={() => setCustomRoomAction('join')}
                                            className={cn(
                                                "py-2 px-3 rounded-lg text-sm font-medium transition-all",
                                                customRoomAction === 'join'
                                                    ? "bg-purple-600 text-white"
                                                    : "bg-slate-800 text-slate-400 hover:bg-slate-700"
                                            )}
                                        >
                                            Join Room
                                        </button>
                                    </div>
                                    
                                    {customRoomAction === 'join' && (
                                        <motion.div
                                            initial={{ opacity: 0 }}
                                            animate={{ opacity: 1 }}
                                            className="relative group"
                                        >
                                            <LinkIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-purple-400 transition-colors" />
                                            <input
                                                type="text"
                                                value={roomCode}
                                                onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
                                                onKeyPress={handleKeyPress}
                                                placeholder="ENTER ROOM CODE..."
                                                className="w-full bg-slate-800/50 border border-slate-700 text-white pl-12 pr-4 py-3 rounded-xl focus:outline-none focus:border-purple-500 focus:ring-1 focus:ring-purple-500 transition-all font-mono placeholder:text-slate-600 uppercase"
                                                maxLength={8}
                                            />
                                        </motion.div>
                                    )}
                                </motion.div>
                            )}
                        </AnimatePresence>

                        {/* Description */}
                        <AnimatePresence mode="wait">
                            <motion.p
                                key={gameMode}
                                initial={{ opacity: 0, y: 5 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -5 }}
                                className="text-center text-slate-400 text-xs h-4"
                            >
                                {gameMode === 'matchmaking' 
                                    ? 'Global matchmaking system. Auto-fallback to Bot after 10s.' 
                                    : gameMode === 'bot'
                                    ? 'Training simulation against Minimax algorithm.'
                                    : customRoomAction === 'create'
                                    ? 'Create a private room and share the code with a friend.'
                                    : 'Enter your friend\'s room code to join.'}
                            </motion.p>
                        </AnimatePresence>

                        {/* Action Buttons */}
                        <div className="pt-4 space-y-3">
                            <motion.button
                                whileHover={{ scale: 1.02 }}
                                whileTap={{ scale: 0.98 }}
                                onClick={handleJoin}
                                disabled={!name.trim() || isCreatingPlayer || (gameMode === 'custom' && customRoomAction === 'join' && !roomCode.trim())}
                                className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white font-bold py-4 rounded-xl shadow-lg disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 group transition-all"
                            >
                                {isCreatingPlayer ? (
                                    <>
                                        <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                                        Initializing...
                                    </>
                                ) : (
                                    <>
                                        {gameMode === 'custom' && customRoomAction === 'create' 
                                            ? 'CREATE ROOM'
                                            : gameMode === 'custom' && customRoomAction === 'join'
                                            ? 'JOIN ROOM'
                                            : 'JOIN BATTLE'}
                                        <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                                    </>
                                )}
                            </motion.button>

                            <button
                                onClick={() => navigate('/leaderboard')}
                                className="w-full py-3 text-slate-400 hover:text-white text-sm font-medium flex items-center justify-center gap-2 transition-colors"
                            >
                                <Trophy className="w-4 h-4" />
                                <span>View Global Rankings</span>
                            </button>
                        </div>
                    </div>
                </div>
            </motion.div>
        </div>
    );
};
