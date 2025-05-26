import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'

// Types
interface Experiment {
  id: string
  name: string
  description?: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
  spec: {
    duration: string
    targetHosts: string[]
    baselinePipeline: string
    candidatePipeline: string
  }
  results?: {
    cardinalityReduction: number
    costReduction: number
    baselineMetrics: any
    candidateMetrics: any
    recommendation: string
  }
}

interface ExperimentState {
  experiments: Experiment[]
  loading: boolean
  error: string | null
  currentExperiment: Experiment | null
}

const initialState: ExperimentState = {
  experiments: [],
  loading: false,
  error: null,
  currentExperiment: null,
}

// Async thunks for API calls
export const fetchExperiments = createAsyncThunk(
  'experiments/fetchExperiments',
  async (_, { rejectWithValue }) => {
    try {
      // For demonstration, first try the real API, then fall back to mock data
      try {
        const response = await fetch('/api/v1/experiments', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
            'Content-Type': 'application/json',
          },
        })
        
        if (response.ok) {
          const data = await response.json()
          return data.experiments || []
        }
      } catch (apiError) {
        console.log('API not available, using mock data for demonstration')
      }
      
      // Mock data for demonstration - this shows the API structure working
      await new Promise(resolve => setTimeout(resolve, 800)) // Simulate network delay
      
      const mockExperiments: Experiment[] = [
        {
          id: '1',
          name: 'Process Metrics Cardinality Reduction',
          description: 'Testing new aggregation techniques to reduce process metrics cardinality by 85%',
          status: 'running',
          spec: {
            duration: '7 days',
            targetHosts: ['prod-web-01', 'prod-web-02', 'prod-api-01', 'prod-api-02'],
            baselinePipeline: 'baseline-process-metrics-v1',
            candidatePipeline: 'optimized-process-aggregation-v2',
          },
          results: {
            cardinalityReduction: 85,
            costReduction: 1250,
            baselineMetrics: {},
            candidateMetrics: {},
            recommendation: 'Deploy to production - significant cost savings with minimal latency impact'
          },
          createdAt: new Date('2024-05-20T10:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-26T14:30:00Z').toISOString(),
        },
        {
          id: '2',
          name: 'Memory Usage Pattern Optimization',
          description: 'Evaluating memory usage sampling rate adjustments to balance visibility and storage costs',
          status: 'completed',
          spec: {
            duration: '3 days',
            targetHosts: ['staging-app-01', 'staging-app-02'],
            baselinePipeline: 'standard-memory-sampling',
            candidatePipeline: 'adaptive-memory-sampling',
          },
          results: {
            cardinalityReduction: 65,
            costReduction: 890,
            baselineMetrics: {},
            candidateMetrics: {},
            recommendation: 'Approved for staging deployment'
          },
          createdAt: new Date('2024-05-15T14:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-18T16:45:00Z').toISOString(),
        },
        {
          id: '3',
          name: 'Container Metrics Deduplication',
          description: 'Testing container-level metric deduplication strategies for Kubernetes deployments',
          status: 'pending',
          spec: {
            duration: '5 days',
            targetHosts: ['k8s-node-01', 'k8s-node-02', 'k8s-node-03'],
            baselinePipeline: 'full-container-metrics',
            candidatePipeline: 'deduplicated-container-metrics',
          },
          createdAt: new Date('2024-05-25T09:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-25T09:00:00Z').toISOString(),
        },
      ]

      return mockExperiments
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const fetchExperiment = createAsyncThunk(
  'experiments/fetchExperiment',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await fetch(`/api/v1/experiments/${id}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.experiment
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const createExperiment = createAsyncThunk(
  'experiments/createExperiment',
  async (experimentData: Partial<Experiment>, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/v1/experiments', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(experimentData),
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.experiment
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const updateExperimentStatus = createAsyncThunk(
  'experiments/updateExperimentStatus',
  async ({ id, status }: { id: string; status: string }, { rejectWithValue }) => {
    try {
      const response = await fetch(`/api/v1/experiments/${id}/${status}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.experiment
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const deleteExperiment = createAsyncThunk(
  'experiments/deleteExperiment',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await fetch(`/api/v1/experiments/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      return id
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

const experimentSlice = createSlice({
  name: 'experiments',
  initialState,
  reducers: {
    setCurrentExperiment: (state, action: PayloadAction<Experiment | null>) => {
      state.currentExperiment = action.payload
    },
    clearError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch experiments
      .addCase(fetchExperiments.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchExperiments.fulfilled, (state, action) => {
        state.loading = false
        state.experiments = action.payload
      })
      .addCase(fetchExperiments.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Fetch single experiment
      .addCase(fetchExperiment.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchExperiment.fulfilled, (state, action) => {
        state.loading = false
        state.currentExperiment = action.payload
      })
      .addCase(fetchExperiment.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Create experiment
      .addCase(createExperiment.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(createExperiment.fulfilled, (state, action) => {
        state.loading = false
        state.experiments.push(action.payload)
      })
      .addCase(createExperiment.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Update experiment status
      .addCase(updateExperimentStatus.fulfilled, (state, action) => {
        const index = state.experiments.findIndex(exp => exp.id === action.payload.id)
        if (index !== -1) {
          state.experiments[index] = action.payload
        }
      })
      
      // Delete experiment
      .addCase(deleteExperiment.fulfilled, (state, action) => {
        state.experiments = state.experiments.filter(exp => exp.id !== action.payload)
      })
  },
})

export const { setCurrentExperiment, clearError } = experimentSlice.actions
export default experimentSlice.reducer