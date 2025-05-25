import React from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Typography,
  Box,
  Chip,
} from '@mui/material'
import { TrendingUp, TrendingDown, Remove } from '@mui/icons-material'
import { MetricsSummary } from '../../types'

interface ComparisonTableProps {
  baseline: MetricsSummary
  candidate: MetricsSummary
}

export const ComparisonTable: React.FC<ComparisonTableProps> = ({
  baseline,
  candidate,
}) => {
  const calculateChange = (baselineValue: number, candidateValue: number) => {
    const change = ((candidateValue - baselineValue) / baselineValue) * 100
    return change
  }

  const formatValue = (value: number, metric: string) => {
    switch (metric) {
      case 'cardinality':
      case 'dataPointsPerSecond':
        return value.toLocaleString()
      case 'cpuUsage':
        return `${value.toFixed(1)}%`
      case 'memoryUsage':
        return `${value.toFixed(0)} MB`
      case 'networkTraffic':
        return `${value.toFixed(0)} KB/s`
      case 'uniqueProcesses':
        return value.toLocaleString()
      default:
        return value.toFixed(2)
    }
  }

  const getChangeIcon = (change: number) => {
    if (Math.abs(change) < 1) return <Remove fontSize="small" />
    if (change > 0) return <TrendingUp fontSize="small" />
    return <TrendingDown fontSize="small" />
  }

  const getChangeColor = (change: number, metric: string) => {
    // For cost-related metrics, reduction is good
    if (['cardinality', 'networkTraffic', 'dataPointsPerSecond'].includes(metric)) {
      return change < 0 ? 'success' : 'error'
    }
    // For performance metrics, increase is bad
    if (['cpuUsage', 'memoryUsage'].includes(metric)) {
      return change > 5 ? 'error' : change > 0 ? 'warning' : 'success'
    }
    // For process retention, reduction is bad
    if (metric === 'uniqueProcesses') {
      return change < -50 ? 'error' : change < 0 ? 'warning' : 'success'
    }
    return 'default'
  }

  const metrics = [
    { key: 'cardinality', label: 'Time Series Cardinality', unit: 'series' },
    { key: 'cpuUsage', label: 'CPU Usage', unit: '%' },
    { key: 'memoryUsage', label: 'Memory Usage', unit: 'MB' },
    { key: 'networkTraffic', label: 'Network Traffic', unit: 'KB/s' },
    { key: 'dataPointsPerSecond', label: 'Data Points/Second', unit: 'points/s' },
    { key: 'uniqueProcesses', label: 'Unique Processes', unit: 'processes' },
  ]

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Metric</TableCell>
            <TableCell align="right">Baseline</TableCell>
            <TableCell align="right">Candidate</TableCell>
            <TableCell align="right">Change</TableCell>
            <TableCell align="center">Impact</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {metrics.map(({ key, label }) => {
            const baselineValue = baseline[key as keyof MetricsSummary]
            const candidateValue = candidate[key as keyof MetricsSummary]
            const change = calculateChange(baselineValue, candidateValue)

            return (
              <TableRow key={key}>
                <TableCell>
                  <Typography variant="body2">{label}</Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2">
                    {formatValue(baselineValue, key)}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2">
                    {formatValue(candidateValue, key)}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
                    {getChangeIcon(change)}
                    <Typography
                      variant="body2"
                      sx={{ ml: 0.5 }}
                      color={
                        Math.abs(change) < 1
                          ? 'text.secondary'
                          : change > 0
                          ? 'error.main'
                          : 'success.main'
                      }
                    >
                      {change > 0 && '+'}
                      {change.toFixed(1)}%
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="center">
                  <Chip
                    label={
                      Math.abs(change) < 1
                        ? 'No Change'
                        : Math.abs(change) < 10
                        ? 'Low'
                        : Math.abs(change) < 50
                        ? 'Medium'
                        : 'High'
                    }
                    size="small"
                    color={getChangeColor(change, key) as any}
                    variant="outlined"
                  />
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}