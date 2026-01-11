import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { Home } from './pages/Home';
import { Game } from './pages/Game';
import { Lobby } from './pages/Lobby';
import { Leaderboard } from './pages/Leaderboard';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import { usePlayer } from './hooks/usePlayer';

function App() {
  const { username } = usePlayer();

  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/auth/callback" element={<AuthCallbackPage />} />
          <Route path="/lobby" element={<Lobby />} />
          <Route path="/leaderboard" element={<Leaderboard />} />
          <Route 
            path="/game" 
            element={username ? <Game /> : <Navigate to="/lobby" />} 
          />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
