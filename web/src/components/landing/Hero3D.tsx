import { Suspense, lazy } from 'react';
import { motion } from 'framer-motion';

// Lazy load Spline to improve initial page load
const Spline = lazy(() => import('@splinetool/react-spline'));

const LoadingFallback = () => (
  <div className="w-full h-full flex items-center justify-center">
    <motion.div
      animate={{ rotate: 360 }}
      transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
      className="w-16 h-16 rounded-full border-4 border-blue-200 border-t-blue-600"
    />
  </div>
);

export const Hero3D = () => {
  return (
    <motion.div 
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.8, delay: 0.3, ease: [0.22, 1, 0.36, 1] }}
      className="w-full h-[400px] md:h-[500px] lg:h-[600px] relative flex items-center justify-center"
    >
      {/* Glow Effect Behind */}
      <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 via-purple-500/20 to-blue-500/20 blur-3xl rounded-full scale-75" />
      
      <Suspense fallback={<LoadingFallback />}>
        <Spline 
          scene="https://prod.spline.design/UIuJ-ml8fUPy2Clu/scene.splinecode"
          className="w-full h-full"
        />
      </Suspense>
    </motion.div>
  );
};
