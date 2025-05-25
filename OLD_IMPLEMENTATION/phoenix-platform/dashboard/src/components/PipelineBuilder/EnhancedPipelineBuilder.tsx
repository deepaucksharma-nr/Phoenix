import React, { useState, useCallback, useRef, useEffect } from 'react'
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  Connection,
  ReactFlowProvider,
  Panel,
  MarkerType,
  NodeTypes,
  Handle,
  Position,
  NodeProps,
} from 'reactflow'
import 'reactflow/dist/style.css'
import {
  Box,
  Paper,
  Typography,
  Button,
  IconButton,
  Tooltip,
  Chip,
  Card,
  CardContent,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Alert,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  SpeedDial,
  SpeedDialAction,
  SpeedDialIcon,
  Drawer,
  Switch,
  FormControlLabel,
  Badge,
  Fab,
  Zoom,
  Snackbar,
} from '@mui/material'
import {
  Save as SaveIcon,
  PlayArrow as PlayIcon,
  Delete as DeleteIcon,
  ContentCopy as CopyIcon,
  Download as DownloadIcon,
  Upload as UploadIcon,
  ZoomIn as ZoomInIcon,
  ZoomOut as ZoomOutIcon,
  CenterFocusStrong as CenterIcon,
  Settings as SettingsIcon,
  Help as HelpIcon,
  Code as CodeIcon,
  Warning as WarningIcon,
  CheckCircle as CheckIcon,
  Timeline as TimelineIcon,
  FilterList as FilterIcon,
  Transform as TransformIcon,
  Storage as StorageIcon,
  Speed as SpeedIcon,
  TrendingDown as TrendingDownIcon,
  AutoAwesome as AutoAwesomeIcon,
  Close as CloseIcon,
} from '@mui/icons-material'
import { useNotification } from '../../hooks/useNotification'

// Processor categories
const PROCESSOR_CATEGORIES = {
  source: { color: '#4CAF50', icon: <StorageIcon /> },
  filter: { color: '#2196F3', icon: <FilterIcon /> },
  transform: { color: '#FF9800', icon: <TransformIcon /> },
  aggregate: { color: '#9C27B0', icon: <TimelineIcon /> },
  sink: { color: '#F44336', icon: <TrendingDownIcon /> },
}

// Available processors
const AVAILABLE_PROCESSORS = [
  {
    id: 'source-metrics',
    category: 'source',
    name: 'Metrics Source',
    description: 'Collect metrics from OpenTelemetry',
    parameters: ['port', 'protocol'],
  },
  {
    id: 'filter-attributes',
    category: 'filter',
    name: 'Attribute Filter',
    description: 'Filter metrics by attributes',
    parameters: ['include', 'exclude'],
  },
  {
    id: 'filter-resource',
    category: 'filter',
    name: 'Resource Filter',
    description: 'Filter by resource attributes',
    parameters: ['resource_attributes'],
  },
  {
    id: 'transform-aggregate',
    category: 'aggregate',
    name: 'Time Aggregation',
    description: 'Aggregate metrics over time windows',
    parameters: ['window', 'aggregation_method'],
  },
  {
    id: 'transform-sample',
    category: 'transform',
    name: 'Sampling',
    description: 'Sample metrics to reduce volume',
    parameters: ['sampling_rate', 'hash_seed'],
  },
  {
    id: 'transform-topk',
    category: 'aggregate',
    name: 'Top-K',
    description: 'Keep only top K series by value',
    parameters: ['k', 'metric_name', 'order'],
  },
  {
    id: 'sink-prometheus',
    category: 'sink',
    name: 'Prometheus Export',
    description: 'Export to Prometheus',
    parameters: ['endpoint', 'namespace'],
  },
]

