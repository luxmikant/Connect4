import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { 
  Play, 
  Cpu, 
  Zap, 
  Users, 
  Code2, 
  ArrowRight,
  Sparkles,
  Shield,
  Globe
} from 'lucide-react';
import { Navbar, Hero3D, FeatureCard, Footer, ScrollReveal } from '../components/landing';

export const Home = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-900 to-slate-950">
      {/* Navbar */}
      <Navbar />

      {/* Hero Section */}
      <section className="relative min-h-screen flex items-center overflow-hidden pt-20">
        {/* Background Elements */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-[10%] left-[5%] w-[600px] h-[600px] bg-blue-600/20 rounded-full blur-[150px]" />
          <div className="absolute bottom-[10%] right-[5%] w-[500px] h-[500px] bg-purple-600/15 rounded-full blur-[150px]" />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-blue-500/5 rounded-full blur-[100px]" />
        </div>

        <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 md:py-20">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-8 items-center">
            {/* Left Side - Text */}
            <motion.div
              initial={{ opacity: 0, x: -50 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
              className="text-center lg:text-left"
            >
              {/* Badge */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-sm font-medium mb-8"
              >
                <Sparkles className="w-4 h-4" />
                AI-Powered Strategy Game
              </motion.div>

              {/* Headline */}
              <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold tracking-tight text-white mb-6">
                The Classic Strategy.
                <br />
                <span className="bg-gradient-to-r from-blue-400 via-blue-500 to-purple-500 bg-clip-text text-transparent">
                  Reimagined.
                </span>
              </h1>

              {/* Subtext */}
              <p className="text-lg sm:text-xl text-slate-400 max-w-xl mx-auto lg:mx-0 mb-10 leading-relaxed">
                Experience Connect 4 with smart AI opponents, real-time multiplayer, 
                and beautiful modern design. No login required.
              </p>

              {/* CTA Buttons */}
              <div className="flex flex-col sm:flex-row items-center gap-4 justify-center lg:justify-start">
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => navigate('/lobby')}
                  className="group w-full sm:w-auto px-8 py-4 rounded-full bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold text-lg shadow-xl shadow-blue-500/25 hover:shadow-blue-500/40 transition-all flex items-center justify-center gap-3"
                >
                  <Play className="w-5 h-5 fill-current" />
                  Start Playing
                  <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                </motion.button>

                <motion.a
                  href="https://github.com/luxmikant/Connect4"
                  target="_blank"
                  rel="noopener noreferrer"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  className="w-full sm:w-auto px-8 py-4 rounded-full bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white font-semibold text-lg transition-all flex items-center justify-center gap-3"
                >
                  <Code2 className="w-5 h-5" />
                  View Code
                </motion.a>
              </div>

              {/* Stats */}
              <div className="flex items-center gap-8 mt-12 justify-center lg:justify-start">
                <Stat value="10K+" label="Games Played" />
                <Stat value="5ms" label="Move Response" />
                <Stat value="100%" label="Open Source" />
              </div>
            </motion.div>

            {/* Right Side - 3D Model */}
            <div className="relative">
              <Hero3D />
            </div>
          </div>
        </div>

        {/* Scroll Indicator */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1.5 }}
          className="absolute bottom-8 left-1/2 -translate-x-1/2"
        >
          <motion.div
            animate={{ y: [0, 10, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
            className="w-6 h-10 rounded-full border-2 border-white/20 flex items-start justify-center p-2"
          >
            <div className="w-1.5 h-3 bg-white/40 rounded-full" />
          </motion.div>
        </motion.div>
      </section>

      {/* Features Section */}
      <section className="relative py-24 md:py-32 bg-gradient-to-b from-slate-950 to-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Section Header */}
          <ScrollReveal className="text-center mb-16">
            <h2 className="text-4xl sm:text-5xl font-bold text-white mb-6 tracking-tight">
              Why Choose Connect4.ai?
            </h2>
            <p className="text-lg text-slate-400 max-w-2xl mx-auto">
              Built with cutting-edge technology for the smoothest gaming experience
            </p>
          </ScrollReveal>

          {/* Feature Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
            <ScrollReveal delay={0.1}>
              <FeatureCard
                icon={<Cpu className="w-6 h-6" />}
                title="Smart AI"
                description="Powered by Minimax Algorithm with alpha-beta pruning. Three difficulty levels to challenge any skill."
                gradient="from-blue-500 to-blue-600"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.2}>
              <FeatureCard
                icon={<Zap className="w-6 h-6" />}
                title="Instant Play"
                description="No signup required. Jump straight into a game within seconds. Your progress saves automatically."
                gradient="from-amber-500 to-orange-600"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.3}>
              <FeatureCard
                icon={<Users className="w-6 h-6" />}
                title="Real-time Multiplayer"
                description="Play against friends or random opponents with WebSocket-powered instant synchronization."
                gradient="from-emerald-500 to-teal-600"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.4}>
              <FeatureCard
                icon={<Globe className="w-6 h-6" />}
                title="Custom Rooms"
                description="Create private game rooms with shareable codes. Perfect for playing with friends anywhere."
                gradient="from-purple-500 to-violet-600"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.5}>
              <FeatureCard
                icon={<Shield className="w-6 h-6" />}
                title="Secure & Fair"
                description="Server-side game validation ensures fair play. No cheating possible with our architecture."
                gradient="from-rose-500 to-pink-600"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.6}>
              <FeatureCard
                icon={<Code2 className="w-6 h-6" />}
                title="Open Source"
                description="Fully open source on GitHub. Built with Go, React, and modern web technologies."
                gradient="from-slate-600 to-slate-700"
              />
            </ScrollReveal>
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="py-24 md:py-32 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <ScrollReveal className="text-center mb-16">
            <h2 className="text-4xl sm:text-5xl font-bold text-slate-900 mb-6 tracking-tight">
              How It Works
            </h2>
            <p className="text-lg text-slate-600 max-w-2xl mx-auto">
              Get started in three simple steps
            </p>
          </ScrollReveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 lg:gap-12">
            <ScrollReveal delay={0.1}>
              <Step
                number="01"
                title="Choose Your Mode"
                description="Play against AI for practice or jump into multiplayer for competitive matches"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.2}>
              <Step
                number="02"
                title="Join a Game"
                description="Get matched instantly or create a private room to play with friends"
              />
            </ScrollReveal>

            <ScrollReveal delay={0.3}>
              <Step
                number="03"
                title="Drop & Win"
                description="Connect four discs in a row to claim victory and climb the leaderboard"
              />
            </ScrollReveal>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-24 md:py-32 bg-gradient-to-b from-white to-slate-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <ScrollReveal>
            <h2 className="text-4xl sm:text-5xl font-bold text-slate-900 mb-6 tracking-tight">
              Ready to Play?
            </h2>
            <p className="text-lg text-slate-600 mb-10 max-w-xl mx-auto">
              Join thousands of players and experience the classic game with a modern twist.
            </p>
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => navigate('/lobby')}
              className="group px-10 py-5 rounded-full bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold text-xl shadow-xl shadow-blue-500/25 hover:shadow-blue-500/40 transition-all inline-flex items-center gap-3"
            >
              <Play className="w-6 h-6 fill-current" />
              Start Playing Now
              <ArrowRight className="w-6 h-6 group-hover:translate-x-1 transition-transform" />
            </motion.button>
          </ScrollReveal>
        </div>
      </section>

      {/* Footer */}
      <Footer />
    </div>
  );
};

// Helper Components
const Stat = ({ value, label }: { value: string; label: string }) => (
  <div className="text-center lg:text-left">
    <div className="text-2xl font-bold text-white">{value}</div>
    <div className="text-sm text-slate-500">{label}</div>
  </div>
);

const Step = ({ 
  number, 
  title, 
  description 
}: { 
  number: string; 
  title: string; 
  description: string;
}) => (
  <div className="text-center">
    <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-blue-600 text-white font-bold text-xl mb-6 shadow-lg shadow-blue-500/25">
      {number}
    </div>
    <h3 className="text-xl font-bold text-slate-900 mb-3">{title}</h3>
    <p className="text-slate-600 leading-relaxed">{description}</p>
  </div>
);