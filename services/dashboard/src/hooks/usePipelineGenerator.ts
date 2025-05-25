import { useCallback } from 'react'
import { Node, Edge } from 'react-flow-renderer'
import { ProcessorNodeData, ValidationResult, GeneratedConfig, PipelineValidationError } from '@/types/pipeline'
import yaml from 'js-yaml'

export const usePipelineGenerator = () => {
  const generateConfig = useCallback((nodes: Node<ProcessorNodeData>[], edges: Edge[]): string => {
    // Build the configuration object
    const config = {
      receivers: {
        hostmetrics: {
          collection_interval: '30s',
          scrapers: {
            process: {
              include: {
                match_type: 'regexp',
                names: ['.*']
              }
            }
          }
        }
      },
      processors: {} as Record<string, any>,
      exporters: {
        otlphttp: {
          endpoint: '${NEW_RELIC_OTLP_ENDPOINT}',
          headers: {
            'api-key': '${NEW_RELIC_API_KEY}'
          }
        },
        prometheus: {
          endpoint: '0.0.0.0:8888'
        }
      },
      service: {
        pipelines: {
          metrics: {
            receivers: ['hostmetrics'],
            processors: [] as string[],
            exporters: ['otlphttp', 'prometheus']
          }
        }
      }
    }

    // Build processing pipeline based on nodes and edges
    const sortedNodes = topologicalSort(nodes, edges)
    
    for (const node of sortedNodes) {
      const processorName = getProcessorName(node.data.processorType, node.id)
      config.processors[processorName] = buildProcessorConfig(node.data)
      config.service.pipelines.metrics.processors.push(processorName)
    }

    // Always add batch processor at the end if not already present
    if (!config.service.pipelines.metrics.processors.some(p => p.includes('batch'))) {
      config.processors.batch = {
        timeout: '10s',
        send_batch_size: 1000
      }
      config.service.pipelines.metrics.processors.push('batch')
    }

    // Convert to YAML
    return yaml.dump(config, {
      indent: 2,
      lineWidth: 120,
      noRefs: true
    })
  }, [])

  const validatePipeline = useCallback((config: { nodes: Node<ProcessorNodeData>[], edges: Edge[] }): ValidationResult => {
    const errors: string[] = []
    const warnings: string[] = []

    // Check for empty pipeline
    if (config.nodes.length === 0) {
      warnings.push('Pipeline is empty. Add some processors to optimize your metrics collection.')
      return { valid: true, errors, warnings }
    }

    // Validate node configurations
    for (const node of config.nodes) {
      const nodeErrors = validateNodeConfig(node)
      errors.push(...nodeErrors.map(err => `${node.data.label}: ${err}`))
    }

    // Check for disconnected nodes
    const connectedNodes = new Set<string>()
    for (const edge of config.edges) {
      connectedNodes.add(edge.source)
      connectedNodes.add(edge.target)
    }

    for (const node of config.nodes) {
      if (config.nodes.length > 1 && !connectedNodes.has(node.id)) {
        warnings.push(`${node.data.label} is not connected to the pipeline`)
      }
    }

    // Check for cycles
    if (hasCycles(config.nodes, config.edges)) {
      errors.push('Pipeline contains cycles. Remove cyclic connections.')
    }

    // Validate processor order
    const orderWarnings = validateProcessorOrder(config.nodes, config.edges)
    warnings.push(...orderWarnings)

    return {
      valid: errors.length === 0,
      errors,
      warnings
    }
  }, [])

  return {
    generateConfig,
    validatePipeline
  }
}

// Helper functions
function topologicalSort(nodes: Node<ProcessorNodeData>[], edges: Edge[]): Node<ProcessorNodeData>[] {
  // Simple topological sort based on edges
  const inDegree = new Map<string, number>()
  const graph = new Map<string, string[]>()
  const nodeMap = new Map<string, Node<ProcessorNodeData>>()

  // Initialize
  for (const node of nodes) {
    inDegree.set(node.id, 0)
    graph.set(node.id, [])
    nodeMap.set(node.id, node)
  }

  // Build graph
  for (const edge of edges) {
    graph.get(edge.source)?.push(edge.target)
    inDegree.set(edge.target, (inDegree.get(edge.target) || 0) + 1)
  }

  // Kahn's algorithm
  const queue: string[] = []
  const result: Node<ProcessorNodeData>[] = []

  // Find nodes with no incoming edges
  for (const [nodeId, degree] of inDegree.entries()) {
    if (degree === 0) {
      queue.push(nodeId)
    }
  }

  while (queue.length > 0) {
    const nodeId = queue.shift()!
    const node = nodeMap.get(nodeId)!
    result.push(node)

    for (const neighbor of graph.get(nodeId) || []) {
      inDegree.set(neighbor, inDegree.get(neighbor)! - 1)
      if (inDegree.get(neighbor) === 0) {
        queue.push(neighbor)
      }
    }
  }

  // If not all nodes are included, return original order (handles cycles)
  return result.length === nodes.length ? result : nodes
}

