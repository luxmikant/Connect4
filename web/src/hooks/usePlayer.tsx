import { createContext, useState, useContext } from 'react';
import type { ReactNode } from 'react';

interface PlayerContextType {
  username: string | null;
  setUsername: (name: string) => void;
}

const PlayerContext = createContext<PlayerContextType | undefined>(undefined);

export const PlayerProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [username, setUsernameState] = useState<string | null>(() => {
    return localStorage.getItem('connect4_username');
  });

  const setUsername = (name: string) => {
    localStorage.setItem('connect4_username', name);
    setUsernameState(name);
  };

  return (
    <PlayerContext.Provider value={{ username, setUsername }}>
      {children}
    </PlayerContext.Provider>
  );
};

export const usePlayer = () => {
  const context = useContext(PlayerContext);
  if (context === undefined) {
    throw new Error('usePlayer must be used within a PlayerProvider');
  }
  return context;
};
