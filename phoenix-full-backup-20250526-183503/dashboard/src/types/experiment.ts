export interface Experiment {
  id: string
  name: string
  description?: string
  spec?: ExperimentSpec
  status: ExperimentStatus
  owner?: string
  createdAt: string
  updatedAt: string
  startedAt?: string
  completedAt?: string
  results?: ExperimentResults
}

export type ExperimentStatus = 
  | 'pending' 
  | 'initializing' 
  | 'running' 
  | 'analyzing' 
  | 'completed' 
  | 'failed' 
  | 'cancelled'

export interface ExperimentSpec {
  baseline: PipelineConfig
  candidate: PipelineConfig
  targetHosts?: string[]
  duration?: string
  loadProfile?: 'realistic' | 'high-cardinality' | 'high-churn'
  successCriteria?: SuccessCriteria
}

export interface SuccessCriteria {
  minCardinalityReduction: number
  maxCostIncrease: number
  maxLatencyIncrease: number
  minCriticalProcessRetention: number
}

export interface PipelineConfig {
  name?: string
  receivers?: string[]
  processors?: ProcessorConfig[]
  exporters?: string[]
}

export interface ProcessorConfig {
  type: string
  config?: Record<string, any>
}

export interface ExperimentResults {
  baselineMetrics: MetricsSummary
  candidateMetrics: MetricsSummary
  costReduction: number
  cardinalityReduction: number
  summary: string
  recommendation?: 'baseline' | 'candidate'
}

export interface MetricsSummary {
  cardinality: number
  cpuUsage: number
  memoryUsage: number
  networkTraffic: number
  dataPointsPerSecond?: number
  uniqueProcesses?: number
}

export interface MetricsData {
  timestamp: number[]
  values: number[]
  metric: string
  variant: 'baseline' | 'candidate'
}

export interface AnalysisResult {
  experimentId: string
  status: 'pending' | 'in-progress' | 'completed' | 'failed'
  startTime: string
  endTime?: string
  comparison: {
    cardinalityReduction: number
    costSavings: number
    criticalProcessRetention: number
    performanceImpact: {
      cpuOverhead: number
      memoryOverhead: number
      latencyIncrease: number
    }
  }
  recommendation: {
    variant: 'baseline' | 'candidate'
    confidence: number
    reasons: string[]
  }
}

export interface CreateExperimentData {
  name: string
  description?: string
  spec: ExperimentSpec
}