import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useExperimentStore } from '../useExperimentStore'
import * as apiService from '../../services/api.service'
import { mockExperiment } from '@/test/utils'

// Mock the API service
vi.mock('../../services/api.service', () => ({
  apiService: {
    getExperiments: vi.fn(),
    getExperiment: vi.fn(),
    createExperiment: vi.fn(),
    updateExperiment: vi.fn(),
    deleteExperiment: vi.fn(),
    startExperiment: vi.fn(),
    stopExperiment: vi.fn(),
    promoteVariant: vi.fn(),
    getExperimentAnalysis: vi.fn(),
    getExperimentMetrics: vi.fn(),
  },
}))

describe('useExperimentStore', () => {
  beforeEach(() => {
    // Clear store state
    useExperimentStore.setState({
      experiments: [],
      currentExperiment: null,
      loading: false,
      error: null,
    })
    
    // Reset all mocks
    vi.clearAllMocks()
  })

  describe('fetchExperiments', () => {
    it('fetches and stores experiments', async () => {
      const mockExperiments = [
        mockExperiment({ id: 'exp-1' }),
        mockExperiment({ id: 'exp-2' }),
      ]

      vi.mocked(apiService.apiService.getExperiments).mockResolvedValue({
        experiments: mockExperiments,
        total: 2,
      })

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.fetchExperiments()
      })

      expect(result.current.experiments).toHaveLength(2)
      expect(result.current.experiments[0].id).toBe('exp-1')
      expect(result.current.loading).toBe(false)
      expect(result.current.error).toBeNull()
    })

    it('handles fetch error', async () => {
      vi.mocked(apiService.apiService.getExperiments).mockRejectedValue(
        new Error('Network error')
      )

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.fetchExperiments()
      })

      expect(result.current.experiments).toHaveLength(0)
      expect(result.current.error).toBe('Network error')
      expect(result.current.loading).toBe(false)
    })
  })

  describe('fetchExperiment', () => {
    it('fetches single experiment and sets as current', async () => {
      const experiment = mockExperiment({ id: 'exp-123' })
      
      vi.mocked(apiService.apiService.getExperiment).mockResolvedValue(experiment)

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.fetchExperiment('exp-123')
      })

      expect(result.current.currentExperiment).toEqual(experiment)
      expect(apiService.apiService.getExperiment).toHaveBeenCalledWith('exp-123')
    })

    it('updates experiment in list when fetched', async () => {
      const oldExperiment = mockExperiment({ id: 'exp-123', status: 'pending' })
      const updatedExperiment = mockExperiment({ id: 'exp-123', status: 'running' })
      
      // Set initial state with old experiment
      useExperimentStore.setState({
        experiments: [oldExperiment],
      })

      vi.mocked(apiService.apiService.getExperiment).mockResolvedValue(updatedExperiment)

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.fetchExperiment('exp-123')
      })

      expect(result.current.experiments[0].status).toBe('running')
    })
  })

  describe('createExperiment', () => {
    it('creates experiment and adds to store', async () => {
      const newExperiment = mockExperiment({ id: 'new-exp' })
      const createData = {
        name: 'New Experiment',
        description: 'Test',
        spec: newExperiment.spec,
      }

      vi.mocked(apiService.apiService.createExperiment).mockResolvedValue(newExperiment)

      const { result } = renderHook(() => useExperimentStore())

      let createdExperiment
      await act(async () => {
        createdExperiment = await result.current.createExperiment(createData)
      })

      expect(createdExperiment).toEqual(newExperiment)
      expect(result.current.experiments).toContainEqual(newExperiment)
      expect(result.current.currentExperiment).toEqual(newExperiment)
    })

    it('handles creation error', async () => {
      vi.mocked(apiService.apiService.createExperiment).mockRejectedValue(
        new Error('Creation failed')
      )

      const { result } = renderHook(() => useExperimentStore())

      await expect(
        act(async () => {
          await result.current.createExperiment({
            name: 'Test',
            description: 'Test',
            spec: {} as any,
          })
        })
      ).rejects.toThrow('Creation failed')

      expect(result.current.error).toBe('Creation failed')
    })
  })

  describe('deleteExperiment', () => {
    it('deletes experiment from store', async () => {
      const experiments = [
        mockExperiment({ id: 'exp-1' }),
        mockExperiment({ id: 'exp-2' }),
      ]

      useExperimentStore.setState({
        experiments,
        currentExperiment: experiments[0],
      })

      vi.mocked(apiService.apiService.deleteExperiment).mockResolvedValue(undefined)

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.deleteExperiment('exp-1')
      })

      expect(result.current.experiments).toHaveLength(1)
      expect(result.current.experiments[0].id).toBe('exp-2')
      expect(result.current.currentExperiment).toBeNull()
    })
  })

  describe('startExperiment', () => {
    it('starts experiment and updates status', async () => {
      const experiment = mockExperiment({ id: 'exp-123', status: 'pending' })
      
      useExperimentStore.setState({
        experiments: [experiment],
      })

      vi.mocked(apiService.apiService.startExperiment).mockResolvedValue({})
      vi.mocked(apiService.apiService.getExperiment).mockResolvedValue({
        ...experiment,
        status: 'running',
      })

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.startExperiment('exp-123')
      })

      // Should update status immediately
      const updatedExp = result.current.experiments.find(e => e.id === 'exp-123')
      expect(updatedExp?.status).toBe('initializing')
      
      // Should have a startedAt timestamp
      expect(updatedExp?.startedAt).toBeDefined()
    })
  })

  describe('stopExperiment', () => {
    it('stops experiment and updates status', async () => {
      const experiment = mockExperiment({ id: 'exp-123', status: 'running' })
      
      useExperimentStore.setState({
        experiments: [experiment],
      })

      vi.mocked(apiService.apiService.stopExperiment).mockResolvedValue({})
      vi.mocked(apiService.apiService.getExperiment).mockResolvedValue({
        ...experiment,
        status: 'cancelled',
      })

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.stopExperiment('exp-123')
      })

      const updatedExp = result.current.experiments.find(e => e.id === 'exp-123')
      expect(updatedExp?.status).toBe('cancelled')
      expect(updatedExp?.completedAt).toBeDefined()
    })
  })

  describe('promoteVariant', () => {
    it('promotes variant and updates experiment', async () => {
      const experiment = mockExperiment({ id: 'exp-123', status: 'completed' })
      
      useExperimentStore.setState({
        experiments: [experiment],
      })

      vi.mocked(apiService.apiService.promoteVariant).mockResolvedValue({})
      vi.mocked(apiService.apiService.getExperiment).mockResolvedValue({
        ...experiment,
        status: 'completed',
      })

      const { result } = renderHook(() => useExperimentStore())

      await act(async () => {
        await result.current.promoteVariant('exp-123', 'candidate')
      })

      expect(apiService.apiService.promoteVariant).toHaveBeenCalledWith(
        'exp-123',
        'candidate'
      )
    })
  })

  describe('analysis and metrics', () => {
    it('fetches experiment analysis', async () => {
      const mockAnalysis = {
        status: 'completed',
        comparison: {
          cardinalityReduction: 65,
          costSavings: 58,
        },
      }

      vi.mocked(apiService.apiService.getExperimentAnalysis).mockResolvedValue(mockAnalysis)

      const { result } = renderHook(() => useExperimentStore())

      let analysis
      await act(async () => {
        analysis = await result.current.fetchExperimentAnalysis('exp-123')
      })

      expect(analysis).toEqual(mockAnalysis)
    })

    it('fetches experiment metrics', async () => {
      const mockMetrics = {
        baseline: { cardinality: 150000 },
        candidate: { cardinality: 52500 },
      }

      vi.mocked(apiService.apiService.getExperimentMetrics).mockResolvedValue(mockMetrics)

      const { result } = renderHook(() => useExperimentStore())

      let metrics
      await act(async () => {
        metrics = await result.current.fetchExperimentMetrics('exp-123')
      })

      expect(metrics).toEqual(mockMetrics)
    })
  })

  describe('UI state management', () => {
    it('sets current experiment', () => {
      const experiment = mockExperiment()
      const { result } = renderHook(() => useExperimentStore())

      act(() => {
        result.current.setCurrentExperiment(experiment)
      })

      expect(result.current.currentExperiment).toEqual(experiment)
    })

    it('clears error', () => {
      useExperimentStore.setState({ error: 'Some error' })

      const { result } = renderHook(() => useExperimentStore())

      act(() => {
        result.current.clearError()
      })

      expect(result.current.error).toBeNull()
    })
  })
})