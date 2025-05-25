import React from 'react'
import { Box, Typography, Button, Alert } from '@mui/material'
import { Security, ArrowBack } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store/useAuthStore'

interface RoleGuardProps {
  children: React.ReactNode
  allowedRoles: ('admin' | 'user' | 'viewer')[]
  fallback?: React.ReactNode
  showFallback?: boolean
}

export const RoleGuard: React.FC<RoleGuardProps> = ({
  children,
  allowedRoles,
  fallback,
  showFallback = true,
}) => {
  const navigate = useNavigate()
  const { user } = useAuthStore()

  // If user is not logged in, this should be handled by PrivateRoute
  if (!user) {
    return null
  }

  // Check if user has required role
  const hasPermission = allowedRoles.includes(user.role)

  if (!hasPermission) {
    // Show custom fallback if provided
    if (fallback) {
      return <>{fallback}</>
    }

    // Show default unauthorized message if showFallback is true
    if (showFallback) {
      return (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '60vh',
            textAlign: 'center',
            px: 3,
          }}
        >
          <Security sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
          
          <Typography variant="h5" gutterBottom fontWeight={600}>
            Access Restricted
          </Typography>
          
          <Typography variant="body1" color="text.secondary" sx={{ mb: 3, maxWidth: 400 }}>
            You don't have permission to access this page. This area is restricted to users with{' '}
            {allowedRoles.length === 1 
              ? `${allowedRoles[0]} role`
              : `${allowedRoles.slice(0, -1).join(', ')} or ${allowedRoles.slice(-1)} roles`
            }.
          </Typography>

          <Alert severity="info" sx={{ mb: 3, maxWidth: 400 }}>
            <Typography variant="body2">
              Your current role: <strong>{user.role.toUpperCase()}</strong>
            </Typography>
          </Alert>

          <Box sx={{ display: 'flex', gap: 2 }}>
            <Button
              variant="outlined"
              startIcon={<ArrowBack />}
              onClick={() => navigate(-1)}
            >
              Go Back
            </Button>
            
            <Button
              variant="contained"
              onClick={() => navigate('/dashboard')}
            >
              Go to Dashboard
            </Button>
          </Box>
        </Box>
      )
    }

    // Return null if showFallback is false
    return null
  }

  // User has permission, render children
  return <>{children}</>
}

// Hook for checking permissions
export const usePermissions = () => {
  const { user } = useAuthStore()

  const hasRole = (requiredRoles: ('admin' | 'user' | 'viewer')[]) => {
    if (!user) return false
    return requiredRoles.includes(user.role)
  }

  const isAdmin = () => hasRole(['admin'])
  const isUser = () => hasRole(['admin', 'user'])
  const canView = () => hasRole(['admin', 'user', 'viewer'])

  const canCreateExperiments = () => isUser()
  const canDeleteExperiments = () => isAdmin()
  const canManageUsers = () => isAdmin()
  const canViewExperiments = () => canView()

  return {
    hasRole,
    isAdmin,
    isUser,
    canView,
    canCreateExperiments,
    canDeleteExperiments,
    canManageUsers,
    canViewExperiments,
    user,
  }
}