import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'

// Types
interface PipelineDeployment {
  id: string
  name: string
  pipeline: string
  namespace: string
  status: 'active' | 'pending' | 'failed' | 'stopped'
  phase: string
  targetNodes: Record<string, string>
  instances: {
    desired: number
    ready: number
  }
  metrics: {
    cardinality: number
    throughput: string
    errorRate: number
    cpuUsage: number
    memoryUsage: number
  }
  createdAt: string
  updatedAt: string
}

interface PipelineTemplate {
  id: string
  name: string
  description: string
  category: string
  version: string
  author: string
  tags: string[]
  performance: {
    avgLatency: string
    cpuUsage: string
    memoryUsage: string
    cardinalityReduction: string
  }
  yaml: string
}

interface PipelineState {
  deployments: PipelineDeployment[]
  templates: PipelineTemplate[]
  loading: boolean
  error: string | null
  currentDeployment: PipelineDeployment | null
}

const initialState: PipelineState = {
  deployments: [],
  templates: [],
  loading: false,
  error: null,
  currentDeployment: null,
}

// Async thunks for API calls
export const fetchPipelineDeployments = createAsyncThunk(
  'pipelines/fetchDeployments',
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/v1/pipelines/deployments', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.deployments || []
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const fetchPipelineTemplates = createAsyncThunk(
  'pipelines/fetchTemplates',
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/v1/pipelines/templates', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.templates || []
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const deployPipeline = createAsyncThunk(
  'pipelines/deployPipeline',
  async (deploymentData: {
    name: string
    pipeline: string
    namespace: string
    targetNodes: Record<string, string>
    config?: any
  }, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/v1/pipelines/deployments', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(deploymentData),
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.deployment
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const fetchPipelineDeployment = createAsyncThunk(
  'pipelines/fetchDeployment',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await fetch(`/api/v1/pipelines/deployments/${id}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data.deployment
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const deletePipelineDeployment = createAsyncThunk(
  'pipelines/deleteDeployment',
  async (id: string, { rejectWithValue }) => {
    try {
      const response = await fetch(`/api/v1/pipelines/deployments/${id}`, {
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

const pipelineSlice = createSlice({
  name: 'pipelines',
  initialState,
  reducers: {
    setCurrentDeployment: (state, action: PayloadAction<PipelineDeployment | null>) => {
      state.currentDeployment = action.payload
    },
    clearError: (state) => {
      state.error = null
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch deployments
      .addCase(fetchPipelineDeployments.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchPipelineDeployments.fulfilled, (state, action) => {
        state.loading = false
        state.deployments = action.payload
      })
      .addCase(fetchPipelineDeployments.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Fetch templates
      .addCase(fetchPipelineTemplates.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchPipelineTemplates.fulfilled, (state, action) => {
        state.loading = false
        state.templates = action.payload
      })
      .addCase(fetchPipelineTemplates.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Deploy pipeline
      .addCase(deployPipeline.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(deployPipeline.fulfilled, (state, action) => {
        state.loading = false
        state.deployments.push(action.payload)
      })
      .addCase(deployPipeline.rejected, (state, action) => {
        state.loading = false
        state.error = action.payload as string
      })
      
      // Fetch single deployment
      .addCase(fetchPipelineDeployment.fulfilled, (state, action) => {
        state.currentDeployment = action.payload
        const index = state.deployments.findIndex(dep => dep.id === action.payload.id)
        if (index !== -1) {
          state.deployments[index] = action.payload
        }
      })
      
      // Delete deployment
      .addCase(deletePipelineDeployment.fulfilled, (state, action) => {
        state.deployments = state.deployments.filter(dep => dep.id !== action.payload)
      })
  },
})

export const { setCurrentDeployment, clearError } = pipelineSlice.actions
export default pipelineSlice.reducer