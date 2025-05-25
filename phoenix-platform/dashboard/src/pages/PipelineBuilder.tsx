import React, { useState, useCallback } from 'react'
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  ReactFlowProvider,
  Panel,
} from 'reactflow'
import 'reactflow/dist/style.css'
import {
  Box,
  Paper,
  Typography,
  Button,
  IconButton,
  Tooltip,
  Alert,
  Snackbar,
} from '@mui/material'
import {
  Save as SaveIcon,
  PlayArrow as RunIcon,
  Clear as ClearIcon,
  Download as ExportIcon,
  Upload as ImportIcon,
} from '@mui/icons-material'

import { ProcessorNode, ProcessorNodeData } from '../components/PipelineBuilder/ProcessorNode'
import { ProcessorLibrary } from '../components/PipelineBuilder/ProcessorLibrary'
import { ConfigurationPanel } from '../components/PipelineBuilder/ConfigurationPanel'
import { apiService } from '../services/api.service'

const nodeTypes = {
  processor: ProcessorNode,
}

export const PipelineBuilder: React.FC = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = useState<Node<ProcessorNodeData> | null>(null)
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [snackbar, setSnackbar] = useState<{
    open: boolean
    message: string
    severity: 'success' | 'error' | 'info'
  }>({ open: false, message: '', severity: 'info' })

  const onConnect = useCallback(
    (params: Connection) => {
      setEdges((eds) => addEdge({ ...params, type: 'smoothstep' }, eds))
    },
    [setEdges]
  )

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const reactFlowBounds = event.currentTarget.getBoundingClientRect()
      const processorData = event.dataTransfer.getData('processor/config')
      
      if (!processorData) return

      const processor = JSON.parse(processorData)
      const position = {
        x: event.clientX - reactFlowBounds.left - 100,
        y: event.clientY - reactFlowBounds.top - 40,
      }

      const newNode: Node<ProcessorNodeData> = {
        id: `${processor.type}_${Date.now()}`,
        type: 'processor',
        position,
        data: {
          label: processor.label,
          processorType: processor.type,
          category: processor.category,
          config: {},
          icon: processor.icon,
        },
      }

      setNodes((nds) => nds.concat(newNode))
    },
    [setNodes]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node as Node<ProcessorNodeData>)
  }, [])

  const handleNodeUpdate = useCallback(
    (nodeId: string, data: ProcessorNodeData) => {
      setNodes((nds) =>
        nds.map((node) => {
          if (node.id === nodeId) {
            return { ...node, data }
          }
          return node
        })
      )
    },
    [setNodes]
  )

  const validatePipeline = useCallback(async () => {
    const errors: string[] = []

    // Check if pipeline is empty
    if (nodes.length === 0) {
      errors.push('Pipeline must have at least one processor')
    }

    // Check for memory_limiter as first processor
    const orderedNodes = getOrderedNodes()
    if (orderedNodes.length > 0 && !orderedNodes[0].data.processorType.includes('memory_limiter')) {
      errors.push('Pipeline should start with a memory_limiter processor')
    }

    // Check for batch processor as last
    if (orderedNodes.length > 0 && !orderedNodes[orderedNodes.length - 1].data.processorType.includes('batch')) {
      errors.push('Pipeline should end with a batch processor')
    }

    // Check for disconnected nodes
    const connectedNodeIds = new Set<string>()
    edges.forEach((edge) => {
      connectedNodeIds.add(edge.source)
      connectedNodeIds.add(edge.target)
    })
    
    nodes.forEach((node) => {
      if (!connectedNodeIds.has(node.id) && nodes.length > 1) {
        errors.push(`Processor "${node.data.label}" is not connected`)
      }
    })

    setValidationErrors(errors)
    return errors.length === 0
  }, [nodes, edges])

  const getOrderedNodes = (): Node<ProcessorNodeData>[] => {
    // Simple topological sort
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
      adjList.get(edge.source)?.push(edge.target)
    })

    // Find start nodes
    const queue: string[] = []
    inDegree.forEach((degree, nodeId) => {
      if (degree === 0) queue.push(nodeId)
    })

    // Process
    const ordered: Node<ProcessorNodeData>[] = []
    while (queue.length > 0) {
      const nodeId = queue.shift()!
      const node = nodeMap.get(nodeId)
      if (node) ordered.push(node)

      adjList.get(nodeId)?.forEach((neighbor) => {
        const newDegree = (inDegree.get(neighbor) || 1) - 1
        inDegree.set(neighbor, newDegree)
        if (newDegree === 0) queue.push(neighbor)
      })
    }

    return ordered
  }

  const generateYAML = (): string => {
    const orderedNodes = getOrderedNodes()
    
    // Build receivers section
    const receivers = {
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
    }

    // Build processors section
    const processors: Record<string, any> = {}
    const processorNames: string[] = []
    
    orderedNodes.forEach((node) => {
      const { processorType, config } = node.data
      const processorName = processorType.replace('/', '_')
      processors[processorName] = config
      processorNames.push(processorName)
    })

    // Build exporters section
    const exporters = {
      otlphttp: {
        endpoint: '${NEW_RELIC_OTLP_ENDPOINT}',
        headers: {
          'api-key': '${NEW_RELIC_API_KEY}',
        },
      },
      prometheus: {
        endpoint: '0.0.0.0:8888',
      },
    }

    // Build service section
    const service = {
      pipelines: {
        metrics: {
          receivers: ['hostmetrics'],
          processors: processorNames,
          exporters: ['otlphttp', 'prometheus'],
        },
      },
    }

    // Combine into final config
    const config = {
      receivers,
      processors,
      exporters,
      service,
    }

    // Convert to YAML (simplified - would use proper YAML library)
    return JSON.stringify(config, null, 2)
  }

  const handleSave = async () => {
    if (!(await validatePipeline())) {
      setSnackbar({
        open: true,
        message: 'Please fix validation errors before saving',
        severity: 'error',
      })
      return
    }

    try {
      // In a real implementation, this would save to the backend
      const pipelineConfig = {
        nodes,
        edges,
        yaml: generateYAML(),
      }
      
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))
      
      setSnackbar({
        open: true,
        message: 'Pipeline saved successfully',
        severity: 'success',
      })
    } catch (error) {
      setSnackbar({
        open: true,
        message: 'Failed to save pipeline',
        severity: 'error',
      })
    }
  }

  const handleRun = async () => {
    if (!(await validatePipeline())) {
      setSnackbar({
        open: true,
        message: 'Please fix validation errors before running',
        severity: 'error',
      })
      return
    }

    setSnackbar({
      open: true,
      message: 'Creating experiment with this pipeline...',
      severity: 'info',
    })

    // Navigate to experiment creation with this pipeline
  }

  const handleClear = () => {
    setNodes([])
    setEdges([])
    setSelectedNode(null)
    setValidationErrors([])
  }

  const handleExport = () => {
    const yaml = generateYAML()
    const blob = new Blob([yaml], { type: 'application/x-yaml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'pipeline-config.yaml'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Box sx={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
      {/* Left sidebar - Processor Library */}
      <Box sx={{ width: 300, borderRight: 1, borderColor: 'divider' }}>
        <ProcessorLibrary />
      </Box>

      {/* Main canvas */}
      <Box sx={{ flex: 1, position: 'relative' }}>
        <ReactFlowProvider>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onDrop={onDrop}
            onDragOver={onDragOver}
            onNodeClick={onNodeClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Panel position="top-left">
              <Paper elevation={1} sx={{ p: 1 }}>
                <Box sx={{ display: 'flex', gap: 1 }}>
                  <Tooltip title="Save Pipeline">
                    <IconButton onClick={handleSave} color="primary">
                      <SaveIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Create Experiment">
                    <IconButton onClick={handleRun} color="success">
                      <RunIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Export YAML">
                    <IconButton onClick={handleExport}>
                      <ExportIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Clear Canvas">
                    <IconButton onClick={handleClear} color="error">
                      <ClearIcon />
                    </IconButton>
                  </Tooltip>
                </Box>
              </Paper>
            </Panel>

            <Panel position="top-center">
              <Typography variant="h5" sx={{ color: 'text.secondary' }}>
                Visual Pipeline Builder
              </Typography>
            </Panel>

            {validationErrors.length > 0 && (
              <Panel position="bottom-left">
                <Alert severity="warning" sx={{ maxWidth: 400 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Validation Issues:
                  </Typography>
                  {validationErrors.map((error, index) => (
                    <Typography key={index} variant="body2">
                      • {error}
                    </Typography>
                  ))}
                </Alert>
              </Panel>
            )}

            <Controls />
            <MiniMap />
            <Background variant="dots" gap={12} size={1} />
          </ReactFlow>
        </ReactFlowProvider>
      </Box>

      {/* Configuration Panel */}
      {selectedNode && (
        <ConfigurationPanel
          node={selectedNode}
          onUpdate={handleNodeUpdate}
          onClose={() => setSelectedNode(null)}
        />
      )}

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        message={snackbar.message}
      />
    </Box>
  )
}