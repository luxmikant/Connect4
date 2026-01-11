import { MessageType } from "../types/websocket";
import type { WebSocketMessage } from "../types/websocket";

type MessageHandler = (message: WebSocketMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private baseUrl: string;
  private reconnectInterval: number = 3000;
  private handlers: Map<MessageType, Set<MessageHandler>> = new Map();
  private isConnecting: boolean = false;
  private currentUserId: string | null = null;

  constructor(baseUrl?: string) {
    // baseUrl should be just the domain (e.g., wss://example.com), not include /ws path
    // Use ws:// for localhost, wss:// for production
    const defaultUrl = window.location.hostname === 'localhost' 
      ? 'ws://localhost:8080'
      : `wss://${window.location.host}`;
    this.baseUrl = baseUrl || import.meta.env.VITE_WS_URL || defaultUrl;
  }

  public connect(userId?: string) {
    if (userId) {
      this.currentUserId = userId;
    }

    // Need userId to connect
    if (!this.currentUserId) {
      console.log("WebSocket: No userId provided, waiting for user to set username");
      return;
    }

    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) return;

    this.isConnecting = true;
    const url = `${this.baseUrl}/ws?userId=${encodeURIComponent(this.currentUserId)}`;
    console.log("WebSocket: Connecting to", url);
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log("WebSocket connected for user:", this.currentUserId);
      this.isConnecting = false;
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        console.log("WebSocket message received:", message.type, message.payload);
        this.notify(message);
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected");
      this.isConnecting = false;
      this.ws = null;
      // Only reconnect if we have a userId
      if (this.currentUserId) {
        setTimeout(() => this.connect(), this.reconnectInterval);
      }
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.ws?.close();
    };
  }

  public disconnect() {
    this.currentUserId = null;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  public send(type: MessageType, payload: any = {}) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message = {
        type,
        payload,
        timestamp: new Date().toISOString(),
      };
      console.log("WebSocket sending:", type, payload);
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket is not connected. Cannot send message:", type);
    }
  }

  public subscribe(type: MessageType, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)?.add(handler);

    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  private notify(message: WebSocketMessage) {
    const handlers = this.handlers.get(message.type);
    handlers?.forEach((handler) => handler(message));
  }
}

export const wsService = new WebSocketService();
