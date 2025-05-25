import React from 'react'
import { Box, Paper, Typography, Divider, Tooltip } from '@mui/material'
import {
  FilterAlt as FilterIcon,
  Transform as TransformIcon,
  GroupWork as GroupIcon,
  Memory as MemoryIcon,
  BatchPrediction as BatchIcon,
  SampleData as SampleIcon,
  Analytics as MetricsIcon,
  Extension as ExtensionIcon,
} from '@mui/icons-material'

interface ProcessorDefinition {
  type: string
  label: string
  icon: React.ReactNode
  category: 'filter' | 'transform' | 'aggregate' | 'system' | 'export'
  description: string
  configSchema?: any
}

const processors: ProcessorDefinition[] = [
  // Filters
  {
    type: 'filter/priority',
    label: 'Priority Filter',
    icon: <FilterIcon />,
    category: 'filter',
    description: 'Filter processes by priority (critical, high, medium, low)',
    configSchema: {
      minPriority: {
        type: 'select',
        options: ['critical', 'high', 'medium', 'low'],
        default: 'high',
        description: 'Minimum priority to retain',
      },
      excludePatterns: {
        type: 'string[]',
        default: [],
        description: 'Process name patterns to exclude',
      },
    },
  },
  {
    type: 'filter/resource',
    label: 'Resource Filter',
    icon: <FilterIcon />,
    category: 'filter',
    description: 'Filter processes by CPU/memory usage thresholds',
    configSchema: {
      minCpu: {
        type: 'number',
        default: 5.0,
        description: 'Minimum CPU usage percentage',
      },
      minMemory: {
        type: 'number',
        default: 50,
        description: 'Minimum memory usage in MB',
      },
    },
  },
  {
    type: 'filter/topk',
    label: 'Top-K Filter',
    icon: <FilterIcon />,
    category: 'filter',
    description: 'Keep only top K processes by resource usage',
    configSchema: {
      k: {
        type: 'number',
        default: 100,
        description: 'Number of top processes to keep',
      },
      metric: {
        type: 'select',
        options: ['cpu', 'memory', 'combined'],
        default: 'combined',
        description: 'Metric to sort by',
      },
    },
  },

  // Transforms
  {
    type: 'transform/classify',
    label: 'Process Classifier',
    icon: <TransformIcon />,
    category: 'transform',
    description: 'Classify processes into priority categories',
    configSchema: {
      rules: {
        type: 'rules',
        default: [],
        description: 'Classification rules',
      },
    },
  },
  {
    type: 'transform/normalize',
    label: 'Name Normalizer',
    icon: <TransformIcon />,
    category: 'transform',
    description: 'Normalize process names for consistency',
    configSchema: {
      removeVersions: {
        type: 'boolean',
        default: true,
        description: 'Remove version numbers from names',
      },
      removePids: {
        type: 'boolean',
        default: true,
        description: 'Remove PIDs from names',
      },
    },
  },
  {
    type: 'transform/enrich',
    label: 'Metadata Enricher',
    icon: <TransformIcon />,
    category: 'transform',
    description: 'Add metadata like department, team, or service',
    configSchema: {
      enrichmentRules: {
        type: 'map',
        default: {},
        description: 'Enrichment mappings',
      },
    },
  },

  // Aggregators
  {
    type: 'groupbyattrs',
    label: 'Group By Attributes',
    icon: <GroupIcon />,
    category: 'aggregate',
    description: 'Aggregate metrics by process attributes',
    configSchema: {
      keys: {
        type: 'string[]',
        default: ['process.name', 'host.name'],
        description: 'Attributes to group by',
      },
      aggregations: {
        type: 'select[]',
        options: ['sum', 'mean', 'max', 'min', 'count'],
        default: ['sum'],
        description: 'Aggregation functions',
      },
    },
  },
  {
    type: 'aggregate/rollup',
    label: 'Process Rollup',
    icon: <GroupIcon />,
    category: 'aggregate',
    description: 'Roll up similar processes into groups',
    configSchema: {
      patterns: {
        type: 'patterns',
        default: [],
        description: 'Rollup patterns',
      },
      keepOriginal: {
        type: 'boolean',
        default: false,
        description: 'Keep original metrics alongside rollups',
      },
    },
  },

  // System processors
  {
    type: 'memory_limiter',
    label: 'Memory Limiter',
    icon: <MemoryIcon />,
    category: 'system',
    description: 'Prevent out-of-memory situations',
    configSchema: {
      checkInterval: {
        type: 'duration',
        default: '1s',
        description: 'Memory check interval',
      },
      limit: {
        type: 'string',
        default: '512MiB',
        description: 'Memory limit',
      },
    },
  },
  {
    type: 'batch',
    label: 'Batch Processor',
    icon: <BatchIcon />,
    category: 'system',
    description: 'Batch metrics before export',
    configSchema: {
      timeout: {
        type: 'duration',
        default: '10s',
        description: 'Batch timeout',
      },
      sendBatchSize: {
        type: 'number',
        default: 1000,
        description: 'Batch size',
      },
    },
  },
  {
    type: 'sample',
    label: 'Sampler',
    icon: <SampleIcon />,
    category: 'system',
    description: 'Sample metrics to reduce volume',
    configSchema: {
      samplingRate: {
        type: 'number',
        default: 0.1,
        description: 'Sampling rate (0-1)',
      },
    },
  },

  // Custom/Extension
  {
    type: 'metricstransform',
    label: 'Metrics Transform',
    icon: <MetricsIcon />,
    category: 'export',
    description: 'Transform metrics before export',
    configSchema: {
      operations: {
        type: 'operations',
        default: [],
        description: 'Transformation operations',
      },
    },
  },
  {
    type: 'extension/custom',
    label: 'Custom Processor',
    icon: <ExtensionIcon />,
    category: 'export',
    description: 'Custom processor configuration',
    configSchema: {
      config: {
        type: 'yaml',
        default: '',
        description: 'Custom configuration',
      },
    },
  },
]