// Custom node component
const ProcessorNode: React.FC<NodeProps> = ({ data, selected }) => {
  const category = PROCESSOR_CATEGORIES[data.category as keyof typeof PROCESSOR_CATEGORIES]
  
  return (
    <Card
      sx={{
        minWidth: 200,
        borderColor: selected ? 'primary.main' : category.color,
        borderWidth: 2,
        borderStyle: 'solid',
        backgroundColor: selected ? 'action.selected' : 'background.paper',
      }}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ background: category.color }}
      />
      <CardContent sx={{ p: 1.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
          <Box sx={{ color: category.color, mr: 1 }}>{category.icon}</Box>
          <Typography variant="subtitle2" noWrap>
            {data.label}
          </Typography>
        </Box>
        <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
          {data.description}
        </Typography>
        {data.parameters && Object.keys(data.parameters).length > 0 && (
          <Box sx={{ mt: 1 }}>
            <Chip
              label={`${Object.keys(data.parameters).length} params`}
              size="small"
              color="primary"
              variant="outlined"
            />
          </Box>
        )}
      </CardContent>
      <Handle
        type="source"
        position={Position.Right}
        style={{ background: category.color }}
      />
    </Card>
  )
}

const nodeTypes: NodeTypes = {
  processor: ProcessorNode,
}

interface EnhancedPipelineBuilderProps {
  initialPipeline?: any
  onSave?: (pipeline: any) => void
  embedded?: boolean
  experimentMode?: boolean
  onValidate?: (isValid: boolean, errors: string[]) => void
}