function getProcessorName(processorType: string, nodeId: string): string {
  // Convert processor type to OpenTelemetry processor name
  const typeMap: Record<string, string> = {
    'filter/priority': 'filter',
    'filter/resource': 'filter',
    'filter/topk': 'filter',
    'transform/classify': 'transform',
    'groupbyattrs': 'groupbyattrs',
    'memory_limiter': 'memory_limiter',
    'batch': 'batch'
  }

  const baseName = typeMap[processorType] || processorType.replace('/', '_')
  return `${baseName}_${nodeId.slice(-8)}` // Use last 8 chars of node ID for uniqueness
}

function buildProcessorConfig(nodeData: ProcessorNodeData): Record<string, any> {
  const { processorType, config } = nodeData

  // Convert UI config to OpenTelemetry processor config
  switch (processorType) {
    case 'filter/priority':
      return {
        metrics: {
          include: {
            match_type: 'regexp',
            metric_names: [`process_.*_priority_${config.minPriority || 'high'}`]
          }
        }
      }

    case 'filter/resource':
      const conditions = []
      if (config.minCpu) {
        conditions.push(`resource["process.cpu.utilization"] >= ${config.minCpu}`)
      }
      if (config.minMemory) {
        conditions.push(`resource["process.memory.usage"] >= ${config.minMemory * 1024 * 1024}`)
      }
      
      return {
        metrics: {
          include: {
            match_type: 'expr',
            expressions: conditions.length ? [conditions.join(config.operator === 'AND' ? ' and ' : ' or ')] : []
          }
        }
      }

    case 'filter/topk':
      return {
        metrics: {
          include: {
            match_type: 'regexp',
            metric_names: ['process_.*']
          }
        }
        // Note: topk functionality would need custom processor implementation
      }

    case 'transform/classify':
      return {
        metric_statements: [
          {
            context: 'metric',
            statements: [
              `set(attributes["process.priority"], "medium") where name matches "process_.*"`
            ]
          }
        ]
      }

    case 'groupbyattrs':
      return {
        keys: config.keys || ['process.name', 'host.name']
      }

    case 'memory_limiter':
      return {
        check_interval: config.checkInterval || '1s',
        limit_mib: parseInt(config.limit?.replace(/[^\d]/g, '') || '512')
      }

    case 'batch':
      return {
        timeout: config.timeout || '10s',
        send_batch_size: config.sendBatchSize || 1000
      }

    default:
      return config
  }
}

function validateNodeConfig(node: Node<ProcessorNodeData>): string[] {
  const errors: string[] = []
  const { processorType, config } = node.data

  // Validate based on processor type
  switch (processorType) {
    case 'filter/topk':
      if (!config.k || config.k < 1) {
        errors.push('Top K value must be at least 1')
      }
      break

    case 'memory_limiter':
      if (!config.limit || !/^\d+[KMG]i?B$/.test(config.limit)) {
        errors.push('Memory limit must be in format like "512MiB" or "1GiB"')
      }
      break

    case 'groupbyattrs':
      if (!config.keys || !Array.isArray(config.keys) || config.keys.length === 0) {
        errors.push('At least one grouping key is required')
      }
      break
  }

  return errors
}

function hasCycles(nodes: Node<ProcessorNodeData>[], edges: Edge[]): boolean {
  const visited = new Set<string>()
  const recursionStack = new Set<string>()
  const graph = new Map<string, string[]>()

  // Build adjacency list
  for (const node of nodes) {
    graph.set(node.id, [])
  }
  for (const edge of edges) {
    graph.get(edge.source)?.push(edge.target)
  }

  function dfs(nodeId: string): boolean {
    visited.add(nodeId)
    recursionStack.add(nodeId)

    for (const neighbor of graph.get(nodeId) || []) {
      if (!visited.has(neighbor)) {
        if (dfs(neighbor)) return true
      } else if (recursionStack.has(neighbor)) {
        return true
      }
    }

    recursionStack.delete(nodeId)
    return false
  }

  for (const node of nodes) {
    if (!visited.has(node.id)) {
      if (dfs(node.id)) return true
    }
  }

  return false
}

function validateProcessorOrder(nodes: Node<ProcessorNodeData>[], edges: Edge[]): string[] {
  const warnings: string[] = []
  
  // Check if memory_limiter is first
  const firstNodes = nodes.filter(node => 
    !edges.some(edge => edge.target === node.id)
  )
  
  const hasMemoryLimiterFirst = firstNodes.some(node => 
    node.data.processorType === 'memory_limiter'
  )
  
  if (nodes.some(node => node.data.processorType === 'memory_limiter') && !hasMemoryLimiterFirst) {
    warnings.push('memory_limiter processor should be placed first in the pipeline')
  }

  // Check if batch is last
  const lastNodes = nodes.filter(node => 
    !edges.some(edge => edge.source === node.id)
  )
  
  const hasBatchLast = lastNodes.some(node => 
    node.data.processorType === 'batch'
  )
  
  if (nodes.some(node => node.data.processorType === 'batch') && !hasBatchLast) {
    warnings.push('batch processor should be placed last in the pipeline')
  }

  return warnings
}