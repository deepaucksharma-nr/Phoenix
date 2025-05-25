import { io, Socket } from 'socket.io-client';
import { env } from '@utils/env';
import { store } from '@/store';
import { updateExperiment, setExperimentMetrics } from '@/store/slices/experimentSlice';
import { addNotification } from '@/store/slices/notificationSlice';

export interface WebSocketMessage {
  type: string;
  payload: any;
  timestamp: number;
}

export type WebSocketEventHandler = (data: any) => void;

class WebSocketService {
  private socket: Socket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private eventHandlers: Map<string, Set<WebSocketEventHandler>> = new Map();
  private isConnecting = false;

  constructor() {
    this.setupEventHandlers();
  }

  private setupEventHandlers() {
    // Built-in event handlers
    this.on('experiment:update', (data) => {
      store.dispatch(updateExperiment(data));
    });

    this.on('experiment:metrics', (data) => {
      store.dispatch(setExperimentMetrics({
        experimentId: data.experimentId,
        metrics: data.metrics,
      }));
    });

    this.on('notification', (data) => {
      store.dispatch(addNotification({
        type: data.type || 'info',
        message: data.message,
        description: data.description,
        duration: data.duration,
      }));
    });
  }

  connect(token?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.socket?.connected || this.isConnecting) {
        resolve();
        return;
      }

      if (!env.ENABLE_WEBSOCKET) {
        console.log('WebSocket is disabled');
        resolve();
        return;
      }

      this.isConnecting = true;

      try {
        this.socket = io(env.WS_URL, {
          auth: token ? { token } : undefined,
          transports: ['websocket'],
          reconnection: true,
          reconnectionAttempts: this.maxReconnectAttempts,
          reconnectionDelay: this.reconnectDelay,
          reconnectionDelayMax: 10000,
        });

        this.socket.on('connect', () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          this.emit('connected', { socketId: this.socket?.id });
          resolve();
        });

        this.socket.on('disconnect', (reason) => {
          console.log('WebSocket disconnected:', reason);
          this.isConnecting = false;
          this.emit('disconnected', { reason });
        });

        this.socket.on('connect_error', (error) => {
          console.error('WebSocket connection error:', error);
          this.isConnecting = false;
          
          if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            this.emit('error', { 
              message: 'Failed to connect to WebSocket server',
              error 
            });
            reject(error);
          }
          
          this.reconnectAttempts++;
        });

        this.socket.on('error', (error) => {
          console.error('WebSocket error:', error);
          this.emit('error', { error });
        });

        // Listen for all events and dispatch to handlers
        this.socket.onAny((event: string, ...args: any[]) => {
          this.handleEvent(event, args[0]);
        });

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  disconnect() {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
  }

  on(event: string, handler: WebSocketEventHandler) {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  off(event: string, handler: WebSocketEventHandler) {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.eventHandlers.delete(event);
      }
    }
  }

  emit(event: string, data: any) {
    if (this.socket?.connected) {
      this.socket.emit(event, data);
    } else {
      console.warn('WebSocket not connected, cannot emit:', event);
    }
  }

  private handleEvent(event: string, data: any) {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${event}:`, error);
        }
      });
    }
  }

  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  getSocketId(): string | undefined {
    return this.socket?.id;
  }

  // Specific methods for common operations
  subscribeToExperiment(experimentId: string) {
    this.emit('experiment:subscribe', { experimentId });
  }

  unsubscribeFromExperiment(experimentId: string) {
    this.emit('experiment:unsubscribe', { experimentId });
  }

  subscribeToPipeline(pipelineId: string) {
    this.emit('pipeline:subscribe', { pipelineId });
  }

  unsubscribeFromPipeline(pipelineId: string) {
    this.emit('pipeline:unsubscribe', { pipelineId });
  }

  sendMessage(type: string, payload: any) {
    this.emit('message', {
      type,
      payload,
      timestamp: Date.now(),
    });
  }
}

// Export singleton instance
export const webSocketService = new WebSocketService();

export default webSocketService;