import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit'

// Types
interface PipelineDeployment {
  id: string
  name: string
  pipeline: string
  namespace: string
  status: 'active' | 'pending' | 'failed' | 'stopped'
  instances: {
    desired: number
    ready: number
    unavailable?: number
  }
  metrics: {
    cardinality: number
    throughput: string
    errorRate: number
    cpuUsage: number
    memoryUsage: number
  }
  configuration?: {
    aggregationWindow: string
    cardinalityLimit: number
    samplingRate: number
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
      // Try real API first, fall back to mock data
      try {
        const response = await fetch('/api/v1/pipelines/deployments', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
            'Content-Type': 'application/json',
          },
        })
        
        if (response.ok) {
          const data = await response.json()
          return data.deployments || []
        }
      } catch (apiError) {
        console.log('Pipeline API not available, using mock data for demonstration')
      }
      
      // Mock data for demonstration
      await new Promise(resolve => setTimeout(resolve, 600))
      
      const mockDeployments: PipelineDeployment[] = [
        {
          id: 'deploy-1',
          name: 'prod-metrics-optimizer-v2',
          pipeline: 'process-metrics-optimizer',
          namespace: 'phoenix-production',
          status: 'active',
          instances: { desired: 3, ready: 3, unavailable: 0 },
          metrics: {
            cardinality: 45000,
            throughput: '15.2k/sec',
            cpuUsage: 18,
            memoryUsage: 65,
            errorRate: 0.001,
          },
          configuration: {
            aggregationWindow: '30s',
            cardinalityLimit: 50000,
            samplingRate: 0.1,
          },
          createdAt: new Date('2024-05-20T10:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-26T14:30:00Z').toISOString(),
        },
        {
          id: 'deploy-2',
          name: 'staging-memory-sampler',
          pipeline: 'adaptive-memory-sampling',
          namespace: 'phoenix-staging',
          status: 'active',
          instances: { desired: 2, ready: 2, unavailable: 0 },
          metrics: {
            cardinality: 12500,
            throughput: '8.7k/sec',
            cpuUsage: 12,
            memoryUsage: 45,
            errorRate: 0.0005,
          },
          configuration: {
            aggregationWindow: '60s',
            cardinalityLimit: 15000,
            samplingRate: 0.05,
          },
          createdAt: new Date('2024-05-15T14:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-25T10:15:00Z').toISOString(),
        },
        {
          id: 'deploy-3',
          name: 'dev-container-dedup',
          pipeline: 'container-deduplication',
          namespace: 'phoenix-development',
          status: 'pending',
          instances: { desired: 1, ready: 0, unavailable: 1 },
          metrics: {
            cardinality: 0,
            throughput: '0/sec',
            cpuUsage: 0,
            memoryUsage: 0,
            errorRate: 0,
          },
          configuration: {
            aggregationWindow: '120s',
            cardinalityLimit: 5000,
            samplingRate: 0.01,
          },
          createdAt: new Date('2024-05-25T09:00:00Z').toISOString(),
          updatedAt: new Date('2024-05-25T09:00:00Z').toISOString(),
        },
      ]

      return mockDeployments
    } catch (error: any) {
      return rejectWithValue(error.message)
    }
  }
)

export const fetchPipelineTemplates = createAsyncThunk(
  'pipelines/fetchTemplates',
  async (_, { rejectWithValue }) => {
    try {
      // Try real API first, fall back to mock data
      try {
        const response = await fetch('/api/v1/pipelines/templates', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
            'Content-Type': 'application/json',
          },
        })
        
        if (response.ok) {
          const data = await response.json()
          return data.templates || []
        }
      } catch (apiError) {
        console.log('Templates API not available, using mock data for demonstration')
      }
      
      // Mock templates data
      await new Promise(resolve => setTimeout(resolve, 400))
      
      const mockTemplates: PipelineTemplate[] = [
        {
          id: '1',
          name: 'Process Metrics Optimizer',
          description: 'Optimizes process-level metrics by aggregating similar processes and reducing cardinality through intelligent grouping.',
          category: 'optimization',
          version: '1.2.0',
          author: 'Phoenix Team',
          tags: ['process', 'cardinality', 'aggregation', 'production-ready'],
          performance: {
            avgLatency: '2.3ms',
            cpuUsage: '15%',
            memoryUsage: '128MB',
            cardinalityReduction: '85%',
          },
          yaml: `processors:
  attributes:
    actions:
      - key: process.executable.name
        action: hash
      - key: process.pid
        action: delete
  resource:
    attributes:
      - key: service.name
        from_attribute: process.executable.name
        action: insert
  batch:
    timeout: 200ms
    send_batch_size: 8192
  memory_limiter:
    check_interval: 1s
    limit_mib: 512`,
        },
        {
          id: '2',
          name: 'Tail Sampling Pipeline',
          description: 'Implements intelligent tail sampling to capture important traces while reducing overall volume.',
          category: 'sampling',
          version: '2.0.1',
          author: 'Phoenix Team',
          tags: ['traces', 'sampling', 'performance', 'errors'],
          performance: {
            avgLatency: '5.1ms',
            cpuUsage: '25%',
            memoryUsage: '256MB',
            cardinalityReduction: '70%',
          },
          yaml: `processors:
  tail_sampling:
    decision_wait: 10s
    num_traces: 100000
    policies:
      - name: errors-policy
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: slow-traces-policy
        type: latency
        latency: {threshold_ms: 1000}`,
        },
        {
          id: '3',
          name: 'Metrics Aggregator',
          description: 'Aggregates metrics at collection time to reduce storage requirements while maintaining query performance.',
          category: 'aggregation',
          version: '1.5.3',
          author: 'Community',
          tags: ['metrics', 'aggregation', 'storage', 'cost-optimization'],
          performance: {
            avgLatency: '3.7ms',
            cpuUsage: '20%',
            memoryUsage: '192MB',
            cardinalityReduction: '75%',
          },
          yaml: `processors:
  metricstransform:
    transforms:
      - include: .*
        match_type: regexp
        action: update
        operations:
          - action: aggregate_labels
            label_set: [service.name, service.namespace]
            aggregation_type: sum`,
        },
      ]

      return mockTemplates
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