import { MessageType } from "../types/websocket";
import type { WebSocketMessage } from "../types/websocket";

type MessageHandler = (message: WebSocketMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectInterval: number = 3000;
  private handlers: Map<MessageType, Set<MessageHandler>> = new Map();
  private isConnecting: boolean = false;

  constructor(url: string = "ws://localhost:8080/ws") {
    this.url = url;
  }

  public connect() {
    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) return;

    this.isConnecting = true;
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      console.log("WebSocket connected");
      this.isConnecting = false;
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.notify(message);
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected");
      this.isConnecting = false;
      this.ws = null;
      setTimeout(() => this.connect(), this.reconnectInterval);
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.ws?.close();
    };
  }

  public send(type: MessageType, payload: any = {}) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message = {
        type,
        payload,
        timestamp: new Date().toISOString(),
      };
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
