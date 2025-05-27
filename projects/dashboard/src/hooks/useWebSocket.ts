import { useEffect, useCallback, useState } from 'react'
import { useAuthStore } from '../store/useAuthStore'
import { webSocketService } from '../services/websocket/WebSocketService'

interface UseWebSocketReturn {
  connected: boolean
  connect: () => void
  disconnect: () => void
  send: (event: string, data: any) => void
  subscribe: (event: string, handler: (data: any) => void) => () => void
}

export const useWebSocket = (): UseWebSocketReturn => {
  const { token } = useAuthStore()
  const [connected, setConnected] = useState(false)

  const connect = useCallback(() => {
    if (webSocketService.isConnected()) return
    
    webSocketService.connect(token).catch(err => {
      console.error('Failed to connect WebSocket:', err)
    })
  }, [token])

  const disconnect = useCallback(() => {
    webSocketService.disconnect()
  }, [])

  const send = useCallback((event: string, data: any) => {
    webSocketService.sendMessage(event, data)
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    webSocketService.on(event, handler)

    // Return unsubscribe function
    return () => {
      webSocketService.off(event, handler)
    }
  }, [])

  // Track connection state
  useEffect(() => {
    const handleConnected = () => setConnected(true)
    const handleDisconnected = () => setConnected(false)
    
    webSocketService.on('connected', handleConnected)
    webSocketService.on('disconnected', handleDisconnected)
    
    // Check initial state
    setConnected(webSocketService.isConnected())
    
    return () => {
      webSocketService.off('connected', handleConnected)
      webSocketService.off('disconnected', handleDisconnected)
    }
  }, [])

  // Auto-connect when token is available
  useEffect(() => {
    if (token) {
      connect()
    } else {
      disconnect()
    }
  }, [token, connect, disconnect])

  return {
    connected,
    connect,
    disconnect,
    send,
    subscribe,
  }
}

// Global WebSocket hook for singleton instance
export const useGlobalWebSocket = (): UseWebSocketReturn => {
  const { token } = useAuthStore()
  const [connected, setConnected] = useState(false)

  const connect = useCallback(() => {
    if (webSocketService.isConnected()) return
    
    webSocketService.connect(token).catch(err => {
      console.error('Failed to connect WebSocket:', err)
    })
  }, [token])

  const disconnect = useCallback(() => {
    webSocketService.disconnect()
  }, [])

  const send = useCallback((event: string, data: any) => {
    webSocketService.sendMessage(event, data)
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    webSocketService.on(event, handler)
    return () => {
      webSocketService.off(event, handler)
    }
  }, [])

  // Track connection state
  useEffect(() => {
    const handleConnected = () => setConnected(true)
    const handleDisconnected = () => setConnected(false)
    
    webSocketService.on('connected', handleConnected)
    webSocketService.on('disconnected', handleDisconnected)
    
    // Check initial state
    setConnected(webSocketService.isConnected())
    
    return () => {
      webSocketService.off('connected', handleConnected)
      webSocketService.off('disconnected', handleDisconnected)
    }
  }, [])

  useEffect(() => {
    if (token && !webSocketService.isConnected()) {
      connect()
    }
  }, [token, connect])

  return {
    connected,
    connect,
    disconnect,
    send,
    subscribe,
  }
}