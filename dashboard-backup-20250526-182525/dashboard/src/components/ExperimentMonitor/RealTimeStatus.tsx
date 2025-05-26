import React, { useEffect, useState } from 'react'
import {
  Box,
  Card,
  CardContent,
  Typography,
  Chip,
  LinearProgress,
  Grid,
  Alert,
  IconButton,
  Tooltip,
  Collapse,
} from '@mui/material'
import {
  PlayArrow,
  Pause,
  Stop,
  Refresh,
  Warning,
  CheckCircle,
  Error as ErrorIcon,
  ExpandMore,
  ExpandLess,
} from '@mui/icons-material'
import { useWebSocket } from '../../hooks/useWebSocket'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { updateExperiment } from '@store/slices/experimentSlice'
import { Experiment, ExperimentStatus } from '../../types'

interface RealTimeStatusProps {
  experimentId: string
  autoRefresh?: boolean
  showDetails?: boolean
}

interface MetricsUpdate {
  experimentId: string
  metrics: {
    throughput: number
    latency: number
    errorRate: number
    metricsVolume: number
    reductionPercentage: number
  }
  timestamp: string
}

interface ExperimentUpdate {
  experimentId: string
  status: ExperimentStatus
  progress?: number
  message?: string
  logs?: string[]
}

