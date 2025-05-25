import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@/test/utils'
import { PrivateRoute } from '../PrivateRoute'
import { useAuthStore } from '@/store/useAuthStore'
import { Navigate } from 'react-router-dom'

// Mock the auth store
vi.mock('@/store/useAuthStore')

// Mock Navigate component
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    Navigate: vi.fn(() => null),
  }
})

describe('PrivateRoute', () => {
  it('renders children when authenticated', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: true,
      loading: false,
    } as any)

    render(
      <PrivateRoute>
        <div>Protected Content</div>
      </PrivateRoute>
    )

    expect(screen.getByText('Protected Content')).toBeInTheDocument()
    expect(Navigate).not.toHaveBeenCalled()
  })

  it('redirects to login when not authenticated', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      loading: false,
    } as any)

    render(
      <PrivateRoute>
        <div>Protected Content</div>
      </PrivateRoute>
    )

    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
    expect(Navigate).toHaveBeenCalledWith(
      expect.objectContaining({
        to: '/login',
        replace: true,
      }),
      expect.anything()
    )
  })

  it('shows loading state while checking authentication', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      loading: true,
    } as any)

    render(
      <PrivateRoute>
        <div>Protected Content</div>
      </PrivateRoute>
    )

    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
    expect(Navigate).not.toHaveBeenCalled()
    // Should show loading indicator (CircularProgress)
    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })

  it('preserves location state when redirecting', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      isAuthenticated: false,
      loading: false,
    } as any)

    // Mock current location
    const mockLocation = {
      pathname: '/experiments/123',
      search: '?tab=analysis',
      hash: '',
      state: null,
      key: 'default',
    }

    // Override useLocation
    vi.mock('react-router-dom', async () => {
      const actual = await vi.importActual('react-router-dom')
      return {
        ...actual,
        useLocation: () => mockLocation,
        Navigate: vi.fn(() => null),
      }
    })

    render(
      <PrivateRoute>
        <div>Protected Content</div>
      </PrivateRoute>
    )

    expect(Navigate).toHaveBeenCalledWith(
      expect.objectContaining({
        to: '/login',
        state: { from: mockLocation },
      }),
      expect.anything()
    )
  })
})