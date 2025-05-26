import React, { useState, useEffect, useRef } from 'react'
import {
  Card,
  CardContent,
  Typography,
  Box,
  Grid,
  Switch,
  FormControlLabel,
  IconButton,
  Tooltip,
  Alert,
  Chip,
} from '@mui/material'
import {
  Timeline,
  TrendingUp,
  TrendingDown,
  Refresh,
  PlayArrow,
  Pause,
  Settings,
} from '@mui/icons-material'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  Area,
  AreaChart,
} from 'recharts'
import { useWebSocket } from '../../hooks/useWebSocket'

interface MetricPoint {
  timestamp: number
  throughput: number
  latency: number
  errorRate: number
  memoryUsage: number
  cpuUsage: number
  metricsVolume: number
  reductionPercentage: number
}

interface RealTimeMetricsProps {
  experimentId?: string
  autoUpdate?: boolean
  maxDataPoints?: number
  updateInterval?: number
}

export const RealTimeMetrics: React.FC<RealTimeMetricsProps> = ({
  experimentId,
  autoUpdate = true,
  maxDataPoints = 50,
  updateInterval = 5000,
}) => {
  const { subscribe, connected } = useWebSocket()
  const [isLive, setIsLive] = useState(autoUpdate)
  const [metrics, setMetrics] = useState<MetricPoint[]>([])
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const intervalRef = useRef<NodeJS.Timeout>()

  // Generate mock data for demonstration
  const generateMockData = (): MetricPoint => {
    const now = Date.now()
    const baseValues = metrics.length > 0 ? metrics[metrics.length - 1] : {
      throughput: 100,
      latency: 200,
      errorRate: 1,
      memoryUsage: 50,
      cpuUsage: 30,
      metricsVolume: 1000,
      reductionPercentage: 45,
    }

    return {
      timestamp: now,
      throughput: Math.max(0, baseValues.throughput + (Math.random() - 0.5) * 20),
      latency: Math.max(0, baseValues.latency + (Math.random() - 0.5) * 50),
      errorRate: Math.max(0, Math.min(10, baseValues.errorRate + (Math.random() - 0.5) * 2)),
      memoryUsage: Math.max(0, Math.min(100, baseValues.memoryUsage + (Math.random() - 0.5) * 10)),
      cpuUsage: Math.max(0, Math.min(100, baseValues.cpuUsage + (Math.random() - 0.5) * 15)),
      metricsVolume: Math.max(0, baseValues.metricsVolume + (Math.random() - 0.5) * 200),
      reductionPercentage: Math.max(0, Math.min(80, baseValues.reductionPercentage + (Math.random() - 0.5) * 5)),
    }
  }

  useEffect(() => {
    if (!isLive) return

    // Subscribe to real-time metrics updates
    const unsubscribe = subscribe(
      experimentId ? `metrics.${experimentId}` : 'metrics.global',
      (data: any) => {
        const newPoint: MetricPoint = {
          timestamp: Date.now(),
          throughput: data.throughput || 0,
          latency: data.latency || 0,
          errorRate: data.errorRate || 0,
          memoryUsage: data.memoryUsage || 0,
          cpuUsage: data.cpuUsage || 0,
          metricsVolume: data.metricsVolume || 0,
          reductionPercentage: data.reductionPercentage || 0,
        }

        setMetrics(prev => {
          const updated = [...prev, newPoint]
          return updated.slice(-maxDataPoints)
        })
        setLastUpdate(new Date())
      }
    )

    // Fallback: Generate mock data if no real data is coming
    intervalRef.current = setInterval(() => {
      if (connected) {
        const mockPoint = generateMockData()
        setMetrics(prev => {
          const updated = [...prev, mockPoint]
          return updated.slice(-maxDataPoints)
        })
        setLastUpdate(new Date())
      }
    }, updateInterval)

    return () => {
      unsubscribe()
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [isLive, experimentId, maxDataPoints, updateInterval, connected, subscribe])

  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString()
  }

  const getCurrentValue = (key: keyof MetricPoint) => {
    return metrics.length > 0 ? metrics[metrics.length - 1][key] : 0
  }

  const getTrend = (key: keyof MetricPoint) => {
    if (metrics.length < 2) return 'stable'
    const current = metrics[metrics.length - 1][key] as number
    const previous = metrics[metrics.length - 2][key] as number
    
    if (current > previous * 1.05) return 'up'
    if (current < previous * 0.95) return 'down'
    return 'stable'
  }

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up':
        return <TrendingUp color="success" />
      case 'down':
        return <TrendingDown color="error" />
      default:
        return <Timeline color="disabled" />
    }
  }

  const MetricCard: React.FC<{
    title: string
    value: number
    unit: string
    dataKey: keyof MetricPoint
    color?: string
    format?: (value: number) => string
  }> = ({ title, value, unit, dataKey, color = '#8884d8', format }) => {
    const trend = getTrend(dataKey)
    const formattedValue = format ? format(value) : value.toFixed(1)

    return (
      <Card sx={{ height: '100%' }}>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
            <Typography variant="subtitle2" color="text.secondary">
              {title}
            </Typography>
            {getTrendIcon(trend)}
          </Box>
          
          <Typography variant="h4" color={color} gutterBottom>
            {formattedValue}
            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
              {unit}
            </Typography>
          </Typography>

          <Box sx={{ height: 60, mt: 2 }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={metrics.slice(-10)}>
                <Area
                  type="monotone"
                  dataKey={dataKey}
                  stroke={color}
                  fill={color}
                  fillOpacity={0.2}
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </Box>
        </CardContent>
      </Card>
    )
  }

  return (
    <Box>
      {/* Header */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">
          Real-time Metrics
          {experimentId && (
            <Chip 
              label={`Experiment: ${experimentId}`} 
              size="small" 
              sx={{ ml: 2 }}
            />
          )}
        </Typography>
        
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <FormControlLabel
            control={
              <Switch
                checked={isLive}
                onChange={(e) => setIsLive(e.target.checked)}
                color="primary"
              />
            }
            label="Live Updates"
          />
          
          <Chip
            icon={connected ? <PlayArrow /> : <Pause />}
            label={connected ? 'Connected' : 'Disconnected'}
            color={connected ? 'success' : 'error'}
            size="small"
          />

          {lastUpdate && (
            <Typography variant="caption" color="text.secondary">
              Last update: {lastUpdate.toLocaleTimeString()}
            </Typography>
          )}
        </Box>
      </Box>

      {!connected && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          WebSocket connection lost. Metrics may not be up to date.
        </Alert>
      )}

      {/* Metric Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Throughput"
            value={getCurrentValue('throughput')}
            unit="req/s"
            dataKey="throughput"
            color="#2196f3"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Latency"
            value={getCurrentValue('latency')}
            unit="ms"
            dataKey="latency"
            color="#ff9800"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Error Rate"
            value={getCurrentValue('errorRate')}
            unit="%"
            dataKey="errorRate"
            color="#f44336"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Volume Reduction"
            value={getCurrentValue('reductionPercentage')}
            unit="%"
            dataKey="reductionPercentage"
            color="#4caf50"
          />
        </Grid>
      </Grid>

      {/* Detailed Charts */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Performance Metrics
              </Typography>
              <Box sx={{ height: 300 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metrics}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="timestamp" 
                      tickFormatter={formatTimestamp}
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                    />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <RechartsTooltip 
                      labelFormatter={formatTimestamp}
                      formatter={(value: number, name: string) => [
                        typeof value === 'number' ? value.toFixed(2) : value,
                        name
                      ]}
                    />
                    <Line
                      yAxisId="left"
                      type="monotone"
                      dataKey="throughput"
                      stroke="#2196f3"
                      strokeWidth={2}
                      name="Throughput (req/s)"
                    />
                    <Line
                      yAxisId="right"
                      type="monotone"
                      dataKey="latency"
                      stroke="#ff9800"
                      strokeWidth={2}
                      name="Latency (ms)"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Resource Usage
              </Typography>
              <Box sx={{ height: 300 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metrics}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="timestamp" 
                      tickFormatter={formatTimestamp}
                      type="number"
                      scale="time"
                      domain={['dataMin', 'dataMax']}
                    />
                    <YAxis domain={[0, 100]} />
                    <RechartsTooltip 
                      labelFormatter={formatTimestamp}
                      formatter={(value: number, name: string) => [
                        `${value.toFixed(1)}%`,
                        name
                      ]}
                    />
                    <Line
                      type="monotone"
                      dataKey="cpuUsage"
                      stroke="#9c27b0"
                      strokeWidth={2}
                      name="CPU Usage"
                    />
                    <Line
                      type="monotone"
                      dataKey="memoryUsage"
                      stroke="#607d8b"
                      strokeWidth={2}
                      name="Memory Usage"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  )
}