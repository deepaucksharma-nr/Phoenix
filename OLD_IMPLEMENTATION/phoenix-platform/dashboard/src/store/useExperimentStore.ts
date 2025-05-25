import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import { apiService } from '../services/api.service'

export interface Experiment {
  id: string
  name: string
  description: string
  spec?: {
    baseline: any
    candidate: any
    targetHosts?: string[]
    duration?: string
    loadProfile?: string
  }
  status: 'pending' | 'initializing' | 'running' | 'analyzing' | 'completed' | 'failed' | 'cancelled'
  owner?: string
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
  results?: ExperimentResults
}

export interface ExperimentResults {
  baselineMetrics: MetricsSummary
  candidateMetrics: MetricsSummary
  costReduction: number
  cardinalityReduction: number
  summary: string
}

export interface MetricsSummary {
  cardinality: number
  cpuUsage: number
  memoryUsage: number
  networkTraffic: number
}

interface ExperimentState {
  // State
  experiments: Experiment[]
  currentExperiment: Experiment | null
  loading: boolean
  error: string | null
  
  // Actions
  fetchExperiments: () => Promise<void>
  fetchExperiment: (id: string) => Promise<void>
  createExperiment: (data: CreateExperimentData) => Promise<Experiment>
  updateExperiment: (id: string, updates: Partial<Experiment>) => Promise<void>
  deleteExperiment: (id: string) => Promise<void>
  startExperiment: (id: string) => Promise<void>
  stopExperiment: (id: string) => Promise<void>
  promoteVariant: (id: string, variant: 'baseline' | 'candidate') => Promise<void>
  fetchExperimentAnalysis: (id: string) => Promise<any>
  fetchExperimentMetrics: (id: string) => Promise<any>
  
  // UI State
  setCurrentExperiment: (experiment: Experiment | null) => void
  clearError: () => void
}

export interface CreateExperimentData {
  name: string
  description: string
  baselinePipeline: string
  candidatePipeline: string
  targetNodes: string[]
  duration?: string
  successCriteria?: {
    minCardinalityReduction: number
    maxCostIncrease: number
    maxLatencyIncrease: number
  }
}

export const useExperimentStore = create<ExperimentState>()(
  devtools(
    (set, get) => ({
      // Initial state
      experiments: [],
      currentExperiment: null,
      loading: false,
      error: null,

      // Fetch all experiments
      fetchExperiments: async () => {
        set({ loading: true, error: null })
        try {
          const response = await apiService.getExperiments()
          set({ experiments: response.experiments, loading: false })
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Fetch single experiment
      fetchExperiment: async (id: string) => {
        set({ loading: true, error: null })
        try {
          const experiment = await apiService.getExperiment(id)
          set({ currentExperiment: experiment, loading: false })
          
          // Update in list as well
          set((state) => ({
            experiments: state.experiments.map((exp) =>
              exp.id === id ? experiment : exp
            ),
          }))
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Create new experiment
      createExperiment: async (data: CreateExperimentData) => {
        set({ loading: true, error: null })
        try {
          const experiment = await apiService.createExperiment({
            ...data,
            status: 'pending',
          })
          
          set((state) => ({
            experiments: [...state.experiments, experiment],
            currentExperiment: experiment,
            loading: false,
          }))
          
          return experiment
        } catch (error) {
          set({ error: error.message, loading: false })
          throw error
        }
      },

      // Update experiment
      updateExperiment: async (id: string, updates: Partial<Experiment>) => {
        set({ loading: true, error: null })
        try {
          const updated = await apiService.updateExperiment(id, updates)
          
          set((state) => ({
            experiments: state.experiments.map((exp) =>
              exp.id === id ? updated : exp
            ),
            currentExperiment:
              state.currentExperiment?.id === id
                ? updated
                : state.currentExperiment,
            loading: false,
          }))
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Delete experiment
      deleteExperiment: async (id: string) => {
        set({ loading: true, error: null })
        try {
          await apiService.deleteExperiment(id)
          
          set((state) => ({
            experiments: state.experiments.filter((exp) => exp.id !== id),
            currentExperiment:
              state.currentExperiment?.id === id
                ? null
                : state.currentExperiment,
            loading: false,
          }))
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Start experiment
      startExperiment: async (id: string) => {
        set({ loading: true, error: null })
        try {
          await apiService.startExperiment(id)
          
          // Update status
          set((state) => ({
            experiments: state.experiments.map((exp) =>
              exp.id === id
                ? { ...exp, status: 'initializing', startedAt: Date.now() }
                : exp
            ),
            loading: false,
          }))
          
          // Fetch updated experiment
          get().fetchExperiment(id)
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Stop experiment
      stopExperiment: async (id: string) => {
        set({ loading: true, error: null })
        try {
          await apiService.stopExperiment(id)
          
          // Update status
          set((state) => ({
            experiments: state.experiments.map((exp) =>
              exp.id === id
                ? { ...exp, status: 'cancelled', completedAt: Date.now() }
                : exp
            ),
            loading: false,
          }))
          
          // Fetch updated experiment
          get().fetchExperiment(id)
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // Promote variant
      promoteVariant: async (id: string, variant: 'baseline' | 'candidate') => {
        set({ loading: true, error: null })
        try {
          await apiService.promoteVariant(id, variant)
          
          // Update status
          set((state) => ({
            experiments: state.experiments.map((exp) =>
              exp.id === id
                ? { ...exp, status: 'completed', completedAt: Date.now() }
                : exp
            ),
            loading: false,
          }))
          
          // Fetch updated experiment
          get().fetchExperiment(id)
        } catch (error) {
          set({ error: error.message, loading: false })
        }
      },

      // UI state management
      setCurrentExperiment: (experiment: Experiment | null) => {
        set({ currentExperiment: experiment })
      },

      // Fetch experiment analysis
      fetchExperimentAnalysis: async (id: string) => {
        try {
          const analysis = await apiService.getExperimentAnalysis(id)
          return analysis
        } catch (error) {
          console.error('Failed to fetch analysis:', error)
          throw error
        }
      },

      // Fetch experiment metrics
      fetchExperimentMetrics: async (id: string) => {
        try {
          const metrics = await apiService.getExperimentMetrics(id)
          return metrics
        } catch (error) {
          console.error('Failed to fetch metrics:', error)
          throw error
        }
      },

      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'experiment-store',
    }
  )
)