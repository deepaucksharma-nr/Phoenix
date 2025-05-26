import React, { useEffect, useState, useRef } from 'react'
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  LinearProgress,
  Chip,
  Alert,
  Button,
  IconButton,
  Tooltip,
  Badge,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Skeleton,
  CircularProgress,
  Fade,
  Collapse,
  SpeedDial,
  SpeedDialAction,
  SpeedDialIcon,
  Dialog,
  DialogTitle,
  DialogContent,
} from '@mui/material'
import {
  PlayArrow,
  Stop,
  Refresh,
  Timeline,
  Assessment,
  Warning,
  CheckCircle,
  Error as ErrorIcon,
  Schedule,
  TrendingDown,
  Memory,
  Storage,
  NetworkCheck,
  Speed,
  AttachMoney,
  AutorenewOutlined,
  ExpandMore,
  ExpandLess,
  Fullscreen,
  FullscreenExit,
  Download,
  Share,
  Pause,
  Close,
  Science,
} from '@mui/icons-material'
import { formatDistanceToNow, format } from 'date-fns'
import { Line, Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler,
} from 'chart.js'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { fetchExperimentById } from '@store/slices/experimentSlice'
import { useExperimentUpdates } from '../../hooks/useExperimentUpdates'
import { Experiment, ExperimentMetrics } from '../../types/experiment'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  ChartTooltip,
  Legend,
  Filler
)

interface ExperimentMonitorProps {
  experimentId: string
  embedded?: boolean
  onClose?: () => void
}

interface MetricValue {
  current: number
  baseline: number
  reduction: number
  trend: 'up' | 'down' | 'stable'
}

interface LiveMetrics {
  cardinality: MetricValue
  dataRate: MetricValue
  errorRate: MetricValue
  latency: MetricValue
  costSavings: number
  timestamp: number
}

const MetricCard: React.FC<{
  title: string
  icon: React.ReactNode
  current: number
  baseline: number
  reduction: number
  trend: 'up' | 'down' | 'stable'
  unit?: string
  color?: string
}> = ({ title, icon, current, baseline, reduction, trend, unit = '', color = 'primary' }) => {
  const getTrendIcon = () => {
    if (trend === 'up') return '↑'
    if (trend === 'down') return '↓'
    return '→'
  }

  const getTrendColor = () => {
    if (title.includes('Error') || title.includes('Latency')) {
      return trend === 'up' ? 'error.main' : 'success.main'
    }
    return trend === 'down' ? 'success.main' : 'error.main'
  }

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          <Box sx={{ color: `${color}.main`, mr: 1 }}>{icon}</Box>
          <Typography variant="subtitle2" color="text.secondary">
            {title}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'baseline', mb: 1 }}>
          <Typography variant="h5" component="span">
            {current.toLocaleString()}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ ml: 0.5 }}>
            {unit}
          </Typography>
          <Typography
            variant="caption"
            sx={{ ml: 1, color: getTrendColor() }}
          >
            {getTrendIcon()}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Typography variant="caption" color="text.secondary">
            Baseline: {baseline.toLocaleString()}{unit}
          </Typography>
          <Chip
            label={`${reduction > 0 ? '-' : '+'} ${Math.abs(reduction)}%`}
            size="small"
            color={reduction > 0 ? 'success' : 'error'}
            sx={{ height: 20 }}
          />
        </Box>
        <LinearProgress
          variant="determinate"
          value={Math.min(100, Math.abs(reduction))}
          color={reduction > 0 ? 'success' : 'error'}
          sx={{ mt: 1, height: 4, borderRadius: 2 }}
        />
      </CardContent>
    </Card>
  )
}

