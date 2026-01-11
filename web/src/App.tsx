import { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Home } from './pages/Home';
import { Game } from './pages/Game';
import { Lobby } from './pages/Lobby';
import { Leaderboard } from './pages/Leaderboard';
import { wsService } from './services/websocket';
import { usePlayer } from './hooks/usePlayer';

function App() {
  const { username } = usePlayer();

  useEffect(() => {
    wsService.connect();
  }, []);

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/lobby" element={<Lobby />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
        <Route 
          path="/game" 
          element={username ? <Game /> : <Navigate to="/lobby" />} 
        />
      </Routes>
    </Router>
  );
}

export default App;
