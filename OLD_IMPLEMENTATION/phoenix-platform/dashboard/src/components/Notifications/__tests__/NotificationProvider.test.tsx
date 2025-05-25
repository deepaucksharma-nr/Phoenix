import { describe, it, expect, vi } from 'vitest'
import { render, screen, act, waitFor } from '@/test/utils'
import { NotificationProvider, useNotifications } from '../NotificationProvider'
import { Button } from '@mui/material'

// Test component that uses the notification hook
const TestComponent = () => {
  const { notifications, addNotification, removeNotification, clearNotifications } = useNotifications()

  return (
    <div>
      <div data-testid="notification-count">{notifications.length}</div>
      <Button
        onClick={() =>
          addNotification({
            type: 'success',
            title: 'Test Success',
            message: 'This is a test message',
          })
        }
      >
        Add Success
      </Button>
      <Button
        onClick={() =>
          addNotification({
            type: 'error',
            title: 'Test Error',
            duration: 3000,
          })
        }
      >
        Add Error
      </Button>
      <Button
        onClick={() => {
          if (notifications.length > 0) {
            removeNotification(notifications[0].id)
          }
        }}
      >
        Remove First
      </Button>
      <Button onClick={clearNotifications}>Clear All</Button>
    </div>
  )
}

describe('NotificationProvider', () => {
  it('provides notification context to children', () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    expect(screen.getByTestId('notification-count')).toHaveTextContent('0')
    expect(screen.getByText('Add Success')).toBeInTheDocument()
  })

  it('adds notifications', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    const addButton = screen.getByText('Add Success')
    
    act(() => {
      addButton.click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('notification-count')).toHaveTextContent('1')
      expect(screen.getByText('Test Success')).toBeInTheDocument()
      expect(screen.getByText('This is a test message')).toBeInTheDocument()
    })
  })

  it('displays notifications with correct severity', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    act(() => {
      screen.getByText('Add Error').click()
    })

    await waitFor(() => {
      expect(screen.getByText('Test Error')).toBeInTheDocument()
      expect(screen.getByRole('alert')).toHaveClass('MuiAlert-standardError')
    })
  })

  it('removes notifications', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    // Add a notification
    act(() => {
      screen.getByText('Add Success').click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('notification-count')).toHaveTextContent('1')
    })

    // Remove it
    act(() => {
      screen.getByText('Remove First').click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('notification-count')).toHaveTextContent('0')
    })
  })

  it('clears all notifications', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    // Add multiple notifications
    act(() => {
      screen.getByText('Add Success').click()
      screen.getByText('Add Error').click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('notification-count')).toHaveTextContent('2')
    })

    // Clear all
    act(() => {
      screen.getByText('Clear All').click()
    })

    await waitFor(() => {
      expect(screen.getByTestId('notification-count')).toHaveTextContent('0')
    })
  })

  it('auto-dismisses notifications after duration', async () => {
    vi.useFakeTimers()

    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    act(() => {
      screen.getByText('Add Error').click() // Has 3000ms duration
    })

    expect(screen.getByText('Test Error')).toBeInTheDocument()

    // Fast-forward time
    act(() => {
      vi.advanceTimersByTime(3000)
    })

    await waitFor(() => {
      expect(screen.queryByText('Test Error')).not.toBeInTheDocument()
    })

    vi.useRealTimers()
  })

  it('handles close button click', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    act(() => {
      screen.getByText('Add Success').click()
    })

    await waitFor(() => {
      expect(screen.getByText('Test Success')).toBeInTheDocument()
    })

    // Click close button
    const closeButton = screen.getByLabelText('close')
    act(() => {
      closeButton.click()
    })

    await waitFor(() => {
      expect(screen.queryByText('Test Success')).not.toBeInTheDocument()
    })
  })

  it('queues multiple notifications', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    )

    // Add multiple notifications quickly
    act(() => {
      screen.getByText('Add Success').click()
      screen.getByText('Add Error').click()
    })

    // Should show first notification
    await waitFor(() => {
      expect(screen.getByText('Test Success')).toBeInTheDocument()
    })

    // Close first notification
    act(() => {
      screen.getByLabelText('close').click()
    })

    // Should show second notification
    await waitFor(() => {
      expect(screen.queryByText('Test Success')).not.toBeInTheDocument()
      expect(screen.getByText('Test Error')).toBeInTheDocument()
    })
  })
})