export const EnhancedPipelineBuilder: React.FC<EnhancedPipelineBuilderProps> = ({
  initialPipeline,
  onSave,
  embedded = false,
  experimentMode = false,
  onValidate,
}) => {
  const { showNotification } = useNotification()
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [configDialogOpen, setConfigDialogOpen] = useState(false)
  const [nodeConfig, setNodeConfig] = useState<Record<string, any>>({})
  const [libraryOpen, setLibraryOpen] = useState(!embedded)
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [speedDialOpen, setSpeedDialOpen] = useState(false)
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const [reactFlowInstance, setReactFlowInstance] = useState<any>(null)

  // Load initial pipeline
  useEffect(() => {
    if (initialPipeline) {
      // Convert pipeline config to nodes and edges
      // This would parse the YAML/JSON config
      loadPipelineConfig(initialPipeline)
    }
  }, [initialPipeline])

  const loadPipelineConfig = (config: any) => {
    // Implementation to convert config to nodes/edges
    showNotification('Pipeline loaded successfully', 'success')
  }

  const onConnect = useCallback(
    (params: Connection) => {
      const newEdge = {
        ...params,
        type: 'smoothstep',
        animated: true,
        markerEnd: {
          type: MarkerType.ArrowClosed,
        },
      }
      setEdges((eds) => addEdge(newEdge, eds))
    },
    [setEdges]
  )

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const type = event.dataTransfer.getData('application/reactflow')
      if (!type || !reactFlowInstance || !reactFlowWrapper.current) return

      const reactFlowBounds = reactFlowWrapper.current.getBoundingClientRect()
      const position = reactFlowInstance.project({
        x: event.clientX - reactFlowBounds.left,
        y: event.clientY - reactFlowBounds.top,
      })

      const processorInfo = AVAILABLE_PROCESSORS.find(p => p.id === type)
      if (!processorInfo) return

      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type: 'processor',
        position,
        data: {
          label: processorInfo.name,
          category: processorInfo.category,
          description: processorInfo.description,
          parameters: {},
        },
      }

      setNodes((nds) => nds.concat(newNode))
      showNotification(`Added ${processorInfo.name} to pipeline`, 'info')
    },
    [reactFlowInstance, setNodes, showNotification]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
    setNodeConfig(node.data.parameters || {})
    setConfigDialogOpen(true)
  }, [])

  const handleDeleteNode = useCallback(() => {
    if (selectedNode) {
      setNodes((nds) => nds.filter((n) => n.id !== selectedNode.id))
      setEdges((eds) => eds.filter((e) => e.source !== selectedNode.id && e.target !== selectedNode.id))
      showNotification('Processor removed', 'info')
      setConfigDialogOpen(false)
      setSelectedNode(null)
    }
  }, [selectedNode, setNodes, setEdges, showNotification])

  const handleSaveConfig = () => {
    if (selectedNode) {
      setNodes((nds) =>
        nds.map((node) => {
          if (node.id === selectedNode.id) {
            return {
              ...node,
              data: {
                ...node.data,
                parameters: nodeConfig,
              },
            }
          }
          return node
        })
      )
      showNotification('Configuration saved', 'success')
      setConfigDialogOpen(false)
    }
  }

  const validatePipeline = useCallback(() => {
    const errors: string[] = []

    // Check if pipeline has source and sink
    const hasSource = nodes.some(n => n.data.category === 'source')
    const hasSink = nodes.some(n => n.data.category === 'sink')

    if (!hasSource) errors.push('Pipeline must have at least one source')
    if (!hasSink) errors.push('Pipeline must have at least one sink')

    // Check for disconnected nodes
    const connectedNodes = new Set<string>()
    edges.forEach(edge => {
      connectedNodes.add(edge.source)
      connectedNodes.add(edge.target)
    })

    nodes.forEach(node => {
      if (!connectedNodes.has(node.id) && nodes.length > 1) {
        errors.push(`${node.data.label} is not connected`)
      }
    })

    // Check for cycles
    // Simple cycle detection would go here

    setValidationErrors(errors)
    onValidate?.(errors.length === 0, errors)
    return errors.length === 0
  }, [nodes, edges, onValidate])

  const handleSavePipeline = () => {
    if (!validatePipeline()) {
      showNotification('Pipeline has validation errors', 'error')
      return
    }

    const pipelineConfig = {
      processors: nodes.map(node => ({
        id: node.id,
        type: node.data.category,
        name: node.data.label,
        parameters: node.data.parameters,
      })),
      connections: edges.map(edge => ({
        from: edge.source,
        to: edge.target,
      })),
    }

    onSave?.(pipelineConfig)
    showNotification('Pipeline saved successfully', 'success')
  }

  const exportPipeline = () => {
    const pipelineConfig = {
      nodes,
      edges,
      metadata: {
        created: new Date().toISOString(),
        version: '1.0',
      },
    }

    const blob = new Blob([JSON.stringify(pipelineConfig, null, 2)], {
      type: 'application/json',
    })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'pipeline-config.json'
    a.click()
    URL.revokeObjectURL(url)
    showNotification('Pipeline exported', 'success')
  }

  const speedDialActions = [
    { icon: <SaveIcon />, name: 'Save Pipeline', action: handleSavePipeline },
    { icon: <DownloadIcon />, name: 'Export', action: exportPipeline },
    { icon: <CopyIcon />, name: 'Duplicate', action: () => {} },
    { icon: <CodeIcon />, name: 'View YAML', action: () => {} },
  ]

  return (
    <Box sx={{ height: '100%', display: 'flex' }}>
      <ReactFlowProvider>
        <Drawer
          variant={embedded ? 'temporary' : 'persistent'}
          open={libraryOpen}
          onClose={() => setLibraryOpen(false)}
          sx={{
            width: libraryOpen ? 280 : 0,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: 280,
              position: embedded ? 'absolute' : 'relative',
              height: '100%',
            },
          }}
        >
          <Box sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Processors
            </Typography>
            <Typography variant="caption" color="text.secondary">
              Drag processors to the canvas
            </Typography>
          </Box>
          <Divider />
          <List>
            {Object.entries(
              AVAILABLE_PROCESSORS.reduce((acc, proc) => {
                if (!acc[proc.category]) acc[proc.category] = []
                acc[proc.category].push(proc)
                return acc
              }, {} as Record<string, typeof AVAILABLE_PROCESSORS>)
            ).map(([category, processors]) => (
              <Box key={category}>
                <ListItem>
                  <ListItemIcon>
                    {PROCESSOR_CATEGORIES[category as keyof typeof PROCESSOR_CATEGORIES].icon}
                  </ListItemIcon>
                  <ListItemText
                    primary={category.charAt(0).toUpperCase() + category.slice(1)}
                    primaryTypographyProps={{ variant: 'subtitle2' }}
                  />
                </ListItem>
                {processors.map((processor) => (
                  <ListItem
                    key={processor.id}
                    sx={{
                      pl: 4,
                      cursor: 'grab',
                      '&:hover': { bgcolor: 'action.hover' },
                    }}
                    draggable
                    onDragStart={(e) => {
                      e.dataTransfer.setData('application/reactflow', processor.id)
                      e.dataTransfer.effectAllowed = 'move'
                    }}
                  >
                    <ListItemText
                      primary={processor.name}
                      secondary={processor.description}
                      primaryTypographyProps={{ variant: 'body2' }}
                      secondaryTypographyProps={{ variant: 'caption' }}
                    />
                  </ListItem>
                ))}
              </Box>
            ))}
          </List>
        </Drawer>

        <Box
          ref={reactFlowWrapper}
          sx={{ flex: 1, position: 'relative' }}
          onDrop={onDrop}
          onDragOver={onDragOver}
        >
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onInit={setReactFlowInstance}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background variant="dots" gap={12} size={1} />
            <Controls />
            <MiniMap
              nodeColor={(node) => {
                const category = node.data?.category
                return PROCESSOR_CATEGORIES[category as keyof typeof PROCESSOR_CATEGORIES]?.color || '#999'
              }}
            />
            
            {/* Validation Panel */}
            {validationErrors.length > 0 && (
              <Panel position="top-center">
                <Alert severity="error" sx={{ mb: 2 }}>
                  <Typography variant="subtitle2">Validation Errors:</Typography>
                  <ul style={{ margin: 0, paddingLeft: 20 }}>
                    {validationErrors.map((error, i) => (
                      <li key={i}>{error}</li>
                    ))}
                  </ul>
                </Alert>
              </Panel>
            )}

            {/* Toolbar */}
            <Panel position="top-left">
              <Paper sx={{ p: 1, display: 'flex', gap: 1 }}>
                {embedded && (
                  <Tooltip title="Show Processors">
                    <IconButton onClick={() => setLibraryOpen(true)}>
                      <AutoAwesomeIcon />
                    </IconButton>
                  </Tooltip>
                )}
                <Tooltip title="Validate Pipeline">
                  <IconButton onClick={validatePipeline}>
                    <CheckIcon />
                  </IconButton>
                </Tooltip>
                <Tooltip title="Center View">
                  <IconButton onClick={() => reactFlowInstance?.fitView()}>
                    <CenterIcon />
                  </IconButton>
                </Tooltip>
                <Divider orientation="vertical" flexItem />
                <Tooltip title="Help">
                  <IconButton>
                    <HelpIcon />
                  </IconButton>
                </Tooltip>
              </Paper>
            </Panel>

            {/* Status Panel */}
            <Panel position="bottom-left">
              <Paper sx={{ p: 1, display: 'flex', alignItems: 'center', gap: 2 }}>
                <Chip
                  label={`${nodes.length} processors`}
                  size="small"
                  color="primary"
                />
                <Chip
                  label={`${edges.length} connections`}
                  size="small"
                />
                {validationErrors.length === 0 && nodes.length > 0 && (
                  <Chip
                    icon={<CheckIcon />}
                    label="Valid"
                    size="small"
                    color="success"
                  />
                )}
              </Paper>
            </Panel>
          </ReactFlow>

          {/* Speed Dial */}
          {!embedded && (
            <SpeedDial
              ariaLabel="Pipeline actions"
              sx={{ position: 'absolute', bottom: 16, right: 16 }}
              icon={<SpeedDialIcon />}
              onClose={() => setSpeedDialOpen(false)}
              onOpen={() => setSpeedDialOpen(true)}
              open={speedDialOpen}
            >
              {speedDialActions.map((action) => (
                <SpeedDialAction
                  key={action.name}
                  icon={action.icon}
                  tooltipTitle={action.name}
                  onClick={() => {
                    setSpeedDialOpen(false)
                    action.action()
                  }}
                />
              ))}
            </SpeedDial>
          )}
        </Box>
      </ReactFlowProvider>

      {/* Node Configuration Dialog */}
      <Dialog
        open={configDialogOpen}
        onClose={() => setConfigDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            Configure {selectedNode?.data.label}
            <IconButton onClick={() => setConfigDialogOpen(false)}>
              <CloseIcon />
            </IconButton>
          </Box>
        </DialogTitle>
        <DialogContent>
          <Alert severity="info" sx={{ mb: 2 }}>
            Configure the processor parameters below. These settings will affect how the processor handles metrics.
          </Alert>
          
          {/* Dynamic parameter fields based on processor type */}
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <TextField
              fullWidth
              label="Processor Name"
              value={selectedNode?.data.label || ''}
              disabled
            />
            
            {selectedNode?.data.category === 'filter' && (
              <>
                <TextField
                  fullWidth
                  label="Include Pattern"
                  value={nodeConfig.include || ''}
                  onChange={(e) => setNodeConfig({ ...nodeConfig, include: e.target.value })}
                  helperText="Regex pattern to include metrics"
                />
                <TextField
                  fullWidth
                  label="Exclude Pattern"
                  value={nodeConfig.exclude || ''}
                  onChange={(e) => setNodeConfig({ ...nodeConfig, exclude: e.target.value })}
                  helperText="Regex pattern to exclude metrics"
                />
              </>
            )}
            
            {selectedNode?.data.category === 'aggregate' && (
              <>
                <FormControl fullWidth>
                  <InputLabel>Window Size</InputLabel>
                  <Select
                    value={nodeConfig.window || '60s'}
                    onChange={(e) => setNodeConfig({ ...nodeConfig, window: e.target.value })}
                    label="Window Size"
                  >
                    <MenuItem value="30s">30 seconds</MenuItem>
                    <MenuItem value="60s">1 minute</MenuItem>
                    <MenuItem value="300s">5 minutes</MenuItem>
                    <MenuItem value="900s">15 minutes</MenuItem>
                  </Select>
                </FormControl>
                <FormControl fullWidth>
                  <InputLabel>Aggregation Method</InputLabel>
                  <Select
                    value={nodeConfig.method || 'avg'}
                    onChange={(e) => setNodeConfig({ ...nodeConfig, method: e.target.value })}
                    label="Aggregation Method"
                  >
                    <MenuItem value="avg">Average</MenuItem>
                    <MenuItem value="sum">Sum</MenuItem>
                    <MenuItem value="min">Minimum</MenuItem>
                    <MenuItem value="max">Maximum</MenuItem>
                    <MenuItem value="p95">95th Percentile</MenuItem>
                  </Select>
                </FormControl>
              </>
            )}
            
            {selectedNode?.data.category === 'transform' && (
              <TextField
                fullWidth
                label="Sampling Rate"
                type="number"
                value={nodeConfig.sampling_rate || 0.1}
                onChange={(e) => setNodeConfig({ ...nodeConfig, sampling_rate: parseFloat(e.target.value) })}
                inputProps={{ min: 0, max: 1, step: 0.01 }}
                helperText="Percentage of metrics to keep (0.1 = 10%)"
              />
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button color="error" onClick={handleDeleteNode}>
            Delete Processor
          </Button>
          <Box sx={{ flex: 1 }} />
          <Button onClick={() => setConfigDialogOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSaveConfig}>
            Save Configuration
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}