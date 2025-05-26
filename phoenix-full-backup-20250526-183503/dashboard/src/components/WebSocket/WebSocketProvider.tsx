import React, { createContext, useContext, useEffect, useState } from 'react'
import { Alert, Snackbar, IconButton } from '@mui/material'
import { Close, Wifi, WifiOff } from '@mui/icons-material'
import { useGlobalWebSocket } from '../../hooks/useWebSocket'
import { useAppSelector } from '@hooks/redux'

interface WebSocketContextType {
  connected: boolean
  reconnecting: boolean
  lastConnected: Date | null
  connectionAttempts: number
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined)

interface WebSocketProviderProps {
  children: React.ReactNode
  showConnectionStatus?: boolean
  autoReconnect?: boolean
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
  showConnectionStatus = true,
  autoReconnect = true,
}) => {
  const { connected, connect, disconnect, subscribe } = useGlobalWebSocket()
  const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated)
  const [reconnecting, setReconnecting] = useState(false)
  const [lastConnected, setLastConnected] = useState<Date | null>(null)
  const [connectionAttempts, setConnectionAttempts] = useState(0)
  const [showNotification, setShowNotification] = useState(false)
  const [notificationMessage, setNotificationMessage] = useState('')
  const [notificationType, setNotificationType] = useState<'error' | 'success' | 'info'>('info')

  useEffect(() => {
    if (!isAuthenticated) {
      disconnect()
      return
    }

    // Subscribe to connection events
    const unsubscribeConnect = subscribe('connect', () => {
      setLastConnected(new Date())
      setReconnecting(false)
      setConnectionAttempts(0)
      
      if (showConnectionStatus) {
        setNotificationMessage('Real-time connection established')
        setNotificationType('success')
        setShowNotification(true)
      }
    })

    const unsubscribeDisconnect = subscribe('disconnect', () => {
      if (showConnectionStatus && lastConnected) {
        setNotificationMessage('Real-time connection lost')
        setNotificationType('error')
        setShowNotification(true)
      }
    })

    const unsubscribeReconnecting = subscribe('reconnecting', (attempt: number) => {
      setReconnecting(true)
      setConnectionAttempts(attempt)
      
      if (showConnectionStatus) {
        setNotificationMessage(`Reconnecting... (attempt ${attempt})`)
        setNotificationType('info')
        setShowNotification(true)
      }
    })

    const unsubscribeReconnectError = subscribe('reconnect_error', () => {
      setReconnecting(false)
      
      if (showConnectionStatus) {
        setNotificationMessage('Failed to reconnect. Please refresh the page.')
        setNotificationType('error')
        setShowNotification(true)
      }
    })

    // Auto-connect if authenticated
    if (autoReconnect && !connected) {
      connect()
    }

    return () => {
      unsubscribeConnect()
      unsubscribeDisconnect()
      unsubscribeReconnecting()
      unsubscribeReconnectError()
    }
  }, [isAuthenticated, autoReconnect, connected, showConnectionStatus, subscribe, connect, disconnect, lastConnected])

  // Auto-reconnect logic
  useEffect(() => {
    if (!isAuthenticated || !autoReconnect) return

    const reconnectInterval = setInterval(() => {
      if (!connected && !reconnecting && connectionAttempts < 10) {
        connect()
      }
    }, 5000) // Try to reconnect every 5 seconds

    return () => clearInterval(reconnectInterval)
  }, [isAuthenticated, connected, reconnecting, connectionAttempts, autoReconnect, connect])

  const handleCloseNotification = () => {
    setShowNotification(false)
  }

  const contextValue: WebSocketContextType = {
    connected,
    reconnecting,
    lastConnected,
    connectionAttempts,
  }

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
      
      {/* Connection Status Notifications */}
      {showConnectionStatus && (
        <Snackbar
          open={showNotification}
          autoHideDuration={notificationType === 'success' ? 3000 : 6000}
          onClose={handleCloseNotification}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
        >
          <Alert
            severity={notificationType}
            onClose={handleCloseNotification}
            icon={connected ? <Wifi /> : <WifiOff />}
            action={
              <IconButton size="small" onClick={handleCloseNotification}>
                <Close fontSize="small" />
              </IconButton>
            }
          >
            {notificationMessage}
          </Alert>
        </Snackbar>
      )}
    </WebSocketContext.Provider>
  )
}

export const useWebSocketContext = (): WebSocketContextType => {
  const context = useContext(WebSocketContext)
  if (context === undefined) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}

// Connection status component
export const ConnectionStatus: React.FC = () => {
  const { connected, reconnecting, lastConnected, connectionAttempts } = useWebSocketContext()

  return (
    <Alert 
      severity={connected ? 'success' : reconnecting ? 'info' : 'error'}
      variant="outlined"
      sx={{ mb: 2 }}
      icon={connected ? <Wifi /> : <WifiOff />}
    >
      {connected && 'Real-time connection active'}
      {reconnecting && `Reconnecting... (attempt ${connectionAttempts})`}
      {!connected && !reconnecting && 'Real-time connection unavailable'}
      {lastConnected && (
        <div style={{ fontSize: '0.8em', marginTop: '4px' }}>
          Last connected: {lastConnected.toLocaleString()}
        </div>
      )}
    </Alert>
  )
}

// Higher-order component for WebSocket integration
export const withWebSocket = <P extends object>(
  Component: React.ComponentType<P>
): React.FC<P & { showConnectionStatus?: boolean }> => {
  return ({ showConnectionStatus = true, ...props }: P & { showConnectionStatus?: boolean }) => (
    <WebSocketProvider showConnectionStatus={showConnectionStatus}>
      <Component {...(props as P)} />
    </WebSocketProvider>
  )
}