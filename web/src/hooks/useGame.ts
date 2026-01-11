import { useState, useEffect, useCallback } from 'react';
import { wsService } from '../services/websocket';
import { MessageType } from '../types/websocket';
import type { WebSocketMessage, GameState } from '../types/websocket';
import { toast } from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';
import { usePlayer } from './usePlayer';

export const useGame = () => {
  const [gameState, setGameState] = useState<GameState | null>(null);
  const { username } = usePlayer();
  const [isMyTurn, setIsMyTurn] = useState(false);
  const navigate = useNavigate();

  const handleGameStarted = useCallback((message: WebSocketMessage) => {
    const state = message.payload as GameState;
    setGameState(state);
    toast.success("Game started!");
  }, []);

  const handleGameState = useCallback((message: WebSocketMessage) => {
    const state = message.payload as GameState;
    setGameState(state);
  }, []);

  const handleMoveMade = useCallback((message: WebSocketMessage) => {
    const state = message.payload as GameState;
    setGameState(state);
  }, []);

  const handleGameEnded = useCallback((message: WebSocketMessage) => {
    const state = message.payload as GameState;
    setGameState(state);
    if (state.winner) {
      const winnerName = state.players[state.winner];
      toast.success(`${winnerName} wins!`);
    } else {
      toast("It's a draw!");
    }
  }, []);

  const handleError = useCallback((message: WebSocketMessage) => {
    toast.error(message.payload.error || "An error occurred.");
    navigate('/');
  }, [navigate]);

  useEffect(() => {
    const subscriptions = [
      wsService.subscribe(MessageType.GameStarted, handleGameStarted),
      wsService.subscribe(MessageType.GameState, handleGameState),
      wsService.subscribe(MessageType.MoveMade, handleMoveMade),
      wsService.subscribe(MessageType.GameEnded, handleGameEnded),
      wsService.subscribe(MessageType.Error, handleError),
    ];

    if (username) {
      wsService.send(MessageType.JoinQueue, { username });
    }

    return () => {
      subscriptions.forEach(unsubscribe => unsubscribe());
      if (username) {
        wsService.send(MessageType.LeaveQueue);
      }
    };
  }, [username, handleGameStarted, handleGameState, handleMoveMade, handleGameEnded, handleError, navigate]);

  useEffect(() => {
    if (gameState && username) {
      const playerEntries = Object.entries(gameState.players);
      const myEntry = playerEntries.find(([, name]) => name === username);
      if (myEntry) {
        const myColor = parseInt(myEntry[0]);
        setIsMyTurn(gameState.currentTurn === myColor);
      }
    } else {
      setIsMyTurn(false);
    }
  }, [gameState, username]);

  const makeMove = (column: number) => {
    if (isMyTurn && gameState) {
      wsService.send(MessageType.MakeMove, { gameId: gameState.id, column });
    }
  };

  return { gameState, isMyTurn, makeMove };
};
