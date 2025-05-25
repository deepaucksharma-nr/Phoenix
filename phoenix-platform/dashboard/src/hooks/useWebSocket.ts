import { useEffect, useRef, useCallback } from 'react'
import { io, Socket } from 'socket.io-client'
import { useAuthStore } from '../store/useAuthStore'
import { useExperimentStore } from '../store/useExperimentStore'
import { WebSocketMessage } from '../types'

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'

interface UseWebSocketReturn {
  connected: boolean
  connect: () => void
  disconnect: () => void
  send: (event: string, data: any) => void
  subscribe: (event: string, handler: (data: any) => void) => () => void
}

export const useWebSocket = (): UseWebSocketReturn => {
  const socketRef = useRef<Socket | null>(null)
  const { token } = useAuthStore()
  const { updateExperiment } = useExperimentStore()

  const connect = useCallback(() => {
    if (socketRef.current?.connected) return

    socketRef.current = io(WS_URL, {
      auth: {
        token,
      },
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionAttempts: 5,
    })

    socketRef.current.on('connect', () => {
      console.log('WebSocket connected')
    })

    socketRef.current.on('disconnect', () => {
      console.log('WebSocket disconnected')
    })

    socketRef.current.on('error', (error: any) => {
      console.error('WebSocket error:', error)
    })

    // Handle system messages
    socketRef.current.on('message', (message: WebSocketMessage) => {
      handleMessage(message)
    })
  }, [token])

  const disconnect = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.disconnect()
      socketRef.current = null
    }
  }, [])

  const send = useCallback((event: string, data: any) => {
    if (socketRef.current?.connected) {
      socketRef.current.emit(event, data)
    } else {
      console.warn('WebSocket not connected')
    }
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    if (!socketRef.current) {
      console.warn('WebSocket not initialized')
      return () => {}
    }

    socketRef.current.on(event, handler)

    // Return unsubscribe function
    return () => {
      socketRef.current?.off(event, handler)
    }
  }, [])

  const handleMessage = (message: WebSocketMessage) => {
    switch (message.type) {
      case 'experiment.update':
        // Update experiment in store
        if (message.payload.experiment) {
          updateExperiment(message.payload.experiment.id, message.payload.experiment)
        }
        break

      case 'metrics.update':
        // Handle real-time metrics updates
        console.log('Metrics update:', message.payload)
        // TODO: Update metrics store when implemented
        break

      case 'alert':
        // Handle system alerts
        console.warn('Alert:', message.payload)
        // TODO: Show notification to user
        break

      case 'notification':
        // Handle general notifications
        console.log('Notification:', message.payload)
        // TODO: Add to notification store
        break

      default:
        console.log('Unknown message type:', message.type)
    }
  }

  // Auto-connect when token is available
  useEffect(() => {
    if (token) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      disconnect()
    }
  }, [token, connect, disconnect])

  return {
    connected: socketRef.current?.connected || false,
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