import { create } from 'zustand'
import { Node, Edge } from 'reactflow'
import { apiService } from '../services/api.service'
import { ProcessorNodeData } from '../components/PipelineBuilder/ProcessorNode'

export interface PipelineTemplate {
  id: string
  name: string
  description: string
  category: string
  nodes: Node<ProcessorNodeData>[]
  edges: Edge[]
  config?: Record<string, any>
}

interface PipelineState {
  // State
  templates: PipelineTemplate[]
  currentPipeline: {
    nodes: Node<ProcessorNodeData>[]
    edges: Edge[]
    name?: string
    description?: string
  }
  loading: boolean
  error: string | null
  validationErrors: string[]
  
  // Actions
  fetchTemplates: () => Promise<void>
  loadTemplate: (templateId: string) => void
  savePipeline: (name: string, description: string) => Promise<void>
  validatePipeline: () => Promise<boolean>
  generateYAML: () => string
  importYAML: (yaml: string) => Promise<void>
  
  // Pipeline editing
  setNodes: (nodes: Node<ProcessorNodeData>[]) => void
  setEdges: (edges: Edge[]) => void
  clearPipeline: () => void
}

export const usePipelineStore = create<PipelineState>((set, get) => ({
  // Initial state
  templates: [],
  currentPipeline: {
    nodes: [],
    edges: [],
  },
  loading: false,
  error: null,
  validationErrors: [],

  // Fetch pipeline templates
  fetchTemplates: async () => {
    set({ loading: true, error: null })
    try {
      const response = await apiService.getPipelineTemplates()
      set({ templates: response.templates, loading: false })
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  // Load a template
  loadTemplate: (templateId: string) => {
    const template = get().templates.find((t) => t.id === templateId)
    if (template) {
      set({
        currentPipeline: {
          nodes: [...template.nodes],
          edges: [...template.edges],
          name: template.name,
          description: template.description,
        },
      })
    }
  },

  // Save current pipeline
  savePipeline: async (name: string, description: string) => {
    const { currentPipeline } = get()
    set({ loading: true, error: null })
    
    try {
      // Validate first
      const isValid = await get().validatePipeline()
      if (!isValid) {
        throw new Error('Pipeline validation failed')
      }

      // Generate YAML
      const yaml = get().generateYAML()
      
      // Save pipeline
      await apiService.savePipeline({
        name,
        description,
        nodes: currentPipeline.nodes,
        edges: currentPipeline.edges,
        yaml,
      })
      
      set({ loading: false })
    } catch (error: any) {
      set({ error: error.message, loading: false })
      throw error
    }
  },

  // Validate pipeline
  validatePipeline: async () => {
    const { nodes, edges } = get().currentPipeline
    const errors: string[] = []

    // Check if pipeline is empty
    if (nodes.length === 0) {
      errors.push('Pipeline must have at least one processor')
    }

    // Get ordered nodes
    const orderedNodes = getOrderedNodes(nodes, edges)

    // Check for memory_limiter as first processor
    if (orderedNodes.length > 0) {
      const firstNode = orderedNodes[0]
      if (!firstNode.data.processorType.includes('memory_limiter')) {
        errors.push('Pipeline should start with a memory_limiter processor for safety')
      }
    }

    // Check for batch processor at the end
    if (orderedNodes.length > 0) {
      const lastNode = orderedNodes[orderedNodes.length - 1]
      if (!lastNode.data.processorType.includes('batch')) {
        errors.push('Pipeline should end with a batch processor for efficiency')
      }
    }

    // Check for disconnected nodes
    const connectedNodeIds = new Set<string>()
    edges.forEach((edge) => {
      connectedNodeIds.add(edge.source)
      connectedNodeIds.add(edge.target)
    })
    
    nodes.forEach((node) => {
      if (!connectedNodeIds.has(node.id) && nodes.length > 1) {
        errors.push(`Processor "${node.data.label}" is not connected to the pipeline`)
      }
    })

    // Check for cycles
    if (hasCycle(nodes, edges)) {
      errors.push('Pipeline contains a cycle - processors must form a directed acyclic graph')
    }

    set({ validationErrors: errors })
    return errors.length === 0
  },

  // Generate YAML configuration
  generateYAML: () => {
    const { nodes, edges } = get().currentPipeline
    const orderedNodes = getOrderedNodes(nodes, edges)
    
    // Build configuration
    const config = {
      receivers: {
        hostmetrics: {
          collection_interval: '30s',
          scrapers: {
            process: {
              include: {
                match_type: 'regexp',
                names: ['.*'],
              },
            },
          },
        },
      },
      processors: {} as Record<string, any>,
      exporters: {
        otlphttp: {
          endpoint: '${NEW_RELIC_OTLP_ENDPOINT}',
          headers: {
            'api-key': '${NEW_RELIC_API_KEY}',
          },
        },
        prometheus: {
          endpoint: '0.0.0.0:8888',
        },
      },
      service: {
        pipelines: {
          metrics: {
            receivers: ['hostmetrics'],
            processors: [] as string[],
            exporters: ['otlphttp', 'prometheus'],
          },
        },
      },
    }

    // Add processors
    orderedNodes.forEach((node, index) => {
      const processorName = `${node.data.processorType.replace('/', '_')}_${index}`
      config.processors[processorName] = node.data.config
      config.service.pipelines.metrics.processors.push(processorName)
    })

    // Convert to YAML string (simplified)
    return JSON.stringify(config, null, 2)
  },

  // Import from YAML
  importYAML: async (yaml: string) => {
    try {
      // Parse YAML (would use proper YAML parser)
      const config = JSON.parse(yaml)
      
      // Convert to nodes and edges
      const nodes: Node<ProcessorNodeData>[] = []
      const edges: Edge[] = []
      
      // Extract processors
      if (config.processors) {
        let y = 50
        let prevNodeId: string | null = null
        
        Object.entries(config.processors).forEach(([name, processorConfig], index) => {
          const nodeId = `imported_${index}`
          const processorType = name.replace(/_\d+$/, '').replace('_', '/')
          
          nodes.push({
            id: nodeId,
            type: 'processor',
            position: { x: 250, y },
            data: {
              label: name,
              processorType,
              category: getProcessorCategory(processorType),
              config: processorConfig as Record<string, any>,
            },
          })
          
          // Create edge from previous node
          if (prevNodeId) {
            edges.push({
              id: `e${prevNodeId}-${nodeId}`,
              source: prevNodeId,
              target: nodeId,
              type: 'smoothstep',
            })
          }
          
          prevNodeId = nodeId
          y += 100
        })
      }
      
      set({
        currentPipeline: {
          nodes,
          edges,
        },
      })
    } catch (error: any) {
      set({ error: `Failed to import YAML: ${error.message}` })
      throw error
    }
  },

  // Pipeline editing actions
  setNodes: (nodes: Node<ProcessorNodeData>[]) => {
    set((state) => ({
      currentPipeline: {
        ...state.currentPipeline,
        nodes,
      },
    }))
  },

  setEdges: (edges: Edge[]) => {
    set((state) => ({
      currentPipeline: {
        ...state.currentPipeline,
        edges,
      },
    }))
  },

  clearPipeline: () => {
    set({
      currentPipeline: {
        nodes: [],
        edges: [],
      },
      validationErrors: [],
    })
  },
}))

// Helper functions
function getOrderedNodes(nodes: Node<ProcessorNodeData>[], edges: Edge[]): Node<ProcessorNodeData>[] {
  const nodeMap = new Map(nodes.map((n) => [n.id, n]))
  const inDegree = new Map<string, number>()
  const adjList = new Map<string, string[]>()

  // Initialize
  nodes.forEach((node) => {
    inDegree.set(node.id, 0)
    adjList.set(node.id, [])
  })

  // Build graph
  edges.forEach((edge) => {
    inDegree.set(edge.target, (inDegree.get(edge.target) || 0) + 1)
    const neighbors = adjList.get(edge.source) || []
    neighbors.push(edge.target)
    adjList.set(edge.source, neighbors)
  })

  // Topological sort
  const queue: string[] = []
  inDegree.forEach((degree, nodeId) => {
    if (degree === 0) queue.push(nodeId)
  })

  const ordered: Node<ProcessorNodeData>[] = []
  while (queue.length > 0) {
    const nodeId = queue.shift()!
    const node = nodeMap.get(nodeId)
    if (node) ordered.push(node)

    const neighbors = adjList.get(nodeId) || []
    neighbors.forEach((neighbor) => {
      const newDegree = (inDegree.get(neighbor) || 1) - 1
      inDegree.set(neighbor, newDegree)
      if (newDegree === 0) queue.push(neighbor)
    })
  }

  return ordered
}

function hasCycle(nodes: Node[], edges: Edge[]): boolean {
  const adjList = new Map<string, string[]>()
  nodes.forEach((node) => adjList.set(node.id, []))
  edges.forEach((edge) => {
    const neighbors = adjList.get(edge.source) || []
    neighbors.push(edge.target)
    adjList.set(edge.source, neighbors)
  })

  const visited = new Set<string>()
  const recursionStack = new Set<string>()

  function dfs(nodeId: string): boolean {
    visited.add(nodeId)
    recursionStack.add(nodeId)

    const neighbors = adjList.get(nodeId) || []
    for (const neighbor of neighbors) {
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

function getProcessorCategory(processorType: string): 'filter' | 'transform' | 'aggregate' | 'system' | 'export' {
  if (processorType.includes('filter')) return 'filter'
  if (processorType.includes('transform')) return 'transform'
  if (processorType.includes('group') || processorType.includes('aggregate')) return 'aggregate'
  if (processorType.includes('batch') || processorType.includes('memory')) return 'system'
  return 'export'
}