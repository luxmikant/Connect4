import { motion } from 'framer-motion';
import type { ReactNode } from 'react';

interface FeatureCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  gradient?: string;
}

export const FeatureCard = ({ 
  icon, 
  title, 
  description, 
  gradient = 'from-blue-500 to-blue-600' 
}: FeatureCardProps) => {
  return (
    <motion.div
      whileHover={{ y: -8, scale: 1.02 }}
      transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
      className="group relative"
    >
      {/* Card Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-3xl blur-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
      
      <div className="relative bg-white/80 backdrop-blur-xl border border-gray-100 rounded-3xl p-8 shadow-lg shadow-gray-200/50 group-hover:shadow-xl group-hover:shadow-blue-500/10 transition-all duration-300">
        {/* Icon Container */}
        <div className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${gradient} flex items-center justify-center mb-6 shadow-lg group-hover:scale-110 transition-transform duration-300`}>
          <div className="text-white">
            {icon}
          </div>
        </div>
        
        {/* Content */}
        <h3 className="text-xl font-bold text-slate-900 mb-3 tracking-tight">
          {title}
        </h3>
        <p className="text-slate-600 leading-relaxed">
          {description}
        </p>
        
        {/* Decorative Element */}
        <div className="absolute top-4 right-4 w-24 h-24 bg-gradient-to-br from-gray-100 to-transparent rounded-full opacity-50" />
      </div>
    </motion.div>
  );
};
