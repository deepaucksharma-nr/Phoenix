import React, { useState, useEffect, useCallback } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  LinearProgress,
  Box,
  IconButton,
} from '@mui/material'
import { Warning, Close } from '@mui/icons-material'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { logout, checkAuth } from '@store/slices/authSlice'
import { useNavigate } from 'react-router-dom'

interface SessionTimeoutProps {
  // Time before showing warning (in minutes)
  warningTime?: number
  // Total session timeout (in minutes)
  sessionTimeout?: number
  // Whether to enable automatic session extension on activity
  autoExtend?: boolean
}

export const SessionTimeout: React.FC<SessionTimeoutProps> = ({
  warningTime = 5,
  sessionTimeout = 30,
  autoExtend = true,
}) => {
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated)
  const [showWarning, setShowWarning] = useState(false)
  const [timeLeft, setTimeLeft] = useState(warningTime * 60) // Convert to seconds
  const [lastActivity, setLastActivity] = useState(Date.now())

  // Convert minutes to milliseconds
  const warningTimeMs = warningTime * 60 * 1000
  const sessionTimeoutMs = sessionTimeout * 60 * 1000

  // Track user activity
  const updateActivity = useCallback(() => {
    setLastActivity(Date.now())
    if (showWarning) {
      setShowWarning(false)
      setTimeLeft(warningTime * 60)
    }
  }, [showWarning, warningTime])

  // Activity event listeners
  useEffect(() => {
    if (!autoExtend || !isAuthenticated) return

    const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click']
    
    const handleActivity = () => {
      updateActivity()
    }

    // Add event listeners
    events.forEach(event => {
      document.addEventListener(event, handleActivity, true)
    })

    return () => {
      // Clean up event listeners
      events.forEach(event => {
        document.removeEventListener(event, handleActivity, true)
      })
    }
  }, [autoExtend, isAuthenticated, updateActivity])

  // Session timeout logic
  useEffect(() => {
    if (!isAuthenticated) return

    const checkSession = () => {
      const now = Date.now()
      const timeSinceActivity = now - lastActivity
      const timeUntilWarning = sessionTimeoutMs - warningTimeMs - timeSinceActivity
      const timeUntilLogout = sessionTimeoutMs - timeSinceActivity

      if (timeUntilLogout <= 0) {
        // Session expired, force logout
        handleForceLogout()
      } else if (timeUntilWarning <= 0 && !showWarning) {
        // Show warning
        setShowWarning(true)
        setTimeLeft(Math.max(0, Math.floor(timeUntilLogout / 1000)))
      }
    }

    const interval = setInterval(checkSession, 1000)
    return () => clearInterval(interval)
  }, [isAuthenticated, lastActivity, sessionTimeoutMs, warningTimeMs, showWarning])

  // Countdown timer for warning dialog
  useEffect(() => {
    if (!showWarning) return

    const interval = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 1) {
          handleForceLogout()
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [showWarning])

  const handleForceLogout = async () => {
    setShowWarning(false)
    await logout()
    navigate('/login', { 
      state: { 
        message: 'Your session has expired. Please log in again.' 
      } 
    })
  }

  const handleExtendSession = async () => {
    try {
      await checkAuth()
      updateActivity()
      setShowWarning(false)
    } catch (error) {
      handleForceLogout()
    }
  }

  const handleLogoutNow = () => {
    handleForceLogout()
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const progressValue = ((warningTime * 60 - timeLeft) / (warningTime * 60)) * 100

  if (!isAuthenticated || !showWarning) {
    return null
  }

  return (
    <Dialog
      open={showWarning}
      disableEscapeKeyDown
      PaperProps={{
        sx: {
          minWidth: 400,
          maxWidth: 500,
        },
      }}
    >
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1, pb: 1 }}>
        <Warning color="warning" />
        Session Timeout Warning
      </DialogTitle>

      <DialogContent>
        <Typography variant="body1" gutterBottom>
          Your session will expire soon due to inactivity. You will be automatically 
          logged out in:
        </Typography>

        <Box sx={{ my: 3, textAlign: 'center' }}>
          <Typography variant="h4" color="warning.main" fontWeight="bold">
            {formatTime(timeLeft)}
          </Typography>
          <LinearProgress 
            variant="determinate" 
            value={progressValue} 
            color="warning"
            sx={{ mt: 2, height: 8, borderRadius: 1 }}
          />
        </Box>

        <Typography variant="body2" color="text.secondary">
          To continue your session, click "Stay Logged In" below or simply 
          interact with the page.
        </Typography>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 3 }}>
        <Button
          onClick={handleLogoutNow}
          color="error"
          variant="outlined"
        >
          Logout Now
        </Button>
        <Button
          onClick={handleExtendSession}
          color="primary"
          variant="contained"
          autoFocus
        >
          Stay Logged In
        </Button>
      </DialogActions>
    </Dialog>
  )
}

// Hook for session management
export const useSessionTimeout = (options?: {
  warningTime?: number
  sessionTimeout?: number
  autoExtend?: boolean
}) => {
  return {
    SessionTimeoutComponent: () => <SessionTimeout {...options} />,
  }
}