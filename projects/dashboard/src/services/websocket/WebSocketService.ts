// Re-export the native WebSocket implementation
// This maintains backward compatibility while switching to native WebSocket
export { webSocketService } from './NativeWebSocketService';
export type { WebSocketMessage, WebSocketEventHandler } from './NativeWebSocketService';
export { default } from './NativeWebSocketService';