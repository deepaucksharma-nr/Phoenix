// Re-export the native WebSocket implementation
// This maintains backward compatibility while switching to native WebSocket
export { 
  webSocketService, 
  WebSocketMessage, 
  WebSocketEventHandler,
  default 
} from './NativeWebSocketService';

export default webSocketService;