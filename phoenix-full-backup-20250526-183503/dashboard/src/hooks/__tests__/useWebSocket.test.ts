import { describe, it, expect, beforeEach, vi, Mock } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from '../useWebSocket'
import { useAuthStore } from '../../store/useAuthStore'
import { useExperimentStore } from '../../store/useExperimentStore'
import { io, Socket } from 'socket.io-client'

// Mock socket.io-client
vi.mock('socket.io-client', () => ({
  io: vi.fn(),
}))

// Mock stores
vi.mock('../../store/useAuthStore')
vi.mock('../../store/useExperimentStore')

describe('useWebSocket', () => {
  let mockSocket: {
    connected: boolean
    on: Mock
    off: Mock
    emit: Mock
    disconnect: Mock
  }

  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks()

    // Create mock socket
    mockSocket = {
      connected: false,
      on: vi.fn(),
      off: vi.fn(),
      emit: vi.fn(),
      disconnect: vi.fn(),
    }

    // Mock io to return our mock socket
    vi.mocked(io).mockReturnValue(mockSocket as any)

    // Mock auth store
    vi.mocked(useAuthStore).mockReturnValue({
      token: 'mock-token',
    } as any)

    // Mock experiment store
    vi.mocked(useExperimentStore).mockReturnValue({
      updateExperiment: vi.fn(),
    } as any)
  })

  it('connects when token is available', () => {
    renderHook(() => useWebSocket())

    expect(io).toHaveBeenCalledWith('ws://localhost:8080', {
      auth: { token: 'mock-token' },
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionAttempts: 5,
    })
  })

  it('does not connect when token is missing', () => {
    vi.mocked(useAuthStore).mockReturnValue({
      token: null,
    } as any)

    renderHook(() => useWebSocket())

    expect(io).not.toHaveBeenCalled()
  })

  it('registers event handlers on connect', () => {
    renderHook(() => useWebSocket())

    expect(mockSocket.on).toHaveBeenCalledWith('connect', expect.any(Function))
    expect(mockSocket.on).toHaveBeenCalledWith('disconnect', expect.any(Function))
    expect(mockSocket.on).toHaveBeenCalledWith('error', expect.any(Function))
    expect(mockSocket.on).toHaveBeenCalledWith('message', expect.any(Function))
  })

  it('handles connect event', () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    
    renderHook(() => useWebSocket())

    // Simulate connect event
    const connectHandler = mockSocket.on.mock.calls.find(
      ([event]) => event === 'connect'
    )?.[1]
    
    act(() => {
      connectHandler?.()
    })

    expect(consoleSpy).toHaveBeenCalledWith('WebSocket connected')
    consoleSpy.mockRestore()
  })

  it('sends messages when connected', () => {
    mockSocket.connected = true
    
    const { result } = renderHook(() => useWebSocket())

    act(() => {
      result.current.send('test-event', { data: 'test' })
    })

    expect(mockSocket.emit).toHaveBeenCalledWith('test-event', { data: 'test' })
  })

  it('warns when sending without connection', () => {
    const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
    mockSocket.connected = false
    
    const { result } = renderHook(() => useWebSocket())

    act(() => {
      result.current.send('test-event', { data: 'test' })
    })

    expect(mockSocket.emit).not.toHaveBeenCalled()
    expect(consoleSpy).toHaveBeenCalledWith('WebSocket not connected')
    consoleSpy.mockRestore()
  })

  it('subscribes to events', () => {
    const { result } = renderHook(() => useWebSocket())
    const handler = vi.fn()

    act(() => {
      result.current.subscribe('custom-event', handler)
    })

    expect(mockSocket.on).toHaveBeenCalledWith('custom-event', handler)
  })

  it('returns unsubscribe function', () => {
    const { result } = renderHook(() => useWebSocket())
    const handler = vi.fn()

    let unsubscribe: () => void
    act(() => {
      unsubscribe = result.current.subscribe('custom-event', handler)
    })

    act(() => {
      unsubscribe()
    })

    expect(mockSocket.off).toHaveBeenCalledWith('custom-event', handler)
  })

  it('handles experiment update messages', () => {
    const updateExperiment = vi.fn()
    vi.mocked(useExperimentStore).mockReturnValue({
      updateExperiment,
    } as any)

    renderHook(() => useWebSocket())

    // Get message handler
    const messageHandler = mockSocket.on.mock.calls.find(
      ([event]) => event === 'message'
    )?.[1]

    const mockMessage = {
      type: 'experiment.update',
      payload: {
        experiment: {
          id: 'exp-123',
          name: 'Updated Experiment',
          status: 'running',
        },
      },
      timestamp: new Date().toISOString(),
    }

    act(() => {
      messageHandler?.(mockMessage)
    })

    expect(updateExperiment).toHaveBeenCalledWith('exp-123', mockMessage.payload.experiment)
  })

  it('handles metrics update messages', () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    
    renderHook(() => useWebSocket())

    const messageHandler = mockSocket.on.mock.calls.find(
      ([event]) => event === 'message'
    )?.[1]

    const mockMessage = {
      type: 'metrics.update',
      payload: { metric: 'cpu', value: 75 },
      timestamp: new Date().toISOString(),
    }

    act(() => {
      messageHandler?.(mockMessage)
    })

    expect(consoleSpy).toHaveBeenCalledWith('Metrics update:', mockMessage.payload)
    consoleSpy.mockRestore()
  })

  it('disconnects when token is removed', () => {
    const { rerender } = renderHook(() => useWebSocket())

    // Initial render with token
    expect(io).toHaveBeenCalled()

    // Update token to null
    vi.mocked(useAuthStore).mockReturnValue({
      token: null,
    } as any)

    rerender()

    expect(mockSocket.disconnect).toHaveBeenCalled()
  })

  it('cleans up on unmount', () => {
    const { unmount } = renderHook(() => useWebSocket())

    unmount()

    expect(mockSocket.disconnect).toHaveBeenCalled()
  })
})