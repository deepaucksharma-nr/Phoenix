import React, { createContext, useContext, useEffect, useState } from 'react'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { checkAuth } from '@store/slices/authSlice'
import { Box, CircularProgress, Typography } from '@mui/material'

interface AuthContextType {
  isAuthenticated: boolean
  user: any
  loading: boolean
  error: string | null
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

interface AuthProviderProps {
  children: React.ReactNode
  showLoadingScreen?: boolean
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ 
  children, 
  showLoadingScreen = true 
}) => {
  const dispatch = useAppDispatch()
  const { isAuthenticated, user, loading, error } = useAppSelector(state => state.auth)
  const [isInitialized, setIsInitialized] = useState(false)

  useEffect(() => {
    const initAuth = async () => {
      try {
        await dispatch(checkAuth())
      } finally {
        setIsInitialized(true)
      }
    }

    if (!isInitialized && !loading) {
      initAuth()
    }
  }, [dispatch, isInitialized, loading])

  // Show loading screen during initial authentication check
  if (!isInitialized && showLoadingScreen) {
    return (
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          gap: 2,
        }}
      >
        <CircularProgress size={48} />
        <Typography variant="body2" color="text.secondary">
          Loading Phoenix Platform...
        </Typography>
      </Box>
    )
  }

  const contextValue: AuthContextType = {
    isAuthenticated,
    user,
    loading,
    error,
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuthContext = (): AuthContextType => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuthContext must be used within an AuthProvider')
  }
  return context
}

// Higher-order component for automatic auth wrapping
export const withAuth = <P extends object>(
  Component: React.ComponentType<P>
): React.FC<P> => {
  return (props: P) => (
    <AuthProvider>
      <Component {...props} />
    </AuthProvider>
  )
}