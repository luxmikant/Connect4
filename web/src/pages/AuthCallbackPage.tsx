import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';

const AuthCallbackPage: React.FC = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Supabase automatically handles the OAuth callback
    // The session will be set via the AuthContext listener
    // We just need to redirect after a brief moment
    const timer = setTimeout(() => {
      // Check if there's an error in the URL
      const hashParams = new URLSearchParams(window.location.hash.substring(1));
      const errorParam = hashParams.get('error');
      const errorDescription = hashParams.get('error_description');

      if (errorParam) {
        setError(errorDescription || 'Authentication failed');
        setTimeout(() => {
          navigate('/login');
        }, 3000);
      } else {
        // Success - redirect to home
        navigate('/');
      }
    }, 1500);

    return () => clearTimeout(timer);
  }, [navigate]);

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-8">
      <motion.div 
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center"
      >
        {error ? (
          <div className="space-y-4">
            <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-red-600">
                <circle cx="12" cy="12" r="10"/>
                <path d="M15 9l-6 6M9 9l6 6"/>
              </svg>
            </div>
            <h2 className="text-2xl font-bold text-slate-900 font-['Space_Grotesk']">
              Authentication Failed
            </h2>
            <p className="text-slate-600 max-w-md">
              {error}
            </p>
            <p className="text-sm text-slate-500">
              Redirecting to login page...
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
              className="w-16 h-16 bg-[#1a365d] rounded-full flex items-center justify-center mx-auto"
            >
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" className="text-[#f59e0b]">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="8" r="2" fill="currentColor"/>
                <circle cx="12" cy="16" r="2" fill="currentColor"/>
              </svg>
            </motion.div>
            <h2 className="text-2xl font-bold text-slate-900 font-['Space_Grotesk']">
              Completing sign in...
            </h2>
            <p className="text-slate-600">
              Please wait while we set up your account
            </p>
          </div>
        )}
      </motion.div>
    </div>
  );
};

export default AuthCallbackPage;
