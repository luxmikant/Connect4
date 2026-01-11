import { useState, useEffect, useCallback } from 'react';
import { wsService } from '../services/websocket';
import { MessageType } from '../types/websocket';
import type { WebSocketMessage, GameState, QueueStatus } from '../types/websocket';
import { toast } from 'react-hot-toast';
import { usePlayer } from './usePlayer';

export const useGame = () => {
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [queueStatus, setQueueStatus] = useState<QueueStatus | null>(null);
  const { username } = usePlayer();
  const [isMyTurn, setIsMyTurn] = useState(false);
  const [myColor, setMyColor] = useState<'red' | 'yellow' | null>(null);

  const handleQueueJoined = useCallback((message: WebSocketMessage) => {
    console.log("Queue joined:", message.payload);
    setQueueStatus({
      inQueue: true,
      position: message.payload.position || 1,
      waitTime: "0s",
      timeRemaining: message.payload.estimatedWait || "10s",
    });
  }, []);

  const handleQueueStatus = useCallback((message: WebSocketMessage) => {
    console.log("Queue status:", message.payload);
    setQueueStatus({
      inQueue: message.payload.inQueue,
      position: message.payload.position,
      waitTime: message.payload.waitTime,
      timeRemaining: message.payload.timeRemaining,
    });
  }, []);

  const handleMatchFound = useCallback((message: WebSocketMessage) => {
    console.log("Match found:", message.payload);
    setQueueStatus(null);
    toast.success(`Match found! Playing against ${message.payload.opponent}`);
  }, []);

  const handleGameStarted = useCallback((message: WebSocketMessage) => {
    console.log("🎮 Game started message received:", message.payload);
    console.log("🎮 Full payload:", JSON.stringify(message.payload, null, 2));
    const payload = message.payload;
    
    // Extract game state from the game_started message
    const state: GameState = {
      id: payload.gameId,
      board: payload.board?.cells || Array(6).fill(null).map(() => Array(7).fill(0)),
      currentTurn: payload.currentTurn === 'red' ? 1 : 2,
      status: 'in_progress',
      players: {
        1: payload.currentTurn === payload.yourColor ? username! : payload.opponent,
        2: payload.currentTurn === payload.yourColor ? payload.opponent : username!,
      },
      isBot: payload.isBot,
    };
    
    console.log("🎮 Extracted board:", state.board);
    console.log("🎮 Board length:", state.board?.length);
    
    // Set player colors
    if (payload.yourColor === 'red') {
      state.players[1] = username!;
      state.players[2] = payload.opponent;
      setMyColor('red');
    } else {
      state.players[1] = payload.opponent;
      state.players[2] = username!;
      setMyColor('yellow');
    }
    
    console.log("🎮 Setting game state:", state);
    setGameState(state);
    console.log("🎮 Clearing queue status");
    setQueueStatus(null);
    toast.success("Game started!");
  }, [username]);

  const handleGameState = useCallback((message: WebSocketMessage) => {
    console.log("Game state update:", message.payload);
    const state = message.payload as GameState;
    setGameState(state);
  }, []);

  const handleMoveMade = useCallback((message: WebSocketMessage) => {
    console.log("Move made:", message.payload);
    const payload = message.payload;
    
    setGameState(prev => {
      if (!prev) return prev;
      
      // Update board from the payload
      const newBoard = payload.board?.cells || prev.board;
      const nextTurn = payload.nextTurn === 'red' ? 1 : (payload.nextTurn === 'yellow' ? 2 : prev.currentTurn);
      
      return {
        ...prev,
        board: newBoard,
        currentTurn: nextTurn,
      };
    });
  }, []);

  const handleGameEnded = useCallback((message: WebSocketMessage) => {
    console.log("Game ended:", message.payload);
    const payload = message.payload;
    
    setGameState(prev => {
      if (!prev || !prev.players) return prev;
      
      // Find winner player number if there's a winner
      let winnerNum: number | undefined;
      if (payload.winner) {
        const entries = Object.entries(prev.players);
        const winnerEntry = entries.find(([, name]) => name === payload.winner);
        if (winnerEntry) {
          winnerNum = parseInt(winnerEntry[0]);
        }
      }
      
      return {
        ...prev,
        status: 'completed',
        winner: winnerNum,
      };
    });
    
    if (payload.winner) {
      toast.success(`${payload.winner} wins!`);
    } else {
      toast("It's a draw!");
    }
  }, []);

  const handleError = useCallback((message: WebSocketMessage) => {
    console.error("Game error:", message.payload);
    toast.error(message.payload.error || message.payload.message || "An error occurred.");
  }, []);

  // Connect WebSocket when username is set
  useEffect(() => {
    if (username) {
      wsService.connect(username);
    }
  }, [username]);

  // Subscribe to WebSocket messages
  useEffect(() => {
    const subscriptions = [
      wsService.subscribe(MessageType.QueueJoined, handleQueueJoined),
      wsService.subscribe(MessageType.QueueStatus, handleQueueStatus),
      wsService.subscribe(MessageType.MatchFound, handleMatchFound),
      wsService.subscribe(MessageType.GameStarted, handleGameStarted),
      wsService.subscribe(MessageType.GameState, handleGameState),
      wsService.subscribe(MessageType.MoveMade, handleMoveMade),
      wsService.subscribe(MessageType.GameEnded, handleGameEnded),
      wsService.subscribe(MessageType.Error, handleError),
    ];

    return () => {
      subscriptions.forEach(unsubscribe => unsubscribe());
    };
  }, [handleQueueJoined, handleQueueStatus, handleMatchFound, handleGameStarted, handleGameState, handleMoveMade, handleGameEnded, handleError]);

  // Update turn tracking when game state changes
  useEffect(() => {
    if (gameState && gameState.players && username) {
      const playerEntries = Object.entries(gameState.players);
      const myEntry = playerEntries.find(([, name]) => name === username);
      if (myEntry) {
        const myColorNum = parseInt(myEntry[0]);
        setIsMyTurn(gameState.currentTurn === myColorNum);
        setMyColor(myColorNum === 1 ? 'red' : 'yellow');
      }
    } else {
      setIsMyTurn(false);
    }
  }, [gameState, username]);

  const makeMove = useCallback((column: number) => {
    if (isMyTurn && gameState) {
      console.log("Making move:", column);
      wsService.send(MessageType.MakeMove, { gameId: gameState.id, column });
    }
  }, [isMyTurn, gameState]);

  const leaveQueue = useCallback(() => {
    wsService.send(MessageType.LeaveQueue);
    setQueueStatus(null);
  }, []);

  return { 
    gameState, 
    queueStatus,
    isMyTurn, 
    myColor,
    makeMove,
    leaveQueue,
  };
};
