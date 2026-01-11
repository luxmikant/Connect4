import { useCallback, useRef } from 'react';

type SoundType = 'click' | 'drop' | 'win' | 'lose' | 'draw' | 'error';

export const useGameSound = () => {
  const audioContext = useRef<AudioContext | null>(null);

  const initAudio = useCallback(() => {
    if (!audioContext.current) {
      audioContext.current = new (window.AudioContext || (window as any).webkitAudioContext)();
    }
    if (audioContext.current.state === 'suspended') {
      audioContext.current.resume();
    }
  }, []);

  const playSound = useCallback((type: SoundType) => {
    initAudio();
    const ctx = audioContext.current;
    if (!ctx) return;

    const oscillator = ctx.createOscillator();
    const gainNode = ctx.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(ctx.destination);

    switch (type) {
      case 'click':
        // High-pitched blip
        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(800, ctx.currentTime);
        oscillator.frequency.exponentialRampToValueAtTime(1200, ctx.currentTime + 0.1);
        gainNode.gain.setValueAtTime(0.1, ctx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.1);
        oscillator.start();
        oscillator.stop(ctx.currentTime + 0.1);
        break;

      case 'drop':
        // Lower thud/whoosh
        oscillator.type = 'triangle';
        oscillator.frequency.setValueAtTime(400, ctx.currentTime);
        oscillator.frequency.exponentialRampToValueAtTime(100, ctx.currentTime + 0.2);
        gainNode.gain.setValueAtTime(0.2, ctx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2);
        oscillator.start();
        oscillator.stop(ctx.currentTime + 0.2);
        break;

      case 'win':
        // Ascending major arpeggio
        const now = ctx.currentTime;
        [523.25, 659.25, 783.99, 1046.50].forEach((freq, i) => {
          const osc = ctx.createOscillator();
          const gn = ctx.createGain();
          osc.connect(gn);
          gn.connect(ctx.destination);
          
          osc.type = 'sine';
          osc.frequency.value = freq;
          
          gn.gain.setValueAtTime(0.1, now + i * 0.1);
          gn.gain.exponentialRampToValueAtTime(0.01, now + i * 0.1 + 0.3);
          
          osc.start(now + i * 0.1);
          osc.stop(now + i * 0.1 + 0.3);
        });
        break;

      case 'lose':
         // Descending tritone (unpleasant but not harsh)
         const nowL = ctx.currentTime;
         [783.99, 554.37, 392.00].forEach((freq, i) => {
           const osc = ctx.createOscillator();
           const gn = ctx.createGain();
           osc.connect(gn);
           gn.connect(ctx.destination);
           
           osc.type = 'sawtooth';
           osc.frequency.value = freq;
           
           gn.gain.setValueAtTime(0.05, nowL + i * 0.15);
           gn.gain.exponentialRampToValueAtTime(0.01, nowL + i * 0.15 + 0.3);
           
           osc.start(nowL + i * 0.15);
           osc.stop(nowL + i * 0.15 + 0.3);
         });
         break;
      
      case 'error':
        oscillator.type = 'sawtooth';
        oscillator.frequency.setValueAtTime(150, ctx.currentTime);
        oscillator.frequency.linearRampToValueAtTime(100, ctx.currentTime + 0.2);
        gainNode.gain.setValueAtTime(0.1, ctx.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2);
        oscillator.start();
        oscillator.stop(ctx.currentTime + 0.2);
        break;
    }
  }, [initAudio]);

  return { playSound };
};
