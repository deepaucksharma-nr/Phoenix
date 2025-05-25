import React from 'react'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { apiService } from '../services/api.service'

interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'user' | 'viewer'
}

interface RegisterData {
  name: string
  email: string
  password: string
  organization: string
}

interface AuthState {
  // State
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null

  // Actions
  login: (email: string, password: string) => Promise<void>
  register: (data: RegisterData) => Promise<void>
  logout: () => Promise<void>
  checkAuth: () => Promise<void>
  requestPasswordReset: (email: string) => Promise<void>
  resetPassword: (email: string, code: string, newPassword: string) => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      token: null,
      isAuthenticated: false,
      loading: false,
      error: null,

      // Login
      login: async (email: string, password: string) => {
        set({ loading: true, error: null })
        try {
          const response = await apiService.login(email, password)
          const { user, token } = response

          // Store token
          localStorage.setItem('auth_token', token)

          set({
            user,
            token,
            isAuthenticated: true,
            loading: false,
          })
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Login failed',
            loading: false,
          })
          throw error
        }
      },

      // Logout
      logout: async () => {
        set({ loading: true })
        try {
          await apiService.logout()
        } catch (error) {
          // Ignore logout errors
        } finally {
          // Clear auth state
          localStorage.removeItem('auth_token')
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            loading: false,
            error: null,
          })
        }
      },

      // Check authentication status
      checkAuth: async () => {
        const token = localStorage.getItem('auth_token')
        if (!token) {
          set({ isAuthenticated: false })
          return
        }

        set({ loading: true })
        try {
          const user = await apiService.getCurrentUser()
          set({
            user,
            token,
            isAuthenticated: true,
            loading: false,
          })
        } catch (error) {
          // Token invalid or expired
          localStorage.removeItem('auth_token')
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            loading: false,
          })
        }
      },

      // Register
      register: async (data: RegisterData) => {
        set({ loading: true, error: null })
        try {
          const response = await apiService.register(data)
          const { user, token } = response

          // Store token
          localStorage.setItem('auth_token', token)

          set({
            user,
            token,
            isAuthenticated: true,
            loading: false,
          })
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Registration failed',
            loading: false,
          })
          throw error
        }
      },

      // Request password reset
      requestPasswordReset: async (email: string) => {
        set({ loading: true, error: null })
        try {
          await apiService.requestPasswordReset(email)
          set({ loading: false })
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Failed to send reset email',
            loading: false,
          })
          throw error
        }
      },

      // Reset password
      resetPassword: async (email: string, code: string, newPassword: string) => {
        set({ loading: true, error: null })
        try {
          await apiService.resetPassword(email, code, newPassword)
          set({ loading: false })
        } catch (error: any) {
          set({
            error: error.response?.data?.message || 'Failed to reset password',
            loading: false,
          })
          throw error
        }
      },

      // Clear error
      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'auth-store',
      partialize: (state) => ({
        token: state.token,
      }),
    }
  )
)

// Hook for checking authentication
export const useAuth = () => {
  const { isAuthenticated, user, checkAuth } = useAuthStore()
  return { isAuthenticated, user, checkAuth }
}

// Hook for requiring authentication
export const useRequireAuth = () => {
  const { isAuthenticated, loading, checkAuth } = useAuthStore()

  // Check auth on mount
  React.useEffect(() => {
    if (!isAuthenticated && !loading) {
      checkAuth()
    }
  }, [isAuthenticated, loading, checkAuth])

  return { isAuthenticated, loading }
}