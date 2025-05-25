import React from 'react'
import { Box, Paper, Typography } from '@mui/material'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { format } from 'date-fns'

interface MetricsChartProps {
  title: string
  data: {
    baseline: Array<{ timestamp: number; value: number }>
    candidate: Array<{ timestamp: number; value: number }>
  }
  yAxisLabel?: string
  height?: number
}

export const MetricsChart: React.FC<MetricsChartProps> = ({
  title,
  data,
  yAxisLabel = 'Value',
  height = 300,
}) => {
  // Merge and format data for recharts
  const chartData = data.baseline.map((point, index) => ({
    timestamp: point.timestamp,
    time: format(new Date(point.timestamp), 'HH:mm'),
    baseline: point.value,
    candidate: data.candidate[index]?.value || 0,
  }))

  const formatYAxis = (value: number) => {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    } else if (value >= 1000) {
      return `${(value / 1000).toFixed(1)}K`
    }
    return value.toFixed(0)
  }

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        {title}
      </Typography>
      <ResponsiveContainer width="100%" height={height}>
        <LineChart
          data={chartData}
          margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
          <XAxis
            dataKey="time"
            stroke="#666"
            style={{ fontSize: 12 }}
          />
          <YAxis
            label={{
              value: yAxisLabel,
              angle: -90,
              position: 'insideLeft',
              style: { fontSize: 12, fill: '#666' },
            }}
            tickFormatter={formatYAxis}
            stroke="#666"
            style={{ fontSize: 12 }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(255, 255, 255, 0.95)',
              border: '1px solid #ccc',
              borderRadius: 4,
            }}
            formatter={(value: number) => [formatYAxis(value), '']}
            labelFormatter={(label) => `Time: ${label}`}
          />
          <Legend
            wrapperStyle={{ fontSize: 12 }}
            iconType="line"
          />
          <Line
            type="monotone"
            dataKey="baseline"
            stroke="#1976d2"
            strokeWidth={2}
            dot={false}
            name="Baseline"
          />
          <Line
            type="monotone"
            dataKey="candidate"
            stroke="#f50057"
            strokeWidth={2}
            dot={false}
            name="Candidate"
          />
        </LineChart>
      </ResponsiveContainer>
    </Paper>
  )
}