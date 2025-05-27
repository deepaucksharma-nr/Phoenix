import React from 'react'
import { Box, Typography, Button, SvgIcon } from '@mui/material'
import { Inbox, SearchOff, ErrorOutline } from '@mui/icons-material'

interface EmptyStateProps {
  icon?: React.ReactElement
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
  type?: 'empty' | 'search' | 'error'
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
  type = 'empty',
}) => {
  const defaultIcons = {
    empty: <Inbox />,
    search: <SearchOff />,
    error: <ErrorOutline />,
  }

  const displayIcon = icon || defaultIcons[type]

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        p: 4,
        minHeight: 300,
        textAlign: 'center',
      }}
    >
      <SvgIcon
        component={() => displayIcon}
        sx={{
          fontSize: 64,
          color: 'text.secondary',
          mb: 2,
          opacity: 0.5,
        }}
      />
      <Typography variant="h6" gutterBottom>
        {title}
      </Typography>
      {description && (
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ mb: 3, maxWidth: 400 }}
        >
          {description}
        </Typography>
      )}
      {action && (
        <Button variant="contained" onClick={action.onClick}>
          {action.label}
        </Button>
      )}
    </Box>
  )
}