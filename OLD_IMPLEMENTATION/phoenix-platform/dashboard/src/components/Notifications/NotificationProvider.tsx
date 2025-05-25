import React, { createContext, useContext, useState, useCallback } from 'react'
import { Snackbar, Alert, AlertTitle, IconButton } from '@mui/material'
import { Close } from '@mui/icons-material'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
  timestamp: Date
}

interface NotificationContextType {
  notifications: Notification[]
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
  showNotification: (message: string, type?: 'success' | 'error' | 'warning' | 'info') => void
}

export const NotificationContext = createContext<NotificationContextType | undefined>(undefined)

export const useNotifications = () => {
  const context = useContext(NotificationContext)
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider')
  }
  return context
}

interface NotificationProviderProps {
  children: React.ReactNode
}

export const NotificationProvider: React.FC<NotificationProviderProps> = ({ children }) => {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [activeNotification, setActiveNotification] = useState<Notification | null>(null)

  const addNotification = useCallback((notification: Omit<Notification, 'id' | 'timestamp'>) => {
    const newNotification: Notification = {
      ...notification,
      id: `${Date.now()}-${Math.random()}`,
      timestamp: new Date(),
      duration: notification.duration || 5000,
    }

    setNotifications((prev) => [...prev, newNotification])
    
    // Show the notification immediately if none is active
    if (!activeNotification) {
      setActiveNotification(newNotification)
    }
  }, [activeNotification])

  const removeNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
    if (activeNotification?.id === id) {
      setActiveNotification(null)
    }
  }, [activeNotification])

  const clearNotifications = useCallback(() => {
    setNotifications([])
    setActiveNotification(null)
  }, [])

  const showNotification = useCallback((message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') => {
    addNotification({
      title: message,
      type,
    })
  }, [addNotification])

  const handleClose = () => {
    if (activeNotification) {
      removeNotification(activeNotification.id)
      
      // Show next notification if any
      const remaining = notifications.filter((n) => n.id !== activeNotification.id)
      if (remaining.length > 0) {
        setTimeout(() => {
          setActiveNotification(remaining[0])
        }, 100)
      }
    }
  }

  return (
    <NotificationContext.Provider
      value={{ notifications, addNotification, removeNotification, clearNotifications, showNotification }}
    >
      {children}
      {activeNotification && (
        <Snackbar
          open={true}
          autoHideDuration={activeNotification.duration}
          onClose={handleClose}
          anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
        >
          <Alert
            severity={activeNotification.type}
            onClose={handleClose}
            action={
              <IconButton
                size="small"
                aria-label="close"
                color="inherit"
                onClick={handleClose}
              >
                <Close fontSize="small" />
              </IconButton>
            }
            sx={{ minWidth: 300 }}
          >
            <AlertTitle>{activeNotification.title}</AlertTitle>
            {activeNotification.message}
          </Alert>
        </Snackbar>
      )}
    </NotificationContext.Provider>
  )
}

// Hook to connect WebSocket notifications to the notification system
export const useWebSocketNotifications = () => {
  const { addNotification } = useNotifications()
  
  React.useEffect(() => {
    // This will be called from useSystemNotifications hook
    const handleSystemAlert = (payload: {
      level: 'info' | 'warning' | 'error'
      title: string
      message: string
    }) => {
      addNotification({
        type: payload.level === 'info' ? 'info' : payload.level,
        title: payload.title,
        message: payload.message,
      })
    }

    // Export for use in WebSocket hooks
    (window as any).__handleSystemAlert = handleSystemAlert

    return () => {
      delete (window as any).__handleSystemAlert
    }
  }, [addNotification])
}