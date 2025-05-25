import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useAuthStore } from '../useAuthStore'
import * as apiService from '../../services/api.service'

// Mock the API service
vi.mock('../../services/api.service', () => ({
  apiService: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    getCurrentUser: vi.fn(),
    requestPasswordReset: vi.fn(),
    resetPassword: vi.fn(),
  },
}))

describe('useAuthStore', () => {
  beforeEach(() => {
    // Clear store state
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
      loading: false,
      error: null,
    })
    
    // Clear localStorage
    localStorage.clear()
    
    // Reset all mocks
    vi.clearAllMocks()
  })

  describe('login', () => {
    it('successfully logs in user', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        role: 'admin',
      }
      const mockToken = 'mock-jwt-token'

      vi.mocked(apiService.apiService.login).mockResolvedValue({
        user: mockUser,
        token: mockToken,
      })

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.login('test@example.com', 'password123')
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.token).toBe(mockToken)
      expect(result.current.isAuthenticated).toBe(true)
      expect(result.current.loading).toBe(false)
      expect(localStorage.getItem('auth_token')).toBe(mockToken)
    })

    it('handles login failure', async () => {
      const mockError = new Error('Invalid credentials')
      vi.mocked(apiService.apiService.login).mockRejectedValue(mockError)

      const { result } = renderHook(() => useAuthStore())

      await expect(
        act(async () => {
          await result.current.login('test@example.com', 'wrong-password')
        })
      ).rejects.toThrow()

      expect(result.current.user).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
      expect(result.current.error).toBe('Login failed')
      expect(result.current.loading).toBe(false)
    })
  })

  describe('register', () => {
    it('successfully registers new user', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'new@example.com',
        name: 'New User',
        role: 'user',
      }
      const mockToken = 'mock-jwt-token'

      vi.mocked(apiService.apiService.register).mockResolvedValue({
        user: mockUser,
        token: mockToken,
      })

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.register({
          name: 'New User',
          email: 'new@example.com',
          password: 'password123',
          organization: 'Test Org',
        })
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.token).toBe(mockToken)
      expect(result.current.isAuthenticated).toBe(true)
      expect(localStorage.getItem('auth_token')).toBe(mockToken)
    })
  })

  describe('logout', () => {
    it('clears authentication state', async () => {
      // Set initial authenticated state
      useAuthStore.setState({
        user: { id: '123', email: 'test@example.com', name: 'Test', role: 'user' },
        token: 'mock-token',
        isAuthenticated: true,
      })
      localStorage.setItem('auth_token', 'mock-token')

      vi.mocked(apiService.apiService.logout).mockResolvedValue(undefined)

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.logout()
      })

      expect(result.current.user).toBeNull()
      expect(result.current.token).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
      expect(localStorage.getItem('auth_token')).toBeNull()
    })
  })

  describe('checkAuth', () => {
    it('validates existing token', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        role: 'admin',
      }
      
      localStorage.setItem('auth_token', 'valid-token')
      vi.mocked(apiService.apiService.getCurrentUser).mockResolvedValue(mockUser)

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.checkAuth()
      })

      expect(result.current.user).toEqual(mockUser)
      expect(result.current.isAuthenticated).toBe(true)
    })

    it('handles invalid token', async () => {
      localStorage.setItem('auth_token', 'invalid-token')
      vi.mocked(apiService.apiService.getCurrentUser).mockRejectedValue(
        new Error('Unauthorized')
      )

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.checkAuth()
      })

      expect(result.current.user).toBeNull()
      expect(result.current.isAuthenticated).toBe(false)
      expect(localStorage.getItem('auth_token')).toBeNull()
    })

    it('handles missing token', async () => {
      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.checkAuth()
      })

      expect(result.current.isAuthenticated).toBe(false)
      expect(apiService.apiService.getCurrentUser).not.toHaveBeenCalled()
    })
  })

  describe('password reset', () => {
    it('requests password reset', async () => {
      vi.mocked(apiService.apiService.requestPasswordReset).mockResolvedValue({})

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.requestPasswordReset('test@example.com')
      })

      expect(apiService.apiService.requestPasswordReset).toHaveBeenCalledWith(
        'test@example.com'
      )
      expect(result.current.loading).toBe(false)
    })

    it('resets password with code', async () => {
      vi.mocked(apiService.apiService.resetPassword).mockResolvedValue({})

      const { result } = renderHook(() => useAuthStore())

      await act(async () => {
        await result.current.resetPassword(
          'test@example.com',
          '123456',
          'newPassword123'
        )
      })

      expect(apiService.apiService.resetPassword).toHaveBeenCalledWith(
        'test@example.com',
        '123456',
        'newPassword123'
      )
      expect(result.current.loading).toBe(false)
    })
  })

  describe('error handling', () => {
    it('clears error', () => {
      useAuthStore.setState({ error: 'Some error' })

      const { result } = renderHook(() => useAuthStore())

      act(() => {
        result.current.clearError()
      })

      expect(result.current.error).toBeNull()
    })
  })
})