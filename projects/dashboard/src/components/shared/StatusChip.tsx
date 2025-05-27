import React from 'react'
import { Chip, ChipProps } from '@mui/material'
import {
  CheckCircle,
  Cancel,
  Schedule,
  PlayArrow,
  Pause,
  Warning,
  Error as ErrorIcon,
} from '@mui/icons-material'

export type Status = 
  | 'active'
  | 'inactive'
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'paused'
  | 'degraded'
  | 'healthy'
  | 'warning'
  | 'error'
  | 'success'

interface StatusChipProps extends Omit<ChipProps, 'color' | 'icon'> {
  status: Status
  showIcon?: boolean
}

const statusConfig: Record<Status, {
  color: ChipProps['color']
  icon: React.ReactElement
  label?: string
}> = {
  active: {
    color: 'success',
    icon: <CheckCircle />,
    label: 'Active',
  },
  inactive: {
    color: 'default',
    icon: <Cancel />,
    label: 'Inactive',
  },
  pending: {
    color: 'warning',
    icon: <Schedule />,
    label: 'Pending',
  },
  running: {
    color: 'info',
    icon: <PlayArrow />,
    label: 'Running',
  },
  completed: {
    color: 'success',
    icon: <CheckCircle />,
    label: 'Completed',
  },
  failed: {
    color: 'error',
    icon: <Cancel />,
    label: 'Failed',
  },
  paused: {
    color: 'default',
    icon: <Pause />,
    label: 'Paused',
  },
  degraded: {
    color: 'warning',
    icon: <Warning />,
    label: 'Degraded',
  },
  healthy: {
    color: 'success',
    icon: <CheckCircle />,
    label: 'Healthy',
  },
  warning: {
    color: 'warning',
    icon: <Warning />,
    label: 'Warning',
  },
  error: {
    color: 'error',
    icon: <ErrorIcon />,
    label: 'Error',
  },
  success: {
    color: 'success',
    icon: <CheckCircle />,
    label: 'Success',
  },
}

export const StatusChip: React.FC<StatusChipProps> = ({
  status,
  showIcon = true,
  label,
  size = 'small',
  ...props
}) => {
  const config = statusConfig[status] || statusConfig.inactive
  
  return (
    <Chip
      color={config.color}
      icon={showIcon ? config.icon : undefined}
      label={label || config.label || status}
      size={size}
      {...props}
    />
  )
}