interface ProcessorLibraryProps {
  onDragStart?: (processor: ProcessorDefinition) => void
}

export const ProcessorLibrary: React.FC<ProcessorLibraryProps> = ({ onDragStart }) => {
  const handleDragStart = (event: React.DragEvent, processor: ProcessorDefinition) => {
    event.dataTransfer.setData('application/reactflow', processor.type)
    event.dataTransfer.setData('processor/config', JSON.stringify(processor))
    event.dataTransfer.effectAllowed = 'move'
    onDragStart?.(processor)
  }

  const categories = ['filter', 'transform', 'aggregate', 'system', 'export'] as const
  
  return (
    <Paper sx={{ p: 2, height: '100%', overflowY: 'auto' }}>
      <Typography variant="h6" gutterBottom>
        Processor Library
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Drag processors to the canvas to build your pipeline
      </Typography>
      
      {categories.map((category) => (
        <Box key={category} sx={{ mb: 3 }}>
          <Typography 
            variant="subtitle2" 
            sx={{ 
              mb: 1, 
              textTransform: 'capitalize',
              fontWeight: 600,
              color: 'text.secondary'
            }}
          >
            {category}
          </Typography>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            {processors
              .filter((p) => p.category === category)
              .map((processor) => (
                <Tooltip
                  key={processor.type}
                  title={processor.description}
                  placement="right"
                  arrow
                >
                  <Paper
                    variant="outlined"
                    sx={{
                      p: 1.5,
                      cursor: 'grab',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                      transition: 'all 0.2s',
                      '&:hover': {
                        backgroundColor: 'action.hover',
                        borderColor: 'primary.main',
                        transform: 'translateX(4px)',
                      },
                      '&:active': {
                        cursor: 'grabbing',
                      },
                    }}
                    draggable
                    onDragStart={(e) => handleDragStart(e, processor)}
                  >
                    <Box sx={{ color: 'primary.main' }}>{processor.icon}</Box>
                    <Typography variant="body2">{processor.label}</Typography>
                  </Paper>
                </Tooltip>
              ))}
          </Box>
          {category !== 'export' && <Divider sx={{ mt: 2 }} />}
        </Box>
      ))}
    </Paper>
  )
}