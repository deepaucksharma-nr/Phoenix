import React from 'react'
import { Card, CardContent, Typography, Box } from '@mui/material'
import { TrendingUp, TrendingDown } from '@mui/icons-material'

interface MetricCardProps {
  title: string
  value: string | number
  change?: number
  icon?: React.ReactNode
  color?: 'primary' | 'secondary' | 'success' | 'error' | 'warning' | 'info'
  subtitle?: string
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  change,
  icon,
  color = 'primary',
  subtitle,
}) => {
  const isPositive = change !== undefined && change > 0
  const isNegative = change !== undefined && change < 0

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          {icon && (
            <Box
              sx={{
                p: 1,
                borderRadius: 1,
                bgcolor: `${color}.light`,
                color: `${color}.main`,
                mr: 2,
              }}
            >
              {icon}
            </Box>
          )}
          <Typography color="text.secondary" variant="subtitle2" sx={{ flexGrow: 1 }}>
            {title}
          </Typography>
        </Box>
        
        <Typography variant="h4" component="div" sx={{ mb: 1 }}>
          {value}
        </Typography>
        
        {change !== undefined && (
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            {isPositive && <TrendingUp color="success" sx={{ mr: 0.5 }} />}
            {isNegative && <TrendingDown color="error" sx={{ mr: 0.5 }} />}
            <Typography
              variant="body2"
              color={isPositive ? 'success.main' : isNegative ? 'error.main' : 'text.secondary'}
            >
              {isPositive && '+'}
              {change.toFixed(1)}%
            </Typography>
          </Box>
        )}
        
        {subtitle && (
          <Typography variant="caption" color="text.secondary">
            {subtitle}
          </Typography>
        )}
      </CardContent>
    </Card>
  )
}