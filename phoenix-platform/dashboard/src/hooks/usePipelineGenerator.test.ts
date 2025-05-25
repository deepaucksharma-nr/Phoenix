/**
 * @jest-environment jsdom
 */

import { renderHook } from '@testing-library/react'
import { usePipelineGenerator } from './usePipelineGenerator'
import { Node, Edge } from 'react-flow-renderer'
import { ProcessorNodeData } from '../types/pipeline'

// Mock yaml library
jest.mock('js-yaml', () => ({
  dump: jest.fn((config) => `# Generated YAML\n${JSON.stringify(config, null, 2)}`),
}))

describe('usePipelineGenerator', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('generateConfig', () => {
    it('should generate basic configuration with no nodes', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      const nodes: Node<ProcessorNodeData>[] = []
      const edges: Edge[] = []

      const config = result.current.generateConfig(nodes, edges)

      expect(config).toContain('receivers:')
      expect(config).toContain('hostmetrics:')
      expect(config).toContain('exporters:')
      expect(config).toContain('otlphttp:')
      expect(config).toContain('prometheus:')
      expect(config).toContain('service:')
      expect(config).toContain('batch:') // Should always include batch processor
    })

    it('should generate configuration with filter nodes', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'filter-1',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Priority Filter',
            processorType: 'filter/priority',
            category: 'filter',
            config: {
              minPriority: 'high',
              excludePatterns: ['test.*']
            }
          }
        }
      ]
      const edges: Edge[] = []

      const config = result.current.generateConfig(nodes, edges)

      expect(config).toContain('filter_')
      expect(config).toContain('batch:')
    })

    it('should generate configuration with multiple processor types', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'memory-1',
          type: 'processor', 
          position: { x: 0, y: 0 },
          data: {
            label: 'Memory Limiter',
            processorType: 'memory_limiter',
            category: 'system',
            config: {
              limit: '512MiB',
              checkInterval: '1s'
            }
          }
        },
        {
          id: 'filter-1',
          type: 'processor',
          position: { x: 100, y: 0 },
          data: {
            label: 'Resource Filter',
            processorType: 'filter/resource',
            category: 'filter',
            config: {
              minCpu: 5.0,
              minMemory: 50
            }
          }
        },
        {
          id: 'group-1',
          type: 'processor',
          position: { x: 200, y: 0 },
          data: {
            label: 'Group By Attrs',
            processorType: 'groupbyattrs',
            category: 'aggregate',
            config: {
              keys: ['process.name', 'host.name']
            }
          }
        }
      ]
      
      const edges: Edge[] = [
        { id: 'e1', source: 'memory-1', target: 'filter-1' },
        { id: 'e2', source: 'filter-1', target: 'group-1' }
      ]

      const config = result.current.generateConfig(nodes, edges)

      expect(config).toContain('memory_limiter_')
      expect(config).toContain('filter_')
      expect(config).toContain('groupbyattrs_')
      expect(config).toContain('batch:')
    })

    it('should handle topological sorting of connected nodes', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'node-3',
          type: 'processor',
          position: { x: 200, y: 0 },
          data: {
            label: 'Batch',
            processorType: 'batch',
            category: 'system',
            config: { timeout: '10s' }
          }
        },
        {
          id: 'node-1',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Memory Limiter',
            processorType: 'memory_limiter',
            category: 'system',
            config: { limit: '256MiB' }
          }
        },
        {
          id: 'node-2',
          type: 'processor',
          position: { x: 100, y: 0 },
          data: {
            label: 'Filter',
            processorType: 'filter/priority',
            category: 'filter',
            config: { minPriority: 'medium' }
          }
        }
      ]
      
      const edges: Edge[] = [
        { id: 'e1', source: 'node-1', target: 'node-2' },
        { id: 'e2', source: 'node-2', target: 'node-3' }
      ]

      const config = result.current.generateConfig(nodes, edges)
      
      // The processors should be ordered correctly in the pipeline
      expect(config).toContain('processors:')
      
      // Should include all processor types
      expect(config).toContain('memory_limiter_')
      expect(config).toContain('filter_')
      expect(config).toContain('batch_')
    })
  })

  describe('validatePipeline', () => {
    it('should validate empty pipeline', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const validation = result.current.validatePipeline({
        nodes: [],
        edges: []
      })

      expect(validation.valid).toBe(true)
      expect(validation.errors).toHaveLength(0)
      expect(validation.warnings).toContain('Pipeline is empty')
    })

    it('should validate single node pipeline', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'single-node',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Memory Limiter',
            processorType: 'memory_limiter',
            category: 'system',
            config: {
              limit: '512MiB'
            }
          }
        }
      ]

      const validation = result.current.validatePipeline({
        nodes,
        edges: []
      })

      expect(validation.valid).toBe(true)
      expect(validation.errors).toHaveLength(0)
    })

    it('should detect disconnected nodes', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'connected-1',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Node 1',
            processorType: 'filter/priority',
            category: 'filter',
            config: {}
          }
        },
        {
          id: 'connected-2', 
          type: 'processor',
          position: { x: 100, y: 0 },
          data: {
            label: 'Node 2',
            processorType: 'batch',
            category: 'system',
            config: {}
          }
        },
        {
          id: 'disconnected',
          type: 'processor',
          position: { x: 200, y: 0 },
          data: {
            label: 'Disconnected',
            processorType: 'groupbyattrs',
            category: 'aggregate',
            config: {}
          }
        }
      ]
      
      const edges: Edge[] = [
        { id: 'e1', source: 'connected-1', target: 'connected-2' }
      ]

      const validation = result.current.validatePipeline({ nodes, edges })

      expect(validation.valid).toBe(true) // Warnings don't make it invalid
      expect(validation.warnings).toContain('Disconnected is not connected to the pipeline')
    })

    it('should validate node configurations', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'invalid-topk',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Invalid TopK',
            processorType: 'filter/topk',
            category: 'filter',
            config: {
              k: 0 // Invalid - should be >= 1
            }
          }
        }
      ]

      const validation = result.current.validatePipeline({
        nodes,
        edges: []
      })

      expect(validation.valid).toBe(false)
      expect(validation.errors).toContain('Invalid TopK: Top K value must be at least 1')
    })

    it('should validate memory limiter format', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'invalid-memory',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Invalid Memory Limiter',
            processorType: 'memory_limiter',
            category: 'system',
            config: {
              limit: 'invalid-format' // Should be like "512MiB"
            }
          }
        }
      ]

      const validation = result.current.validatePipeline({
        nodes,
        edges: []
      })

      expect(validation.valid).toBe(false)
      expect(validation.errors).toContain('Invalid Memory Limiter: Memory limit must be in format like "512MiB" or "1GiB"')
    })

    it('should warn about processor ordering', () => {
      const { result } = renderHook(() => usePipelineGenerator())
      
      const nodes: Node<ProcessorNodeData>[] = [
        {
          id: 'batch-first',
          type: 'processor',
          position: { x: 0, y: 0 },
          data: {
            label: 'Batch (should be last)',
            processorType: 'batch',
            category: 'system',
            config: {}
          }
        },
        {
          id: 'memory-second',
          type: 'processor',
          position: { x: 100, y: 0 },
          data: {
            label: 'Memory Limiter (should be first)',
            processorType: 'memory_limiter',
            category: 'system',
            config: { limit: '512MiB' }
          }
        }
      ]
      
      const edges: Edge[] = [
        { id: 'e1', source: 'batch-first', target: 'memory-second' }
      ]

      const validation = result.current.validatePipeline({ nodes, edges })

      expect(validation.valid).toBe(true) // Warnings don't make it invalid
      expect(validation.warnings).toContain('memory_limiter processor should be placed first in the pipeline')
      expect(validation.warnings).toContain('batch processor should be placed last in the pipeline')
    })
  })
})