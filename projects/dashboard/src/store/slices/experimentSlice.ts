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
      const response = await fetch('/api/v1/experiments', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.experiments || []
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