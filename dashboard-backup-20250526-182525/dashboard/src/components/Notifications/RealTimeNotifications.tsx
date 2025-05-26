import React, { useState, useEffect } from 'react'
import {
  Snackbar,
  Alert,
  IconButton,
  Badge,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Typography,
  Box,
  Divider,
  Button,
  Paper,
  List,
  ListItem,
  Avatar,
} from '@mui/material'
import {
  Notifications,
  Close,
  Info,
  Warning,
  Error as ErrorIcon,
  CheckCircle,
  Experiment,
  Timeline,
  Delete,
  MarkAsUnread,
  MarkEmailRead,
} from '@mui/icons-material'
import { useWebSocket } from '../../hooks/useWebSocket'
import { useNotification } from '../../hooks/useNotification'

interface NotificationData {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message: string
  timestamp: string
  category: 'experiment' | 'system' | 'metrics' | 'alert'
  experimentId?: string
  read: boolean
  persistent?: boolean
}

interface ToastNotificationProps {
  notification: NotificationData
  onClose: () => void
}

const ToastNotification: React.FC<ToastNotificationProps> = ({ notification, onClose }) => {
  return (
    <Snackbar
      open={true}
      autoHideDuration={notification.persistent ? null : 6000}
      onClose={onClose}
      anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
    >
      <Alert
        severity={notification.type}
        onClose={onClose}
        sx={{ minWidth: 300 }}
        action={
          <IconButton size="small" onClick={onClose}>
            <Close fontSize="small" />
          </IconButton>
        }
      >
        <Typography variant="subtitle2" gutterBottom>
          {notification.title}
        </Typography>
        <Typography variant="body2">
          {notification.message}
        </Typography>
      </Alert>
    </Snackbar>
  )
}

export const RealTimeNotifications: React.FC = () => {
  const { subscribe } = useWebSocket()
  const { showNotification } = useNotification()
  const [notifications, setNotifications] = useState<NotificationData[]>([])
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const [toastQueue, setToastQueue] = useState<NotificationData[]>([])

  const unreadCount = notifications.filter(n => !n.read).length

  useEffect(() => {
    // Subscribe to real-time notifications
    const unsubscribe = subscribe('notification', (data: any) => {
      const notification: NotificationData = {
        id: data.id || Date.now().toString(),
        type: data.type || 'info',
        title: data.title || 'Notification',
        message: data.message || '',
        timestamp: data.timestamp || new Date().toISOString(),
        category: data.category || 'system',
        experimentId: data.experimentId,
        read: false,
        persistent: data.persistent || false,
      }

      // Add to notifications list
      setNotifications(prev => [notification, ...prev])

      // Add to toast queue if it's important
      if (notification.type === 'error' || notification.type === 'warning' || notification.persistent) {
        setToastQueue(prev => [...prev, notification])
      }

      // Show in-app notification for info/success
      if (notification.type === 'info' || notification.type === 'success') {
        showNotification(notification.message, notification.type)
      }
    })

    // Subscribe to experiment updates
    const unsubscribeExperiment = subscribe('experiment.update', (data: any) => {
      const notification: NotificationData = {
        id: `exp-${data.experimentId}-${Date.now()}`,
        type: 'info',
        title: 'Experiment Update',
        message: `Experiment ${data.experimentName || data.experimentId} is now ${data.status}`,
        timestamp: new Date().toISOString(),
        category: 'experiment',
        experimentId: data.experimentId,
        read: false,
      }

      setNotifications(prev => [notification, ...prev])
    })

    // Subscribe to system alerts
    const unsubscribeAlert = subscribe('alert', (data: any) => {
      const notification: NotificationData = {
        id: `alert-${Date.now()}`,
        type: data.severity || 'warning',
        title: 'System Alert',
        message: data.message || 'System alert received',
        timestamp: new Date().toISOString(),
        category: 'alert',
        read: false,
        persistent: true,
      }

      setNotifications(prev => [notification, ...prev])
      setToastQueue(prev => [...prev, notification])
    })

    return () => {
      unsubscribe()
      unsubscribeExperiment()
      unsubscribeAlert()
    }
  }, [subscribe, showNotification])

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleClose = () => {
    setAnchorEl(null)
  }

  const markAsRead = (id: string) => {
    setNotifications(prev =>
      prev.map(n => n.id === id ? { ...n, read: true } : n)
    )
  }

  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  const deleteNotification = (id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  const clearAll = () => {
    setNotifications([])
  }

  const getIcon = (category: string) => {
    switch (category) {
      case 'experiment':
        return <Experiment />
      case 'metrics':
        return <Timeline />
      case 'alert':
        return <Warning />
      default:
        return <Info />
    }
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (days > 0) {
      return `${days}d ago`
    } else if (hours > 0) {
      return `${hours}h ago`
    } else if (minutes > 0) {
      return `${minutes}m ago`
    } else {
      return 'Just now'
    }
  }

  const dismissToast = (id: string) => {
    setToastQueue(prev => prev.filter(n => n.id !== id))
  }

  return (
    <>
      {/* Notification Bell Icon */}
      <IconButton color="inherit" onClick={handleClick}>
        <Badge badgeContent={unreadCount} color="error">
          <Notifications />
        </Badge>
      </IconButton>

      {/* Notification Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
        PaperProps={{
          sx: {
            width: 400,
            maxHeight: 500,
          },
        }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="h6">
              Notifications
            </Typography>
            <Box>
              {unreadCount > 0 && (
                <Button size="small" onClick={markAllAsRead}>
                  Mark all read
                </Button>
              )}
              <Button size="small" color="error" onClick={clearAll}>
                Clear all
              </Button>
            </Box>
          </Box>
        </Box>

        {notifications.length === 0 ? (
          <Box sx={{ p: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              No notifications
            </Typography>
          </Box>
        ) : (
          <List sx={{ p: 0, maxHeight: 400, overflow: 'auto' }}>
            {notifications.slice(0, 20).map((notification) => (
              <ListItem
                key={notification.id}
                sx={{
                  backgroundColor: notification.read ? 'transparent' : 'action.hover',
                  borderLeft: 4,
                  borderLeftColor: notification.type === 'error' ? 'error.main' :
                                  notification.type === 'warning' ? 'warning.main' :
                                  notification.type === 'success' ? 'success.main' : 'info.main',
                }}
              >
                <Box sx={{ display: 'flex', width: '100%', alignItems: 'flex-start', gap: 2 }}>
                  <Avatar sx={{ width: 32, height: 32, bgcolor: 'transparent' }}>
                    {getIcon(notification.category)}
                  </Avatar>
                  
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography variant="subtitle2" noWrap>
                      {notification.title}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                      {notification.message}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {formatTimestamp(notification.timestamp)}
                    </Typography>
                  </Box>

                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                    <IconButton
                      size="small"
                      onClick={() => notification.read ? null : markAsRead(notification.id)}
                      disabled={notification.read}
                    >
                      {notification.read ? <MarkEmailRead /> : <MarkAsUnread />}
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => deleteNotification(notification.id)}
                      color="error"
                    >
                      <Delete />
                    </IconButton>
                  </Box>
                </Box>
              </ListItem>
            ))}
          </List>
        )}
      </Menu>

      {/* Toast Notifications */}
      {toastQueue.map((notification, index) => (
        <Box
          key={notification.id}
          sx={{
            position: 'fixed',
            top: 80 + (index * 80),
            right: 20,
            zIndex: 9999,
          }}
        >
          <ToastNotification
            notification={notification}
            onClose={() => dismissToast(notification.id)}
          />
        </Box>
      ))}
    </>
  )
}