export const ExperimentMonitor: React.FC<ExperimentMonitorProps> = ({
  experimentId,
  embedded = false,
  onClose,
}) => {
  const dispatch = useAppDispatch()
  const experiments = useAppSelector((state) => state.experiments.experiments)
  const { metrics, events, status } = useExperimentUpdates(experimentId)
  const [experiment, setExperiment] = useState<Experiment | null>(null)
  const [liveMetrics, setLiveMetrics] = useState<LiveMetrics>({
    cardinality: { current: 150000, baseline: 150000, reduction: 0, trend: 'stable' },
    dataRate: { current: 2.3, baseline: 2.3, reduction: 0, trend: 'stable' },
    errorRate: { current: 0.05, baseline: 0.05, reduction: 0, trend: 'stable' },
    latency: { current: 45, baseline: 45, reduction: 0, trend: 'stable' },
    costSavings: 0,
    timestamp: Date.now(),
  })
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [fullscreen, setFullscreen] = useState(false)
  const [detailsExpanded, setDetailsExpanded] = useState(true)
  const [speedDialOpen, setSpeedDialOpen] = useState(false)
  const chartRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const exp = experiments.find(e => e.id === experimentId)
    if (exp) {
      setExperiment(exp)
    } else {
      dispatch(fetchExperimentById(experimentId)).then((action) => {
        if (fetchExperimentById.fulfilled.match(action)) {
          setExperiment(action.payload)
        }
      })
    }
  }, [experimentId, experiments, dispatch])

  useEffect(() => {
    if (!metrics) return

    // Simulate live metrics updates based on WebSocket data
    const latestMetric = metrics[metrics.length - 1]
    if (latestMetric) {
      setLiveMetrics(prev => {
        const newCardinality = latestMetric.candidate_cardinality || prev.cardinality.current
        const cardinalityReduction = ((prev.cardinality.baseline - newCardinality) / prev.cardinality.baseline) * 100
        
        const newDataRate = latestMetric.candidate_data_rate || prev.dataRate.current
        const dataRateReduction = ((prev.dataRate.baseline - newDataRate) / prev.dataRate.baseline) * 100
        
        return {
          cardinality: {
            current: newCardinality,
            baseline: prev.cardinality.baseline,
            reduction: cardinalityReduction,
            trend: newCardinality < prev.cardinality.current ? 'down' : 'up',
          },
          dataRate: {
            current: newDataRate,
            baseline: prev.dataRate.baseline,
            reduction: dataRateReduction,
            trend: newDataRate < prev.dataRate.current ? 'down' : 'up',
          },
          errorRate: {
            current: latestMetric.error_rate || prev.errorRate.current,
            baseline: prev.errorRate.baseline,
            reduction: ((prev.errorRate.baseline - (latestMetric.error_rate || prev.errorRate.current)) / prev.errorRate.baseline) * 100,
            trend: (latestMetric.error_rate || 0) > prev.errorRate.current ? 'up' : 'down',
          },
          latency: {
            current: latestMetric.p95_latency || prev.latency.current,
            baseline: prev.latency.baseline,
            reduction: ((prev.latency.baseline - (latestMetric.p95_latency || prev.latency.current)) / prev.latency.baseline) * 100,
            trend: (latestMetric.p95_latency || 0) > prev.latency.current ? 'up' : 'down',
          },
          costSavings: cardinalityReduction * 2500, // $2500 per 1% reduction
          timestamp: Date.now(),
        }
      })
    }
  }, [metrics])

  const getStatusIcon = () => {
    switch (experiment?.status) {
      case 'running':
        return <AutorenewOutlined className="rotating" />
      case 'completed':
        return <CheckCircle color="success" />
      case 'failed':
        return <ErrorIcon color="error" />
      case 'pending':
        return <Schedule />
      default:
        return <Science />
    }
  }

  const getStatusColor = () => {
    switch (experiment?.status) {
      case 'running':
        return 'primary'
      case 'completed':
        return 'success'
      case 'failed':
        return 'error'
      default:
        return 'default'
    }
  }

  const chartData = {
    labels: metrics?.map(m => format(new Date(m.timestamp), 'HH:mm')) || [],
    datasets: [
      {
        label: 'Baseline Cardinality',
        data: metrics?.map(m => m.baseline_cardinality) || [],
        borderColor: 'rgb(255, 99, 132)',
        backgroundColor: 'rgba(255, 99, 132, 0.1)',
        tension: 0.4,
      },
      {
        label: 'Candidate Cardinality',
        data: metrics?.map(m => m.candidate_cardinality) || [],
        borderColor: 'rgb(54, 162, 235)',
        backgroundColor: 'rgba(54, 162, 235, 0.1)',
        tension: 0.4,
      },
    ],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: false,
      },
    },
    scales: {
      y: {
        beginAtZero: true,
      },
    },
  }

  const speedDialActions = [
    { icon: <Download />, name: 'Export Report', action: () => {} },
    { icon: <Share />, name: 'Share Results', action: () => {} },
    { icon: <Assessment />, name: 'View Analysis', action: () => {} },
    { icon: <Stop />, name: 'Stop Experiment', action: () => {} },
  ]

  if (!experiment) {
    return (
      <Box sx={{ p: 3 }}>
        <Skeleton variant="rectangular" height={200} />
      </Box>
    )
  }

  const content = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <Paper sx={{ p: 2, mb: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            {getStatusIcon()}
            <Box sx={{ ml: 2 }}>
              <Typography variant="h6">{experiment.name}</Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                <Chip
                  label={experiment.status}
                  size="small"
                  color={getStatusColor()}
                />
                <Typography variant="caption" color="text.secondary">
                  Started {formatDistanceToNow(new Date(experiment.createdAt))} ago
                </Typography>
                {experiment.status === 'running' && (
                  <Chip
                    icon={<AutorenewOutlined />}
                    label="Live"
                    size="small"
                    color="error"
                    sx={{ animation: 'pulse 2s infinite' }}
                  />
                )}
              </Box>
            </Box>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Tooltip title={autoRefresh ? 'Disable auto-refresh' : 'Enable auto-refresh'}>
              <IconButton onClick={() => setAutoRefresh(!autoRefresh)}>
                {autoRefresh ? <Pause /> : <PlayArrow />}
              </IconButton>
            </Tooltip>
            <Tooltip title="Refresh now">
              <IconButton onClick={() => dispatch(fetchExperimentById(experimentId))}>
                <Refresh />
              </IconButton>
            </Tooltip>
            {!embedded && (
              <Tooltip title={fullscreen ? 'Exit fullscreen' : 'Fullscreen'}>
                <IconButton onClick={() => setFullscreen(!fullscreen)}>
                  {fullscreen ? <FullscreenExit /> : <Fullscreen />}
                </IconButton>
              </Tooltip>
            )}
          </Box>
        </Box>
      </Paper>

      {/* Key Metrics */}
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Time Series Cardinality"
            icon={<Timeline />}
            {...liveMetrics.cardinality}
            color="primary"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Data Rate"
            icon={<Speed />}
            {...liveMetrics.dataRate}
            unit="M/min"
            color="info"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Error Rate"
            icon={<Warning />}
            {...liveMetrics.errorRate}
            unit="%"
            color="warning"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="P95 Latency"
            icon={<Schedule />}
            {...liveMetrics.latency}
            unit="ms"
            color="secondary"
          />
        </Grid>
      </Grid>

      {/* Cost Savings Alert */}
      {liveMetrics.costSavings > 0 && (
        <Fade in>
          <Alert
            severity="success"
            icon={<AttachMoney />}
            sx={{ mb: 2 }}
          >
            <Typography variant="subtitle2">
              Projected Annual Savings: ${liveMetrics.costSavings.toLocaleString()}
            </Typography>
          </Alert>
        </Fade>
      )}

      {/* Charts */}
      <Paper sx={{ p: 2, mb: 2, flexGrow: 1 }}>
        <Typography variant="h6" gutterBottom>
          Cardinality Comparison
        </Typography>
        <Box ref={chartRef} sx={{ height: 300 }}>
          <Line data={chartData} options={chartOptions} />
        </Box>
      </Paper>

      {/* Experiment Details */}
      <Paper sx={{ p: 2 }}>
        <Box
          sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', cursor: 'pointer' }}
          onClick={() => setDetailsExpanded(!detailsExpanded)}
        >
          <Typography variant="h6">Experiment Details</Typography>
          <IconButton size="small">
            {detailsExpanded ? <ExpandLess /> : <ExpandMore />}
          </IconButton>
        </Box>
        <Collapse in={detailsExpanded}>
          <Divider sx={{ my: 1 }} />
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                Pipeline Configuration
              </Typography>
              <Box sx={{ mt: 1 }}>
                <Chip label={`Baseline: ${experiment.spec?.baselinePipeline}`} size="small" sx={{ mr: 1 }} />
                <Chip label={`Candidate: ${experiment.spec?.candidatePipeline}`} size="small" color="primary" />
              </Box>
            </Grid>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                Target Hosts
              </Typography>
              <Typography variant="body2">
                {experiment.spec?.targetHosts?.length || 0} hosts ({experiment.spec?.targetPercentage || 10}%)
              </Typography>
            </Grid>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                Duration
              </Typography>
              <Typography variant="body2">
                {experiment.spec?.duration || '1 hour'}
              </Typography>
            </Grid>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                Success Criteria
              </Typography>
              <Typography variant="body2">
                Min reduction: {experiment.spec?.successCriteria?.minReduction || 30}%
              </Typography>
            </Grid>
          </Grid>
        </Collapse>
      </Paper>

      {/* Speed Dial for Actions */}
      {!embedded && (
        <SpeedDial
          ariaLabel="Experiment actions"
          sx={{ position: 'fixed', bottom: 16, right: 16 }}
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
  )

  if (embedded) {
    return content
  }

  return (
    <Dialog
      open
      onClose={onClose}
      fullScreen={fullscreen}
      maxWidth="lg"
      fullWidth
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h5">Experiment Monitor</Typography>
          <IconButton onClick={onClose}>
            <Close />
          </IconButton>
        </Box>
      </DialogTitle>
      <DialogContent>{content}</DialogContent>
    </Dialog>
  )
}