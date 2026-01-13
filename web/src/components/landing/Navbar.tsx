import { useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Menu, X, User, LogIn, ChevronDown, Trophy, LogOut } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';

export const Navbar = () => {
  const navigate = useNavigate();
  const { user, profile, signOut } = useAuth();
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 20);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <motion.nav
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        isScrolled
          ? 'bg-white/80 backdrop-blur-xl border-b border-gray-100 shadow-sm'
          : 'bg-transparent'
      }`}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 md:h-20">
          {/* Logo */}
          <motion.button
            onClick={() => navigate('/')}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            className="flex items-center gap-2"
          >
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-600 to-blue-700 flex items-center justify-center shadow-lg shadow-blue-500/25">
              <span className="text-white font-bold text-lg">C4</span>
            </div>
            <span className={`text-xl font-bold tracking-tight ${isScrolled ? 'text-slate-900' : 'text-white'}`}>
              Connect4<span className="text-blue-600">.ai</span>
            </span>
          </motion.button>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-6">
            <NavLink isScrolled={isScrolled} onClick={() => navigate('/leaderboard')}>
              Leaderboard
            </NavLink>
            <NavLink isScrolled={isScrolled} href="https://github.com/luxmikant/Connect4" external>
              GitHub
            </NavLink>

            {user ? (
              <div className="relative">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => setShowUserMenu(!showUserMenu)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-xl transition-all ${
                    isScrolled
                      ? 'bg-slate-100 text-slate-900 hover:bg-slate-200'
                      : 'bg-white/10 text-white hover:bg-white/20'
                  }`}
                >
                  <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                    <span className="text-white text-sm font-bold">
                      {profile?.username?.charAt(0)?.toUpperCase() || 'U'}
                    </span>
                  </div>
                  <span className="font-medium">{profile?.username || 'User'}</span>
                  <ChevronDown className={`w-4 h-4 transition-transform ${showUserMenu ? 'rotate-180' : ''}`} />
                </motion.button>

                <AnimatePresence>
                  {showUserMenu && (
                    <motion.div
                      initial={{ opacity: 0, y: 10, scale: 0.95 }}
                      animate={{ opacity: 1, y: 0, scale: 1 }}
                      exit={{ opacity: 0, y: 10, scale: 0.95 }}
                      transition={{ duration: 0.2 }}
                      className="absolute right-0 mt-2 w-52 bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden"
                    >
                      <div className="p-2">
                        <MenuButton icon={<User className="w-4 h-4" />} onClick={() => { navigate('/profile'); setShowUserMenu(false); }}>
                          My Profile
                        </MenuButton>
                        <MenuButton icon={<Trophy className="w-4 h-4" />} onClick={() => { navigate('/leaderboard'); setShowUserMenu(false); }}>
                          Leaderboard
                        </MenuButton>
                        <hr className="my-2 border-gray-100" />
                        <MenuButton icon={<LogOut className="w-4 h-4" />} onClick={() => { signOut(); setShowUserMenu(false); }} danger>
                          Sign Out
                        </MenuButton>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => navigate('/login')}
                  className={`flex items-center gap-2 px-4 py-2 rounded-xl font-medium transition-all ${
                    isScrolled
                      ? 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
                      : 'text-white/80 hover:text-white hover:bg-white/10'
                  }`}
                >
                  <LogIn className="w-4 h-4" />
                  Login
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => navigate('/lobby')}
                  className="px-6 py-2.5 rounded-full bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 transition-all"
                >
                  Play Now
                </motion.button>
              </div>
            )}
          </div>

          {/* Mobile Menu Button */}
          <motion.button
            whileTap={{ scale: 0.95 }}
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className={`md:hidden p-2 rounded-xl ${isScrolled ? 'text-slate-900' : 'text-white'}`}
          >
            {isMobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </motion.button>
        </div>
      </div>

      {/* Mobile Menu */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="md:hidden bg-white border-t border-gray-100"
          >
            <div className="px-4 py-4 space-y-2">
              <MobileNavButton onClick={() => { navigate('/leaderboard'); setIsMobileMenuOpen(false); }}>
                Leaderboard
              </MobileNavButton>
              <MobileNavButton href="https://github.com/luxmikant/Connect4" external>
                GitHub
              </MobileNavButton>
              {user ? (
                <>
                  <MobileNavButton onClick={() => { navigate('/profile'); setIsMobileMenuOpen(false); }}>
                    My Profile
                  </MobileNavButton>
                  <MobileNavButton onClick={() => { signOut(); setIsMobileMenuOpen(false); }} danger>
                    Sign Out
                  </MobileNavButton>
                </>
              ) : (
                <>
                  <MobileNavButton onClick={() => { navigate('/login'); setIsMobileMenuOpen(false); }}>
                    Login
                  </MobileNavButton>
                </>
              )}
              <motion.button
                whileTap={{ scale: 0.98 }}
                onClick={() => { navigate('/lobby'); setIsMobileMenuOpen(false); }}
                className="w-full py-3 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold text-center"
              >
                Play Now
              </motion.button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.nav>
  );
};

// Helper Components
const NavLink = ({ 
  children, 
  isScrolled, 
  onClick, 
  href, 
  external 
}: { 
  children: ReactNode; 
  isScrolled: boolean; 
  onClick?: () => void; 
  href?: string; 
  external?: boolean;
}) => {
  const className = `font-medium transition-colors ${
    isScrolled ? 'text-slate-600 hover:text-slate-900' : 'text-white/80 hover:text-white'
  }`;
  
  if (href) {
    return (
      <a href={href} target={external ? '_blank' : undefined} rel={external ? 'noopener noreferrer' : undefined} className={className}>
        {children}
      </a>
    );
  }
  
  return <button onClick={onClick} className={className}>{children}</button>;
};

const MenuButton = ({ 
  children, 
  icon, 
  onClick, 
  danger 
}: { 
  children: ReactNode; 
  icon: ReactNode; 
  onClick: () => void; 
  danger?: boolean;
}) => (
  <button
    onClick={onClick}
    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left transition-colors ${
      danger
        ? 'text-red-600 hover:bg-red-50'
        : 'text-slate-700 hover:bg-slate-50'
    }`}
  >
    {icon}
    <span className="font-medium">{children}</span>
  </button>
);

const MobileNavButton = ({ 
  children, 
  onClick, 
  href, 
  external, 
  danger 
}: { 
  children: ReactNode; 
  onClick?: () => void; 
  href?: string; 
  external?: boolean;
  danger?: boolean;
}) => {
  const className = `block w-full py-3 px-4 rounded-xl text-left font-medium transition-colors ${
    danger ? 'text-red-600 hover:bg-red-50' : 'text-slate-700 hover:bg-slate-50'
  }`;
  
  if (href) {
    return (
      <a href={href} target={external ? '_blank' : undefined} rel={external ? 'noopener noreferrer' : undefined} className={className}>
        {children}
      </a>
    );
  }
  
  return <button onClick={onClick} className={className}>{children}</button>;
};
