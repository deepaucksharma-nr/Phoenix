import { useEffect, useCallback, useState } from 'react'
import { useAuthStore } from '../store/useAuthStore'
import { useExperimentStore } from '../store/useExperimentStore'
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
  const { updateExperiment } = useExperimentStore()
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
    if (webSocketService.isConnected()) {
      // For native WebSocket, we need to send structured messages
      webSocketService.send({
        type: event,
        data: data,
        timestamp: new Date().toISOString()
      })
    } else {
      console.warn('WebSocket not connected')
    }
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    webSocketService.on(event, handler)

    // Return unsubscribe function
    return () => {
      webSocketService.off(event, handler)
    }
  }, [])

  // WebSocketService already handles message routing via event handlers

  // Track connection state
  useEffect(() => {
    const unsubscribeConnected = webSocketService.on('connected', () => setConnected(true))
    const unsubscribeDisconnected = webSocketService.on('disconnected', () => setConnected(false))
    
    // Check initial state
    setConnected(webSocketService.isConnected())
    
    return () => {
      webSocketService.off('connected', unsubscribeConnected)
      webSocketService.off('disconnected', unsubscribeDisconnected)
    }
  }, [])

  // Auto-connect when token is available
  useEffect(() => {
    if (token) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      // Don't disconnect on unmount if still needed elsewhere
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
let globalSocket: Socket | null = null

export const useGlobalWebSocket = (): UseWebSocketReturn => {
  const { token } = useAuthStore()

  const connect = useCallback(() => {
    if (globalSocket?.connected) return

    globalSocket = io(WS_URL, {
      auth: { token },
      transports: ['websocket'],
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionAttempts: 5,
    })

    globalSocket.on('connect', () => {
      console.log('Global WebSocket connected')
    })

    globalSocket.on('disconnect', () => {
      console.log('Global WebSocket disconnected')
    })
  }, [token])

  const disconnect = useCallback(() => {
    if (globalSocket) {
      globalSocket.disconnect()
      globalSocket = null
    }
  }, [])

  const send = useCallback((event: string, data: any) => {
    if (globalSocket?.connected) {
      globalSocket.emit(event, data)
    }
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    if (!globalSocket) {
      return () => {}
    }

    globalSocket.on(event, handler)
    return () => {
      globalSocket?.off(event, handler)
    }
  }, [])

  useEffect(() => {
    if (token && !globalSocket) {
      connect()
    }

    return () => {
      // Don't disconnect on unmount as this is global
    }
  }, [token, connect])

  return {
    connected: globalSocket?.connected || false,
    connect,
    disconnect,
    send,
    subscribe,
  }
}