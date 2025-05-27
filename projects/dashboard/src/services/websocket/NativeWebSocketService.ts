import { env } from '@utils/env';
import { store } from '@/store';
import { updateExperiment, setExperimentMetrics } from '@/store/slices/experimentSlice';
import { addNotification } from '@/store/slices/notificationSlice';

export interface WebSocketMessage {
  type: string;
  topic?: string;
  data: any;
  timestamp: string;
}

export type WebSocketEventHandler = (data: any) => void;

class NativeWebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private eventHandlers: Map<string, Set<WebSocketEventHandler>> = new Map();
  private isConnecting = false;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private subscriptions: Set<string> = new Set();

  constructor() {
    this.setupEventHandlers();
  }

  private setupEventHandlers() {
    // Built-in event handlers for Redux integration
    this.on('experiment_update', (data) => {
      if (data.experiment) {
        store.dispatch(updateExperiment(data.experiment));
      }
    });

    this.on('metric_update', (data) => {
      if (data.experiment_id && data.metrics) {
        store.dispatch(setExperimentMetrics({
          experimentId: data.experiment_id,
          metrics: data.metrics,
        }));
      }
    });

    this.on('notification', (data) => {
      store.dispatch(addNotification({
        type: data.severity || 'info',
        message: data.message,
        description: data.description,
        duration: data.duration,
      }));
    });
  }

  connect(token?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) {
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
        // Convert HTTP URL to WebSocket URL
        const wsUrl = env.WS_URL.replace(/^http/, 'ws');
        const url = new URL(`${wsUrl}/api/v1/ws`);
        
        // Add token to query params if provided
        if (token) {
          url.searchParams.set('token', token);
        }

        this.ws = new WebSocket(url.toString());

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          this.emit('connected', { connected: true });
          
          // Re-subscribe to previous topics
          this.subscriptions.forEach(topic => {
            this.sendSubscribe(topic);
          });
          
          // Start heartbeat
          this.startHeartbeat();
          
          resolve();
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.reason);
          this.isConnecting = false;
          this.stopHeartbeat();
          this.emit('disconnected', { reason: event.reason });
          
          // Attempt reconnection if not a normal closure
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.isConnecting = false;
          this.emit('error', { error });
          
          if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            reject(new Error('Failed to connect to WebSocket server'));
          }
          
          this.reconnectAttempts++;
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  private handleMessage(message: WebSocketMessage) {
    // Handle different message types
    switch (message.type) {
      case 'heartbeat':
        // Server heartbeat received
        break;
        
      case 'error':
        console.error('WebSocket server error:', message.data);
        this.emit('error', message.data);
        break;
        
      case 'notification':
        // Handle subscription confirmation
        if (message.data.subscribed !== undefined) {
          console.log(`Subscription ${message.data.subscribed ? 'confirmed' : 'removed'} for topic:`, message.data.topic);
        } else {
          // Regular notification
          this.emit('notification', message.data);
        }
        break;
        
      default:
        // Emit to specific handlers
        this.emit(message.type, message.data);
        
        // Also emit with topic prefix if present
        if (message.topic) {
          this.emit(`${message.topic}:${message.type}`, message.data);
        }
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    
    const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts), 10000);
    console.log(`Reconnecting in ${delay}ms...`);
    
    this.emit('reconnecting', { attempt: this.reconnectAttempts + 1 });
    
    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    
    // Send heartbeat every 30 seconds
    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({
          type: 'heartbeat',
          data: { timestamp: new Date().toISOString() },
          timestamp: new Date().toISOString()
        });
      }
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  disconnect() {
    this.stopHeartbeat();
    
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    this.subscriptions.clear();
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

  private emit(event: string, data: any) {
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

  send(message: WebSocketMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send:', message.type);
    }
  }

  private sendSubscribe(topic: string) {
    this.send({
      type: 'subscribe',
      data: { topic },
      timestamp: new Date().toISOString()
    });
  }

  private sendUnsubscribe(topic: string) {
    this.send({
      type: 'unsubscribe',
      data: { topic },
      timestamp: new Date().toISOString()
    });
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN || false;
  }

  sendMessage(type: string, payload: any) {
    this.send({
      type,
      data: payload,
      timestamp: new Date().toISOString(),
    });
  }

  // Topic subscription methods
  subscribeToExperiment(experimentId: string) {
    const topic = `experiment:${experimentId}`;
    this.subscriptions.add(topic);
    this.sendSubscribe(topic);
  }

  unsubscribeFromExperiment(experimentId: string) {
    const topic = `experiment:${experimentId}`;
    this.subscriptions.delete(topic);
    this.sendUnsubscribe(topic);
  }

  subscribeToPipeline(pipelineId: string) {
    const topic = `pipeline:${pipelineId}`;
    this.subscriptions.add(topic);
    this.sendSubscribe(topic);
  }

  unsubscribeFromPipeline(pipelineId: string) {
    const topic = `pipeline:${pipelineId}`;
    this.subscriptions.delete(topic);
    this.sendUnsubscribe(topic);
  }

  subscribeToMetrics(experimentId: string) {
    const topic = `metrics:${experimentId}`;
    this.subscriptions.add(topic);
    this.sendSubscribe(topic);
  }

  unsubscribeFromMetrics(experimentId: string) {
    const topic = `metrics:${experimentId}`;
    this.subscriptions.delete(topic);
    this.sendUnsubscribe(topic);
  }
}

// Export singleton instance
export const webSocketService = new NativeWebSocketService();

export default webSocketService;