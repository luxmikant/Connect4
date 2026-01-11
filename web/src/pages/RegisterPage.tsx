import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { motion } from 'framer-motion';

const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const { signUpWithEmail, signInWithGoogle, signInWithGitHub } = useAuth();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const validateUsername = (username: string): string | null => {
    if (username.length < 3 || username.length > 20) {
      return 'Username must be 3-20 characters';
    }
    if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      return 'Username can only contain letters, numbers, and underscores';
    }
    return null;
  };

  const handleEmailRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    const usernameError = validateUsername(username);
    if (usernameError) {
      setError(usernameError);
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setLoading(true);

    const { error } = await signUpWithEmail(email, password, username);
    
    if (error) {
      setError(error.message);
      setLoading(false);
    } else {
      // Show success message
      navigate('/verify-email');
    }
  };

  const handleGoogleRegister = async () => {
    setError('');
    const { error } = await signInWithGoogle();
    if (error) {
      setError(error.message);
    }
  };

  const handleGitHubRegister = async () => {
    setError('');
    const { error } = await signInWithGitHub();
    if (error) {
      setError(error.message);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 flex">
      {/* Left Side - Branding */}
      <motion.div 
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.6 }}
        className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-[#1a365d] to-[#2d4a7c] p-12 flex-col justify-between relative overflow-hidden"
      >
        {/* Decorative circles */}
        <div className="absolute top-20 right-20 w-64 h-64 bg-[#f59e0b]/10 rounded-full blur-3xl" />
        <div className="absolute bottom-20 left-20 w-96 h-96 bg-[#f59e0b]/5 rounded-full blur-3xl" />
        
        <div className="relative z-10">
          <Link to="/" className="inline-flex items-center gap-3 group">
            <div className="w-12 h-12 bg-white/10 backdrop-blur-sm rounded-xl flex items-center justify-center border border-white/20">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" className="text-[#f59e0b]">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="8" r="2" fill="currentColor"/>
                <circle cx="12" cy="16" r="2" fill="currentColor"/>
              </svg>
            </div>
            <span className="text-2xl font-bold text-white font-['Space_Grotesk']">Connect 4</span>
          </Link>
        </div>

        <div className="relative z-10 space-y-8">
          <h1 className="text-5xl font-bold text-white leading-tight font-['Space_Grotesk']">
            Join the<br />competition
          </h1>
          <p className="text-xl text-slate-200 max-w-md leading-relaxed">
            Create your account to track your stats, compete in tournaments, 
            and challenge friends to epic Connect 4 battles.
          </p>
          
          <div className="space-y-4 pt-8">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-[#f59e0b]/20 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M10 6v4l3 2" strokeLinecap="round"/>
                  <circle cx="10" cy="10" r="7"/>
                </svg>
              </div>
              <div>
                <div className="font-semibold text-white mb-1">Real-time Multiplayer</div>
                <div className="text-sm text-slate-300">Instant WebSocket connections for lag-free gameplay</div>
              </div>
            </div>
            
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-[#f59e0b]/20 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M4 7h12M4 13h12" strokeLinecap="round"/>
                  <rect x="3" y="3" width="14" height="14" rx="2"/>
                </svg>
              </div>
              <div>
                <div className="font-semibold text-white mb-1">Track Your Progress</div>
                <div className="text-sm text-slate-300">Detailed stats and leaderboards to track your improvement</div>
              </div>
            </div>
            
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-[#f59e0b]/20 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M14 7L10 3L6 7M10 3v11" strokeLinecap="round"/>
                  <path d="M3 14v2a1 1 0 001 1h12a1 1 0 001-1v-2"/>
                </svg>
              </div>
              <div>
                <div className="font-semibold text-white mb-1">Play with Friends</div>
                <div className="text-sm text-slate-300">Invite friends for private matches and friendly competition</div>
              </div>
            </div>
          </div>
        </div>

        <div className="relative z-10 text-slate-300 text-sm">
          © 2024 Connect 4 Multiplayer. All rights reserved.
        </div>
      </motion.div>

      {/* Right Side - Register Form */}
      <div className="flex-1 flex items-center justify-center p-8">
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="w-full max-w-md"
        >
          {/* Mobile Logo */}
          <div className="lg:hidden mb-8 text-center">
            <Link to="/" className="inline-flex items-center gap-3">
              <div className="w-10 h-10 bg-[#1a365d] rounded-xl flex items-center justify-center">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" className="text-[#f59e0b]">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                  <circle cx="12" cy="8" r="2" fill="currentColor"/>
                  <circle cx="12" cy="16" r="2" fill="currentColor"/>
                </svg>
              </div>
              <span className="text-xl font-bold text-[#1a365d] font-['Space_Grotesk']">Connect 4</span>
            </Link>
          </div>

          <div className="mb-8">
            <h2 className="text-3xl font-bold text-slate-900 mb-2 font-['Space_Grotesk']">
              Create your account
            </h2>
            <p className="text-slate-600">
              Start playing and tracking your wins today
            </p>
          </div>

          {error && (
            <motion.div 
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm"
            >
              {error}
            </motion.div>
          )}

          {/* OAuth Buttons */}
          <div className="space-y-3 mb-6">
            <button
              onClick={handleGoogleRegister}
              className="w-full flex items-center justify-center gap-3 px-4 py-3 bg-white border-2 border-slate-200 rounded-xl hover:border-slate-300 hover:bg-slate-50 transition-all duration-200 font-medium text-slate-700"
            >
              <svg width="20" height="20" viewBox="0 0 20 20">
                <path fill="#4285F4" d="M19.6 10.23c0-.82-.1-1.42-.25-2.05H10v3.72h5.5c-.15.96-.74 2.31-2.04 3.22v2.45h3.16c1.89-1.73 2.98-4.3 2.98-7.34z"/>
                <path fill="#34A853" d="M13.46 15.13c-.83.59-1.96 1-3.46 1-2.64 0-4.88-1.74-5.68-4.15H1.07v2.52C2.72 17.75 6.09 20 10 20c2.7 0 4.96-.89 6.62-2.42l-3.16-2.45z"/>
                <path fill="#FBBC05" d="M3.99 10c0-.69.12-1.35.32-1.97V5.51H1.07A9.973 9.973 0 000 10c0 1.61.39 3.14 1.07 4.49l3.24-2.52c-.2-.62-.32-1.28-.32-1.97z"/>
                <path fill="#EA4335" d="M10 3.88c1.88 0 3.13.81 3.85 1.48l2.84-2.76C14.96.99 12.7 0 10 0 6.09 0 2.72 2.25 1.07 5.51l3.24 2.52C5.12 5.62 7.36 3.88 10 3.88z"/>
              </svg>
              Sign up with Google
            </button>

            <button
              onClick={handleGitHubRegister}
              className="w-full flex items-center justify-center gap-3 px-4 py-3 bg-[#24292e] border-2 border-[#24292e] rounded-xl hover:bg-[#2f363d] transition-all duration-200 font-medium text-white"
            >
              <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" clipRule="evenodd" d="M10 0C4.477 0 0 4.477 0 10c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0110 4.836c.85.004 1.705.115 2.504.337 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C17.137 18.163 20 14.418 20 10c0-5.523-4.477-10-10-10z"/>
              </svg>
              Sign up with GitHub
            </button>
          </div>

          <div className="relative mb-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-200"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-4 bg-slate-50 text-slate-500">or sign up with email</span>
            </div>
          </div>

          {/* Email Form */}
          <form onSubmit={handleEmailRegister} className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-slate-700 mb-2">
                Username
              </label>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                className="w-full px-4 py-3 bg-white border-2 border-slate-200 rounded-xl focus:border-[#f59e0b] focus:ring-4 focus:ring-[#f59e0b]/10 outline-none transition-all duration-200"
                placeholder="player123"
              />
              <p className="text-xs text-slate-500 mt-1">3-20 characters, letters, numbers, and underscores only</p>
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-slate-700 mb-2">
                Email address
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="w-full px-4 py-3 bg-white border-2 border-slate-200 rounded-xl focus:border-[#f59e0b] focus:ring-4 focus:ring-[#f59e0b]/10 outline-none transition-all duration-200"
                placeholder="you@example.com"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-slate-700 mb-2">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full px-4 py-3 bg-white border-2 border-slate-200 rounded-xl focus:border-[#f59e0b] focus:ring-4 focus:ring-[#f59e0b]/10 outline-none transition-all duration-200"
                placeholder="••••••••"
              />
              <p className="text-xs text-slate-500 mt-1">At least 8 characters</p>
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-slate-700 mb-2">
                Confirm password
              </label>
              <input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                className="w-full px-4 py-3 bg-white border-2 border-slate-200 rounded-xl focus:border-[#f59e0b] focus:ring-4 focus:ring-[#f59e0b]/10 outline-none transition-all duration-200"
                placeholder="••••••••"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-gradient-to-r from-[#1a365d] to-[#2d4a7c] text-white font-semibold rounded-xl hover:shadow-lg hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Creating account...' : 'Create account'}
            </button>

            <p className="text-xs text-slate-500 text-center">
              By signing up, you agree to our{' '}
              <Link to="/terms" className="text-[#f59e0b] hover:underline">Terms of Service</Link>
              {' '}and{' '}
              <Link to="/privacy" className="text-[#f59e0b] hover:underline">Privacy Policy</Link>
            </p>
          </form>

          <div className="mt-6 text-center text-sm text-slate-600">
            Already have an account?{' '}
            <Link to="/login" className="text-[#f59e0b] hover:text-[#d97706] font-semibold">
              Sign in
            </Link>
          </div>

          <div className="mt-8 text-center">
            <Link to="/" className="text-sm text-slate-500 hover:text-slate-700 inline-flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <path d="M10 4L6 8L10 12"/>
              </svg>
              Continue as guest
            </Link>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default RegisterPage;
