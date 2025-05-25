import React, { useState, useEffect } from 'react'
import {
  Drawer,
  Box,
  Typography,
  TextField,
  Select,
  MenuItem,
  FormControl,
  FormLabel,
  Switch,
  Button,
  IconButton,
  Divider,
  Chip,
  FormHelperText,
  Alert,
} from '@mui/material'
import {
  Close as CloseIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material'
import { Node } from 'react-flow-renderer'
import { ProcessorNodeData } from './ProcessorNode'

interface ConfigurationPanelProps {
  node: Node<ProcessorNodeData>
  onUpdate: (nodeId: string, data: ProcessorNodeData) => void
  onClose: () => void
}

export const ConfigurationPanel: React.FC<ConfigurationPanelProps> = ({
  node,
  onUpdate,
  onClose,
}) => {
  const [config, setConfig] = useState(node.data.config || {})
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Get config schema based on processor type
  const getConfigSchema = () => {
    // This would normally come from a schema registry
    const schemas: Record<string, any> = {
      'filter/priority': {
        minPriority: {
          type: 'select',
          label: 'Minimum Priority',
          options: ['critical', 'high', 'medium', 'low'],
          default: 'high',
          required: true,
          description: 'Only processes with this priority or higher will pass through',
        },
        excludePatterns: {
          type: 'string[]',
          label: 'Exclude Patterns',
          default: [],
          description: 'Regular expressions for process names to exclude',
        },
      },
      'filter/resource': {
        minCpu: {
          type: 'number',
          label: 'Minimum CPU %',
          default: 5.0,
          min: 0,
          max: 100,
          description: 'Filter out processes using less CPU than this',
        },
        minMemory: {
          type: 'number',
          label: 'Minimum Memory (MB)',
          default: 50,
          min: 0,
          description: 'Filter out processes using less memory than this',
        },
        operator: {
          type: 'select',
          label: 'Operator',
          options: ['AND', 'OR'],
          default: 'OR',
          description: 'How to combine CPU and memory filters',
        },
      },
      'filter/topk': {
        k: {
          type: 'number',
          label: 'Top K Processes',
          default: 100,
          min: 1,
          max: 1000,
          required: true,
          description: 'Number of top processes to keep',
        },
        metric: {
          type: 'select',
          label: 'Sort Metric',
          options: ['cpu', 'memory', 'combined'],
          default: 'combined',
          description: 'Metric to use for ranking processes',
        },
        includeIdle: {
          type: 'boolean',
          label: 'Include Idle Processes',
          default: false,
          description: 'Whether to include processes with 0% resource usage',
        },
      },
      'transform/classify': {
        rules: {
          type: 'rules',
          label: 'Classification Rules',
          default: [
            { pattern: 'postgres|mysql|redis', priority: 'critical' },
            { pattern: 'nginx|apache', priority: 'high' },
            { pattern: 'python|node|java', priority: 'medium' },
          ],
        },
        defaultPriority: {
          type: 'select',
          label: 'Default Priority',
          options: ['critical', 'high', 'medium', 'low'],
          default: 'low',
          description: 'Priority for processes that dont match any rule',
        },
      },
      'groupbyattrs': {
        keys: {
          type: 'string[]',
          label: 'Group By Keys',
          default: ['process.name', 'host.name'],
          required: true,
          description: 'Attributes to group metrics by',
        },
        aggregations: {
          type: 'multiselect',
          label: 'Aggregation Functions',
          options: ['sum', 'mean', 'max', 'min', 'count'],
          default: ['sum'],
          description: 'How to aggregate grouped metrics',
        },
      },
      'memory_limiter': {
        checkInterval: {
          type: 'duration',
          label: 'Check Interval',
          default: '1s',
          description: 'How often to check memory usage',
        },
        limit: {
          type: 'string',
          label: 'Memory Limit',
          default: '512MiB',
          pattern: '^\\d+[KMG]i?B$',
          description: 'Maximum memory usage (e.g., 512MiB, 1GiB)',
        },
      },
      'batch': {
        timeout: {
          type: 'duration',
          label: 'Batch Timeout',
          default: '10s',
          description: 'Maximum time to wait before sending a batch',
        },
        sendBatchSize: {
          type: 'number',
          label: 'Batch Size',
          default: 1000,
          min: 1,
          max: 10000,
          description: 'Number of data points per batch',
        },
      },
    }

    return schemas[node.data.processorType] || {}
  }

  const schema = getConfigSchema()

  useEffect(() => {
    // Initialize config with defaults
    const initialConfig: Record<string, any> = {}
    Object.entries(schema).forEach(([key, field]: [string, any]) => {
      if (config[key] === undefined && field.default !== undefined) {
        initialConfig[key] = field.default
      }
    })
    if (Object.keys(initialConfig).length > 0) {
      setConfig((prev) => ({ ...initialConfig, ...prev }))
    }
  }, [node.data.processorType])

  const validateField = (key: string, value: any, fieldSchema: any) => {
    if (fieldSchema.required && !value) {
      return 'This field is required'
    }
    if (fieldSchema.type === 'number') {
      if (isNaN(value)) return 'Must be a number'
      if (fieldSchema.min !== undefined && value < fieldSchema.min) {
        return `Must be at least ${fieldSchema.min}`
      }
      if (fieldSchema.max !== undefined && value > fieldSchema.max) {
        return `Must be at most ${fieldSchema.max}`
      }
    }
    if (fieldSchema.pattern) {
      const regex = new RegExp(fieldSchema.pattern)
      if (!regex.test(value)) {
        return `Invalid format`
      }
    }
    return ''
  }

  const handleChange = (key: string, value: any) => {
    setConfig((prev) => ({ ...prev, [key]: value }))
    const error = validateField(key, value, schema[key])
    setErrors((prev) => ({ ...prev, [key]: error }))
  }

  const handleSave = () => {
    // Validate all fields
    const newErrors: Record<string, string> = {}
    let isValid = true

    Object.entries(schema).forEach(([key, fieldSchema]: [string, any]) => {
      const error = validateField(key, config[key], fieldSchema)
      if (error) {
        newErrors[key] = error
        isValid = false
      }
    })

    setErrors(newErrors)

    if (isValid) {
      onUpdate(node.id, {
        ...node.data,
        config,
        isValid: true,
        errors: [],
      })
      onClose()
    }
  }

  const renderField = (key: string, fieldSchema: any) => {
    const value = config[key]
    const error = errors[key]

    switch (fieldSchema.type) {
      case 'select':
        return (
          <FormControl fullWidth error={!!error}>
            <FormLabel>{fieldSchema.label}</FormLabel>
            <Select
              value={value || fieldSchema.default || ''}
              onChange={(e) => handleChange(key, e.target.value)}
              size="small"
            >
              {fieldSchema.options.map((option: string) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </Select>
            {fieldSchema.description && (
              <FormHelperText>{fieldSchema.description}</FormHelperText>
            )}
            {error && <FormHelperText error>{error}</FormHelperText>}
          </FormControl>
        )

      case 'multiselect':
        return (
          <FormControl fullWidth error={!!error}>
            <FormLabel>{fieldSchema.label}</FormLabel>
            <Select
              multiple
              value={value || fieldSchema.default || []}
              onChange={(e) => handleChange(key, e.target.value)}
              size="small"
              renderValue={(selected) => (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                  {(selected as string[]).map((value) => (
                    <Chip key={value} label={value} size="small" />
                  ))}
                </Box>
              )}
            >
              {fieldSchema.options.map((option: string) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </Select>
            {fieldSchema.description && (
              <FormHelperText>{fieldSchema.description}</FormHelperText>
            )}
          </FormControl>
        )

      case 'number':
        return (
          <TextField
            fullWidth
            label={fieldSchema.label}
            type="number"
            value={value || fieldSchema.default || ''}
            onChange={(e) => handleChange(key, parseFloat(e.target.value))}
            error={!!error}
            helperText={error || fieldSchema.description}
            size="small"
            InputProps={{
              inputProps: {
                min: fieldSchema.min,
                max: fieldSchema.max,
              },
            }}
          />
        )

      case 'boolean':
        return (
          <FormControl fullWidth>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <FormLabel sx={{ flex: 1 }}>{fieldSchema.label}</FormLabel>
              <Switch
                checked={value || fieldSchema.default || false}
                onChange={(e) => handleChange(key, e.target.checked)}
              />
            </Box>
            {fieldSchema.description && (
              <FormHelperText>{fieldSchema.description}</FormHelperText>
            )}
          </FormControl>
        )

      case 'string[]':
        return (
          <FormControl fullWidth error={!!error}>
            <FormLabel>{fieldSchema.label}</FormLabel>
            <Box sx={{ mt: 1 }}>
              {(value || []).map((item: string, index: number) => (
                <Box key={index} sx={{ display: 'flex', gap: 1, mb: 1 }}>
                  <TextField
                    fullWidth
                    value={item}
                    onChange={(e) => {
                      const newValue = [...(value || [])]
                      newValue[index] = e.target.value
                      handleChange(key, newValue)
                    }}
                    size="small"
                  />
                  <IconButton
                    size="small"
                    onClick={() => {
                      const newValue = [...(value || [])]
                      newValue.splice(index, 1)
                      handleChange(key, newValue)
                    }}
                  >
                    <DeleteIcon />
                  </IconButton>
                </Box>
              ))}
              <Button
                startIcon={<AddIcon />}
                onClick={() => handleChange(key, [...(value || []), ''])}
                size="small"
              >
                Add Item
              </Button>
            </Box>
            {fieldSchema.description && (
              <FormHelperText>{fieldSchema.description}</FormHelperText>
            )}
          </FormControl>
        )

      case 'rules':
        return (
          <FormControl fullWidth>
            <FormLabel>{fieldSchema.label}</FormLabel>
            <Alert severity="info" sx={{ mt: 1 }}>
              Rules editor coming soon. Using default rules for now.
            </Alert>
          </FormControl>
        )

      default:
        return (
          <TextField
            fullWidth
            label={fieldSchema.label}
            value={value || fieldSchema.default || ''}
            onChange={(e) => handleChange(key, e.target.value)}
            error={!!error}
            helperText={error || fieldSchema.description}
            size="small"
          />
        )
    }
  }

  return (
    <Drawer
      anchor="right"
      open={true}
      onClose={onClose}
      PaperProps={{
        sx: { width: 400 },
      }}
    >
      <Box sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
          <Typography variant="h6" sx={{ flex: 1 }}>
            Configure {node.data.label}
          </Typography>
          <IconButton onClick={onClose}>
            <CloseIcon />
          </IconButton>
        </Box>

        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          {node.data.processorType}
        </Typography>

        <Divider sx={{ mb: 3 }} />

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          {Object.entries(schema).map(([key, fieldSchema]) => (
            <Box key={key}>{renderField(key, fieldSchema)}</Box>
          ))}
        </Box>

        <Box sx={{ mt: 4, display: 'flex', gap: 2 }}>
          <Button variant="contained" onClick={handleSave} fullWidth>
            Save Configuration
          </Button>
          <Button variant="outlined" onClick={onClose} fullWidth>
            Cancel
          </Button>
        </Box>
      </Box>
    </Drawer>
  )
}