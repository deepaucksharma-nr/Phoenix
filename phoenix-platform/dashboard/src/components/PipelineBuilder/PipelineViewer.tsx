import React from 'react'
import {
  Box,
  Paper,
  Typography,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Grid,
} from '@mui/material'
import {
  ExpandMore,
  FilterList,
  Transform,
  Memory,
  Timer,
  GroupWork,
  Assignment,
  Send,
  Storage,
} from '@mui/icons-material'
import { PipelineConfig } from '../../types'

interface PipelineViewerProps {
  pipeline: PipelineConfig
  compact?: boolean
}

const getProcessorIcon = (type: string) => {
  if (type.includes('filter')) return <FilterList />
  if (type.includes('transform')) return <Transform />
  if (type.includes('memory')) return <Memory />
  if (type.includes('batch')) return <Timer />
  if (type.includes('group')) return <GroupWork />
  if (type.includes('attributes')) return <Assignment />
  return <Transform />
}

const getProcessorColor = (type: string): any => {
  if (type.includes('filter')) return 'error'
  if (type.includes('transform')) return 'primary'
  if (type.includes('memory')) return 'warning'
  if (type.includes('batch')) return 'info'
  if (type.includes('group')) return 'success'
  return 'default'
}

export const PipelineViewer: React.FC<PipelineViewerProps> = ({ pipeline, compact = false }) => {
  if (!pipeline) {
    return (
      <Paper sx={{ p: 2, bgcolor: 'grey.100' }}>
        <Typography variant="body2" color="text.secondary">
          No pipeline configuration
        </Typography>
      </Paper>
    )
  }

  const renderProcessorConfig = (config: any) => {
    if (!config || typeof config !== 'object') return null

    return (
      <Box sx={{ mt: 1 }}>
        {Object.entries(config).map(([key, value]) => (
          <Box key={key} sx={{ display: 'flex', alignItems: 'baseline', mb: 0.5 }}>
            <Typography variant="caption" color="text.secondary" sx={{ mr: 1 }}>
              {key}:
            </Typography>
            <Typography variant="caption">
              {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
            </Typography>
          </Box>
        ))}
      </Box>
    )
  }

  if (compact) {
    return (
      <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
        {pipeline.processors?.map((processor, index) => (
          <Chip
            key={index}
            icon={getProcessorIcon(processor.type)}
            label={processor.type}
            size="small"
            color={getProcessorColor(processor.type)}
            variant="outlined"
          />
        ))}
      </Box>
    )
  }

  return (
    <Box>
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={12} sm={6}>
          <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Storage sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2">Receivers</Typography>
            </Box>
            <List dense>
              {pipeline.receivers?.map((receiver, index) => (
                <ListItem key={index}>
                  <ListItemText
                    primary={receiver}
                    primaryTypographyProps={{ variant: 'caption' }}
                  />
                </ListItem>
              ))}
            </List>
          </Paper>
        </Grid>
        <Grid item xs={12} sm={6}>
          <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Send sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2">Exporters</Typography>
            </Box>
            <List dense>
              {pipeline.exporters?.map((exporter, index) => (
                <ListItem key={index}>
                  <ListItemText
                    primary={exporter}
                    primaryTypographyProps={{ variant: 'caption' }}
                  />
                </ListItem>
              ))}
            </List>
          </Paper>
        </Grid>
      </Grid>

      <Typography variant="subtitle2" sx={{ mb: 1 }}>
        Processors ({pipeline.processors?.length || 0})
      </Typography>
      {pipeline.processors?.map((processor, index) => (
        <Accordion key={index} defaultExpanded={index === 0}>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
              {getProcessorIcon(processor.type)}
              <Typography sx={{ ml: 1, flexGrow: 1 }}>
                {processor.type}
              </Typography>
              <Chip
                label={`Step ${index + 1}`}
                size="small"
                color={getProcessorColor(processor.type)}
              />
            </Box>
          </AccordionSummary>
          <AccordionDetails>
            {processor.config ? (
              renderProcessorConfig(processor.config)
            ) : (
              <Typography variant="caption" color="text.secondary">
                No configuration
              </Typography>
            )}
          </AccordionDetails>
        </Accordion>
      ))}
    </Box>
  )
}