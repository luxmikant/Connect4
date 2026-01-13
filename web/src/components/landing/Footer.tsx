import { motion } from 'framer-motion';
import type { ReactNode } from 'react';
import { Github, Twitter, Heart } from 'lucide-react';

export const Footer = () => {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="relative bg-slate-900 text-white overflow-hidden">
      {/* Background Gradient */}
      <div className="absolute inset-0 bg-gradient-to-t from-slate-950 to-slate-900" />
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-blue-500/10 rounded-full blur-3xl" />
      
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
          {/* Brand Section */}
          <div className="md:col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center">
                <span className="text-white font-bold text-lg">C4</span>
              </div>
              <span className="text-xl font-bold tracking-tight">
                Connect4<span className="text-blue-400">.ai</span>
              </span>
            </div>
            <p className="text-slate-400 max-w-sm leading-relaxed">
              Experience the classic strategy game reimagined with modern AI, beautiful design, and seamless multiplayer gameplay.
            </p>
          </div>
          
          {/* Quick Links */}
          <div>
            <h4 className="font-semibold text-white mb-4">Quick Links</h4>
            <ul className="space-y-3">
              <FooterLink href="/lobby">Play Now</FooterLink>
              <FooterLink href="/leaderboard">Leaderboard</FooterLink>
              <FooterLink href="/login">Login</FooterLink>
              <FooterLink href="/register">Register</FooterLink>
            </ul>
          </div>
          
          {/* Resources */}
          <div>
            <h4 className="font-semibold text-white mb-4">Resources</h4>
            <ul className="space-y-3">
              <FooterLink href="https://github.com/luxmikant/Connect4" external>
                GitHub Repository
              </FooterLink>
              <FooterLink href="https://github.com/luxmikant/Connect4/issues" external>
                Report an Issue
              </FooterLink>
              <FooterLink href="https://github.com/luxmikant/Connect4#readme" external>
                Documentation
              </FooterLink>
            </ul>
          </div>
        </div>
        
        {/* Divider */}
        <div className="border-t border-slate-800 pt-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            {/* Copyright */}
            <p className="text-slate-500 text-sm">
              © {currentYear} Connect4.ai. All rights reserved.
            </p>
            
            {/* Made with Love */}
            <p className="flex items-center gap-2 text-slate-500 text-sm">
              Made with <Heart className="w-4 h-4 text-red-500 fill-red-500" /> by Luxmikant
            </p>
            
            {/* Social Links */}
            <div className="flex items-center gap-4">
              <SocialLink href="https://github.com/luxmikant/Connect4" icon={<Github className="w-5 h-5" />} />
              <SocialLink href="https://twitter.com" icon={<Twitter className="w-5 h-5" />} />
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
};

const FooterLink = ({ 
  children, 
  href, 
  external 
}: { 
  children: ReactNode; 
  href: string; 
  external?: boolean;
}) => (
  <li>
    <a
      href={href}
      target={external ? '_blank' : undefined}
      rel={external ? 'noopener noreferrer' : undefined}
      className="text-slate-400 hover:text-white transition-colors duration-200"
    >
      {children}
    </a>
  </li>
);

const SocialLink = ({ 
  href, 
  icon 
}: { 
  href: string; 
  icon: ReactNode;
}) => (
  <motion.a
    href={href}
    target="_blank"
    rel="noopener noreferrer"
    whileHover={{ scale: 1.1 }}
    whileTap={{ scale: 0.95 }}
    className="w-10 h-10 rounded-xl bg-slate-800 hover:bg-slate-700 flex items-center justify-center text-slate-400 hover:text-white transition-colors"
  >
    {icon}
  </motion.a>
);
