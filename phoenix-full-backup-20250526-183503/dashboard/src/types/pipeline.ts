import { Node, Edge } from 'react-flow-renderer'

export interface ProcessorNodeData {
  label: string
  processorType: string
  category: 'filter' | 'transform' | 'aggregate' | 'system' | 'export'
  config: Record<string, any>
  isValid?: boolean
  errors?: string[]
  icon?: React.ReactNode
}

export interface PipelineConfig {
  nodes: Node<ProcessorNodeData>[]
  edges: Edge[]
  metadata?: {
    createdAt: string
    version: string
    description?: string
    author?: string
  }
}

export interface ValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export interface ProcessorTemplate {
  type: string
  category: ProcessorNodeData['category']
  label: string
  description: string
  icon?: React.ReactNode
  defaultConfig: Record<string, any>
  schema: ProcessorSchema
}

export interface ProcessorSchema {
  [key: string]: {
    type: 'string' | 'number' | 'boolean' | 'select' | 'multiselect' | 'string[]' | 'duration' | 'rules'
    label: string
    description?: string
    required?: boolean
    default?: any
    min?: number
    max?: number
    options?: string[]
    pattern?: string
  }
}

export interface GeneratedConfig {
  yaml: string
  receivers: Record<string, any>
  processors: Record<string, any>
  exporters: Record<string, any>
  service: {
    pipelines: {
      metrics: {
        receivers: string[]
        processors: string[]
        exporters: string[]
      }
    }
  }
}

export interface PipelineValidationError {
  nodeId?: string
  edgeId?: string
  type: 'error' | 'warning'
  message: string
  field?: string
}