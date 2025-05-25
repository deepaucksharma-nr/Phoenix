import React, { memo } from 'react'
import { Handle, Position, NodeProps } from 'react-flow-renderer'
import { Paper, Typography, Box, Chip, IconButton } from '@mui/material'
import {
  Settings as SettingsIcon,
  Error as ErrorIcon,
  CheckCircle as ValidIcon,
} from '@mui/icons-material'

export interface ProcessorNodeData {
  label: string
  processorType: string
  category: 'filter' | 'transform' | 'aggregate' | 'system' | 'export'
  config: Record<string, any>
  isValid?: boolean
  errors?: string[]
  icon?: React.ReactNode
}

const categoryColors = {
  filter: '#FF6B6B',
  transform: '#4ECDC4',
  aggregate: '#45B7D1',
  system: '#FFA07A',
  export: '#98D8C8',
}

export const ProcessorNode = memo<NodeProps<ProcessorNodeData>>(
  ({ data, selected, id }) => {
    const { label, processorType, category, config, isValid = true, errors = [] } = data

    const getCategoryColor = () => categoryColors[category] || '#666'

    return (
      <>
        <Handle
          type="target"
          position={Position.Top}
          style={{
            background: '#555',
            width: 10,
            height: 10,
          }}
        />
        
        <Paper
          elevation={selected ? 8 : 2}
          sx={{
            padding: 2,
            minWidth: 200,
            border: selected ? '2px solid' : '1px solid',
            borderColor: selected ? 'primary.main' : 'divider',
            borderRadius: 2,
            position: 'relative',
            backgroundColor: 'background.paper',
            transition: 'all 0.2s',
            '&:hover': {
              elevation: 4,
              borderColor: 'primary.light',
            },
          }}
        >
          {/* Category indicator */}
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              height: 4,
              backgroundColor: getCategoryColor(),
              borderRadius: '8px 8px 0 0',
            }}
          />

          {/* Header */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1, mt: 0.5 }}>
            {data.icon && (
              <Box sx={{ color: getCategoryColor(), display: 'flex' }}>
                {data.icon}
              </Box>
            )}
            <Typography variant="subtitle2" sx={{ fontWeight: 600, flex: 1 }}>
              {label}
            </Typography>
            
            {/* Status icon */}
            {!isValid && (
              <ErrorIcon color="error" fontSize="small" />
            )}
            {isValid && Object.keys(config).length > 0 && (
              <ValidIcon color="success" fontSize="small" />
            )}
          </Box>

          {/* Processor type */}
          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
            {processorType}
          </Typography>

          {/* Configuration summary */}
          {Object.keys(config).length > 0 && (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 1 }}>
              {Object.entries(config).slice(0, 3).map(([key, value]) => (
                <Chip
                  key={key}
                  label={`${key}: ${JSON.stringify(value)}`}
                  size="small"
                  variant="outlined"
                  sx={{ 
                    fontSize: '0.7rem',
                    height: 20,
                    '& .MuiChip-label': { px: 1 }
                  }}
                />
              ))}
              {Object.keys(config).length > 3 && (
                <Chip
                  label={`+${Object.keys(config).length - 3} more`}
                  size="small"
                  sx={{ 
                    fontSize: '0.7rem',
                    height: 20,
                    '& .MuiChip-label': { px: 1 }
                  }}
                />
              )}
            </Box>
          )}

          {/* Error display */}
          {errors.length > 0 && (
            <Box sx={{ mt: 1 }}>
              {errors.map((error, index) => (
                <Typography
                  key={index}
                  variant="caption"
                  color="error"
                  sx={{ display: 'block' }}
                >
                  • {error}
                </Typography>
              ))}
            </Box>
          )}

          {/* Settings button */}
          <IconButton
            size="small"
            sx={{
              position: 'absolute',
              top: 8,
              right: 8,
              opacity: 0.7,
              '&:hover': { opacity: 1 },
            }}
            onClick={(e) => {
              e.stopPropagation()
              // Configuration will be handled by parent
            }}
          >
            <SettingsIcon fontSize="small" />
          </IconButton>
        </Paper>

        <Handle
          type="source"
          position={Position.Bottom}
          style={{
            background: '#555',
            width: 10,
            height: 10,
          }}
        />
      </>
    )
  }
)

ProcessorNode.displayName = 'ProcessorNode'