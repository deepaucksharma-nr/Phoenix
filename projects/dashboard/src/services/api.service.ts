import axios, { AxiosInstance } from 'axios'
import { 
  Experiment, 
  ExperimentSpec, 
  ExperimentStatus,
  MetricsData,
  AnalysisResult,
  PipelineConfig,
} from '@/types'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

class ApiService {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })
  }

  // Experiments
  async getExperiments(params?: {
    owner?: string
    status?: string
    limit?: number
    offset?: number
  }) {
    const response = await this.client.get<{
      experiments: Experiment[]
      total: number
    }>('/experiments', { params })
    return response.data
  }

  async getExperiment(id: string) {
    const response = await this.client.get<Experiment>(`/experiments/${id}`)
    return response.data
  }

  async createExperiment(spec: ExperimentSpec) {
    const response = await this.client.post<Experiment>('/experiments', { spec })
    return response.data
  }

  async updateExperiment(id: string, updates: Partial<ExperimentSpec>) {
    const response = await this.client.patch<Experiment>(
      `/experiments/${id}`,
      updates
    )
    return response.data
  }

  async deleteExperiment(id: string) {
    await this.client.delete(`/experiments/${id}`)
  }

  async getExperimentStatus(id: string) {
    const response = await this.client.get<ExperimentStatus>(
      `/experiments/${id}/status`
    )
    return response.data
  }

  async getExperimentMetrics(
    id: string,
    params?: {
      metric?: string
      timeRange?: string
      variant?: string
    }
  ) {
    const response = await this.client.get<MetricsData>(
      `/experiments/${id}/metrics`,
      { params }
    )
    return response.data
  }

  async getExperimentAnalysis(id: string) {
    const response = await this.client.get<AnalysisResult>(
      `/experiments/${id}/analysis`
    )
    return response.data
  }

  async promoteVariant(experimentId: string, variant: 'baseline' | 'candidate') {
    const response = await this.client.post(
      `/experiments/${experimentId}/promote`,
      { variant }
    )
    return response.data
  }

  async startExperiment(id: string) {
    const response = await this.client.post(`/experiments/${id}/start`)
    return response.data
  }

  async stopExperiment(id: string) {
    const response = await this.client.post(`/experiments/${id}/stop`)
    return response.data
  }

  // Pipelines
  async validatePipeline(pipeline: PipelineConfig) {
    const response = await this.client.post<{
      valid: boolean
      errors: string[]
      warnings: string[]
    }>('/pipelines/validate', pipeline)
    return response.data
  }

  async previewPipeline(pipeline: PipelineConfig) {
    const response = await this.client.post<string>(
      '/pipelines/preview',
      pipeline,
      {
        headers: {
          Accept: 'application/x-yaml',
        },
      }
    )
    return response.data
  }

  async getPipelineTemplates() {
    const response = await this.client.get<{
      templates: Array<{
        id: string
        name: string
        description: string
        category: string
        config: PipelineConfig
      }>
    }>('/pipelines/templates')
    return response.data
  }

  async getPipelineDeployments() {
    const response = await this.client.get('/pipelines/deployments')
    return response.data
  }

  // Processors
  async getProcessorLibrary() {
    const response = await this.client.get('/processors')
    return response.data
  }

  // Health check
  async checkHealth() {
    const response = await this.client.get('/health')
    return response.data
  }
}

export const apiService = new ApiService()
export default apiService