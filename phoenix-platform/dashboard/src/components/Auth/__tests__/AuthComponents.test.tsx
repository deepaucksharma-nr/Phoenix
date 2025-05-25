import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { PrivateRoute } from '../PrivateRoute'
import { RoleGuard, usePermissions } from '../RoleGuard'
import { UserProfile } from '../UserProfile'
import { useAuthStore } from '../../../store/useAuthStore'

// Mock the auth store
vi.mock('../../../store/useAuthStore')

const mockUseAuthStore = useAuthStore as any

// Test wrapper component
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <BrowserRouter>{children}</BrowserRouter>
)

describe('Authentication Components', () => {
  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      isAuthenticated: false,
      loading: false,
      user: null,
      login: vi.fn(),
      logout: vi.fn(),
      checkAuth: vi.fn(),
      clearError: vi.fn(),
      error: null,
    })
  })

  describe('PrivateRoute', () => {
    it('renders loading state when checking auth', () => {
      mockUseAuthStore.mockReturnValue({
        isAuthenticated: false,
        loading: true,
        checkAuth: vi.fn(),
      })

      render(
        <TestWrapper>
          <PrivateRoute>
            <div>Protected Content</div>
          </PrivateRoute>
        </TestWrapper>
      )

      expect(screen.getByText('Verifying authentication...')).toBeInTheDocument()
    })

    it('redirects to login when not authenticated', () => {
      mockUseAuthStore.mockReturnValue({
        isAuthenticated: false,
        loading: false,
        checkAuth: vi.fn(),
      })

      render(
        <TestWrapper>
          <PrivateRoute>
            <div>Protected Content</div>
          </PrivateRoute>
        </TestWrapper>
      )

      // Should not render protected content
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
    })

    it('renders protected content when authenticated', () => {
      mockUseAuthStore.mockReturnValue({
        isAuthenticated: true,
        loading: false,
        checkAuth: vi.fn(),
      })

      render(
        <TestWrapper>
          <PrivateRoute>
            <div>Protected Content</div>
          </PrivateRoute>
        </TestWrapper>
      )

      expect(screen.getByText('Protected Content')).toBeInTheDocument()
    })
  })

  describe('RoleGuard', () => {
    const mockUser = {
      id: '1',
      name: 'Test User',
      email: 'test@example.com',
      role: 'user' as const,
    }

    it('renders content when user has required role', () => {
      mockUseAuthStore.mockReturnValue({
        user: mockUser,
      })

      render(
        <TestWrapper>
          <RoleGuard allowedRoles={['user']}>
            <div>Role Protected Content</div>
          </RoleGuard>
        </TestWrapper>
      )

      expect(screen.getByText('Role Protected Content')).toBeInTheDocument()
    })

    it('shows access denied when user lacks required role', () => {
      mockUseAuthStore.mockReturnValue({
        user: mockUser,
      })

      render(
        <TestWrapper>
          <RoleGuard allowedRoles={['admin']}>
            <div>Admin Only Content</div>
          </RoleGuard>
        </TestWrapper>
      )

      expect(screen.getByText('Access Restricted')).toBeInTheDocument()
      expect(screen.getByText('Your current role: USER')).toBeInTheDocument()
      expect(screen.queryByText('Admin Only Content')).not.toBeInTheDocument()
    })

    it('renders custom fallback when provided', () => {
      mockUseAuthStore.mockReturnValue({
        user: mockUser,
      })

      render(
        <TestWrapper>
          <RoleGuard 
            allowedRoles={['admin']} 
            fallback={<div>Custom Unauthorized Message</div>}
          >
            <div>Admin Only Content</div>
          </RoleGuard>
        </TestWrapper>
      )

      expect(screen.getByText('Custom Unauthorized Message')).toBeInTheDocument()
      expect(screen.queryByText('Access Restricted')).not.toBeInTheDocument()
    })

    it('returns null when showFallback is false', () => {
      mockUseAuthStore.mockReturnValue({
        user: mockUser,
      })

      const { container } = render(
        <TestWrapper>
          <RoleGuard allowedRoles={['admin']} showFallback={false}>
            <div>Admin Only Content</div>
          </RoleGuard>
        </TestWrapper>
      )

      expect(container.firstChild).toBeNull()
    })
  })

  describe('UserProfile', () => {
    const mockUser = {
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin' as const,
    }

    beforeEach(() => {
      mockUseAuthStore.mockReturnValue({
        user: mockUser,
        logout: vi.fn(),
      })
    })

    it('renders user profile menu when user is logged in', () => {
      render(
        <TestWrapper>
          <UserProfile />
        </TestWrapper>
      )

      // Should show avatar with initials
      expect(screen.getByText('JD')).toBeInTheDocument()
    })

    it('renders full profile card when variant is card', () => {
      render(
        <TestWrapper>
          <UserProfile variant="card" showFullProfile />
        </TestWrapper>
      )

      expect(screen.getByText('John Doe')).toBeInTheDocument()
      expect(screen.getByText('john@example.com')).toBeInTheDocument()
      expect(screen.getByText('ADMIN')).toBeInTheDocument()
    })

    it('opens menu when avatar is clicked', async () => {
      render(
        <TestWrapper>
          <UserProfile />
        </TestWrapper>
      )

      // Click avatar
      fireEvent.click(screen.getByText('JD'))

      // Menu should appear
      await waitFor(() => {
        expect(screen.getByText('Profile')).toBeInTheDocument()
        expect(screen.getByText('Settings')).toBeInTheDocument()
        expect(screen.getByText('Logout')).toBeInTheDocument()
      })
    })

    it('returns null when user is not logged in', () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        logout: vi.fn(),
      })

      const { container } = render(
        <TestWrapper>
          <UserProfile />
        </TestWrapper>
      )

      expect(container.firstChild).toBeNull()
    })
  })

  describe('usePermissions hook', () => {
    const TestComponent: React.FC = () => {
      const permissions = usePermissions()
      return (
        <div>
          <div>Is Admin: {permissions.isAdmin().toString()}</div>
          <div>Can Create: {permissions.canCreateExperiments().toString()}</div>
          <div>Can Delete: {permissions.canDeleteExperiments().toString()}</div>
        </div>
      )
    }

    it('returns correct permissions for admin user', () => {
      mockUseAuthStore.mockReturnValue({
        user: {
          id: '1',
          name: 'Admin User',
          email: 'admin@example.com',
          role: 'admin',
        },
      })

      render(
        <TestWrapper>
          <TestComponent />
        </TestWrapper>
      )

      expect(screen.getByText('Is Admin: true')).toBeInTheDocument()
      expect(screen.getByText('Can Create: true')).toBeInTheDocument()
      expect(screen.getByText('Can Delete: true')).toBeInTheDocument()
    })

    it('returns correct permissions for regular user', () => {
      mockUseAuthStore.mockReturnValue({
        user: {
          id: '1',
          name: 'Regular User',
          email: 'user@example.com',
          role: 'user',
        },
      })

      render(
        <TestWrapper>
          <TestComponent />
        </TestWrapper>
      )

      expect(screen.getByText('Is Admin: false')).toBeInTheDocument()
      expect(screen.getByText('Can Create: true')).toBeInTheDocument()
      expect(screen.getByText('Can Delete: false')).toBeInTheDocument()
    })

    it('returns correct permissions for viewer', () => {
      mockUseAuthStore.mockReturnValue({
        user: {
          id: '1',
          name: 'Viewer User',
          email: 'viewer@example.com',
          role: 'viewer',
        },
      })

      render(
        <TestWrapper>
          <TestComponent />
        </TestWrapper>
      )

      expect(screen.getByText('Is Admin: false')).toBeInTheDocument()
      expect(screen.getByText('Can Create: false')).toBeInTheDocument()
      expect(screen.getByText('Can Delete: false')).toBeInTheDocument()
    })
  })
})