export const RealTimeStatus: React.FC<RealTimeStatusProps> = ({
  experimentId,
  autoRefresh = true,
  showDetails = true,
}) => {
  const { subscribe } = useWebSocket()
  const dispatch = useAppDispatch()
  const experiments = useAppSelector((state) => state.experiments.experiments)
  const [metrics, setMetrics] = useState<MetricsUpdate | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const [expanded, setExpanded] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'reconnecting'>('disconnected')

  const experiment = experiments.find(exp => exp.id === experimentId)

  useEffect(() => {
    if (!autoRefresh) return

    // Subscribe to experiment updates
    const unsubscribeExperiment = subscribe(`experiment.${experimentId}`, (update: ExperimentUpdate) => {
      console.log('Experiment update received:', update)
      setLastUpdate(new Date())
      
      if (update.experimentId === experimentId && experiment) {
        dispatch(updateExperiment({
          ...experiment,
          status: update.status,
          progress: update.progress,
        }))
      }
    })

    // Subscribe to metrics updates
    const unsubscribeMetrics = subscribe(`metrics.${experimentId}`, (update: MetricsUpdate) => {
      console.log('Metrics update received:', update)
      setMetrics(update)
      setLastUpdate(new Date())
    })

    // Subscribe to connection status
    const unsubscribeConnection = subscribe('connect', () => {
      setConnectionStatus('connected')
    })

    const unsubscribeDisconnect = subscribe('disconnect', () => {
      setConnectionStatus('disconnected')
    })

    const unsubscribeReconnecting = subscribe('reconnecting', () => {
      setConnectionStatus('reconnecting')
    })

    // Cleanup subscriptions
    return () => {
      unsubscribeExperiment()
      unsubscribeMetrics()
      unsubscribeConnection()
      unsubscribeDisconnect()
      unsubscribeReconnecting()
    }
  }, [experimentId, autoRefresh, subscribe, experiment, dispatch])

  const getStatusColor = (status: ExperimentStatus) => {
    switch (status) {
      case 'pending':
        return 'warning'
      case 'initializing':
        return 'info'
      case 'running':
        return 'success'
      case 'analyzing':
        return 'info'
      case 'completed':
        return 'success'
      case 'failed':
        return 'error'
      case 'cancelled':
        return 'default'
      default:
        return 'default'
    }
  }

  const getStatusIcon = (status: ExperimentStatus) => {
    switch (status) {
      case 'running':
        return <PlayArrow />
      case 'completed':
        return <CheckCircle />
      case 'failed':
        return <ErrorIcon />
      case 'cancelled':
        return <Stop />
      default:
        return <Pause />
    }
  }

  const formatDuration = (start: string, end?: string) => {
    const startTime = new Date(start)
    const endTime = end ? new Date(end) : new Date()
    const duration = Math.floor((endTime.getTime() - startTime.getTime()) / 1000)
    
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    const seconds = duration % 60
    
    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`
    } else {
      return `${seconds}s`
    }
  }

  if (!experiment) {
    return (
      <Alert severity="error">
        Experiment {experimentId} not found
      </Alert>
    )
  }

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Typography variant="h6">
              Real-time Status
            </Typography>
            <Chip
              icon={getStatusIcon(experiment.status)}
              label={experiment.status.toUpperCase()}
              color={getStatusColor(experiment.status) as any}
              size="small"
            />
            <Chip
              label={connectionStatus.toUpperCase()}
              color={connectionStatus === 'connected' ? 'success' : connectionStatus === 'reconnecting' ? 'warning' : 'error'}
              size="small"
              variant="outlined"
            />
          </Box>
          
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {lastUpdate && (
              <Typography variant="caption" color="text.secondary">
                Updated: {lastUpdate.toLocaleTimeString()}
              </Typography>
            )}
            <Tooltip title={expanded ? "Collapse" : "Expand"}>
              <IconButton size="small" onClick={() => setExpanded(!expanded)}>
                {expanded ? <ExpandLess /> : <ExpandMore />}
              </IconButton>
            </Tooltip>
          </Box>
        </Box>

        {/* Progress Bar */}
        {experiment.status === 'running' && experiment.progress !== undefined && (
          <Box sx={{ mb: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
              <Typography variant="body2">Progress</Typography>
              <Typography variant="body2">{Math.round(experiment.progress)}%</Typography>
            </Box>
            <LinearProgress 
              variant="determinate" 
              value={experiment.progress} 
              sx={{ height: 8, borderRadius: 1 }}
            />
          </Box>
        )}

        {/* Duration */}
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Duration: {formatDuration(experiment.created_at, experiment.completed_at)}
        </Typography>

        {/* Real-time Metrics */}
        {metrics && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              Live Metrics
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={6} sm={3}>
                <Box sx={{ textAlign: 'center', p: 1, bgcolor: 'background.paper', borderRadius: 1 }}>
                  <Typography variant="h6" color="primary">
                    {metrics.metrics.throughput.toFixed(1)}
                  </Typography>
                  <Typography variant="caption">
                    Throughput (req/s)
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Box sx={{ textAlign: 'center', p: 1, bgcolor: 'background.paper', borderRadius: 1 }}>
                  <Typography variant="h6" color="primary">
                    {metrics.metrics.latency.toFixed(0)}ms
                  </Typography>
                  <Typography variant="caption">
                    Latency
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Box sx={{ textAlign: 'center', p: 1, bgcolor: 'background.paper', borderRadius: 1 }}>
                  <Typography variant="h6" color={metrics.metrics.errorRate > 5 ? 'error' : 'primary'}>
                    {metrics.metrics.errorRate.toFixed(2)}%
                  </Typography>
                  <Typography variant="caption">
                    Error Rate
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={6} sm={3}>
                <Box sx={{ textAlign: 'center', p: 1, bgcolor: 'background.paper', borderRadius: 1 }}>
                  <Typography variant="h6" color="success.main">
                    -{metrics.metrics.reductionPercentage.toFixed(1)}%
                  </Typography>
                  <Typography variant="caption">
                    Volume Reduction
                  </Typography>
                </Box>
              </Grid>
            </Grid>
          </Box>
        )}

        {/* Detailed Information */}
        <Collapse in={expanded}>
          <Box sx={{ pt: 2, borderTop: 1, borderColor: 'divider' }}>
            <Typography variant="subtitle2" gutterBottom>
              Experiment Details
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  ID: {experiment.id}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  Owner: {experiment.owner}
                </Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary">
                  Description: {experiment.description || 'No description'}
                </Typography>
              </Grid>
              {experiment.result?.message && (
                <Grid item xs={12}>
                  <Typography variant="body2" color="text.secondary">
                    Message: {experiment.result.message}
                  </Typography>
                </Grid>
              )}
            </Grid>

            {/* WebSocket Connection Info */}
            <Box sx={{ mt: 2, p: 2, bgcolor: 'background.default', borderRadius: 1 }}>
              <Typography variant="subtitle2" gutterBottom>
                Real-time Connection
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Status: {connectionStatus}
              </Typography>
              {lastUpdate && (
                <Typography variant="body2" color="text.secondary">
                  Last Update: {lastUpdate.toLocaleString()}
                </Typography>
              )}
            </Box>
          </Box>
        </Collapse>
      </CardContent>
    </Card>
  )
}