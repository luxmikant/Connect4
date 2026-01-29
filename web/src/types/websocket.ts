// Client to Server
export const MessageType = {
  JoinQueue: "join_queue",
  LeaveQueue: "leave_queue",
  JoinGame: "join_game",
  MakeMove: "make_move",
  Reconnect: "reconnect",
  LeaveGame: "leave_game",
  RematchCustomRoom: "rematch_custom_room",
  Ping: "ping",
  PlayWithBot: "play_with_bot",
  CreateCustomRoom: "create_custom_room",
  JoinCustomRoom: "join_custom_room",
  // Server to Client
  QueueJoined: "queue_joined",
  QueueStatus: "queue_status",
  MatchFound: "match_found",
  RoomCreated: "room_created",
  WaitingForOpponent: "waiting_for_opponent",
  GameStarted: "game_started",
  MoveMade: "move_made",
  GameEnded: "game_ended",
  GameState: "game_state",
  PlayerJoined: "player_joined",
  PlayerLeft: "player_left",
  Error: "error",
  Pong: "pong",
} as const;

export type MessageType = (typeof MessageType)[keyof typeof MessageType];

export interface WebSocketMessage {
  type: MessageType;
  payload: any;
  timestamp: string;
}

export interface QueueStatus {
  inQueue: boolean;
  position: number;
  waitTime: string;
  timeRemaining: string;
}

export interface GameState {
  id: string;
  board: number[][];
  currentTurn: number;
  status: string;
  players: {
    [key: number]: string; // playerId -> name
  };
  winner?: number;
  isBot?: boolean;
}
