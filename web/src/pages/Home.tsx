import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Play, Trophy, Cpu, LogIn, UserPlus } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export const Home: React.FC = () => {
  const navigate = useNavigate();
  const { user, signOut } = useAuth();

  return (
    <div className="min-h-screen bg-game-bg flex flex-col relative overflow-hidden">
      {/* Navigation Bar */}
      <nav className="relative z-20 w-full px-6 py-4">
        <div className="container mx-auto flex justify-between items-center">
          <div className="text-2xl font-black text-white">CONNECT 4</div>
          <div className="flex gap-3">
            {user ? (
              <>
                <span className="text-slate-300 px-4 py-2">{user.email}</span>
                <button
                  onClick={() => signOut()}
                  className="px-4 py-2 rounded-lg bg-slate-700 text-white hover:bg-slate-600 transition-colors"
                >
                  Sign Out
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={() => navigate('/login')}
                  className="px-4 py-2 rounded-lg bg-slate-700 text-white hover:bg-slate-600 transition-colors flex items-center gap-2"
                >
                  <LogIn className="w-4 h-4" />
                  Login
                </button>
                <button
                  onClick={() => navigate('/register')}
                  className="px-4 py-2 rounded-lg bg-game-accent text-white hover:bg-blue-600 transition-colors flex items-center gap-2"
                >
                  <UserPlus className="w-4 h-4" />
                  Register
                </button>
              </>
            )}
          </div>
        </div>
      </nav>

      {/* Dynamic Background */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-[20%] -left-[10%] w-[60%] h-[60%] bg-purple-600/20 rounded-full blur-[120px] animate-pulse-glow" />
        <div className="absolute top-[40%] -right-[10%] w-[50%] h-[50%] bg-blue-600/20 rounded-full blur-[120px] animate-pulse-glow delay-1000" />
      </div>

      <div className="container mx-auto px-4 relative z-10 flex flex-col items-center justify-center flex-1">
        
        {/* Animated Headline */}
        <motion.div 
          initial={{ y: -50, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.8, type: "spring" }}
          className="text-center mb-16"
        >
          <h1 className="text-6xl md:text-8xl font-black tracking-tighter mb-4 bg-clip-text text-transparent bg-gradient-to-r from-game-accent to-purple-500 drop-shadow-2xl">
            CONNECT 4
          </h1>
          <p className="text-xl md:text-2xl text-slate-400 font-medium tracking-wide">
            MULTIPLAYER BATTLES
          </p>
        </motion.div>

        {/* Action Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-4xl">
           
           <motion.button
             whileHover={{ scale: 1.05 }}
             whileTap={{ scale: 0.95 }}
             onClick={() => navigate('/lobby')}
             className="group relative h-48 rounded-3xl bg-slate-800/50 backdrop-blur-md border border-slate-700 hover:border-game-accent hover:bg-slate-800 transition-all overflow-hidden flex flex-col items-center justify-center p-6 text-left"
           >
              <div className="absolute inset-0 bg-gradient-to-br from-game-accent/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="bg-game-accent p-4 rounded-full mb-4 shadow-[0_0_20px_rgba(59,130,246,0.5)]">
                 <Play className="w-8 h-8 text-white fill-current" />
              </div>
              <h2 className="text-2xl font-bold text-white mb-2">Play Now</h2>
              <p className="text-slate-400 text-center">Challenge players online or train against AI</p>
           </motion.button>

           <motion.button
             whileHover={{ scale: 1.05 }}
             whileTap={{ scale: 0.95 }}
             onClick={() => navigate('/leaderboard')}
             className="group relative h-48 rounded-3xl bg-slate-800/50 backdrop-blur-md border border-slate-700 hover:border-yellow-500 hover:bg-slate-800 transition-all overflow-hidden flex flex-col items-center justify-center p-6 text-left"
           >
              <div className="absolute inset-0 bg-gradient-to-br from-yellow-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="bg-yellow-500 p-4 rounded-full mb-4 shadow-[0_0_20px_rgba(234,179,8,0.5)]">
                 <Trophy className="w-8 h-8 text-black fill-current" />
              </div>
              <h2 className="text-2xl font-bold text-white mb-2">Leaderboard</h2>
              <p className="text-slate-400 text-center">See top players and global rankings</p>
           </motion.button>

        </div>

        {/* Footer Info */}
        <motion.div 
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="mt-16 flex items-center gap-2 text-slate-500 text-sm font-mono"
        >
          <Cpu className="w-4 h-4" />
          <span>POWERED BY GO + REACT + KAFKA</span>
        </motion.div>

      </div>
    </div>
  );
};
