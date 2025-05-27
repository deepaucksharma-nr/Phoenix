import React from 'react'
import {
  Box,
  CircularProgress,
  LinearProgress,
  Typography,
  Skeleton,
  Stack,
} from '@mui/material'

interface LoadingStateProps {
  type?: 'circular' | 'linear' | 'skeleton'
  message?: string
  fullHeight?: boolean
  skeletonRows?: number
  skeletonHeight?: number
}

export const LoadingState: React.FC<LoadingStateProps> = ({
  type = 'circular',
  message = 'Loading...',
  fullHeight = false,
  skeletonRows = 5,
  skeletonHeight = 60,
}) => {
  const containerProps = {
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 2,
    p: 4,
    ...(fullHeight && { minHeight: '60vh' }),
  }

  if (type === 'skeleton') {
    return (
      <Stack spacing={2} sx={{ p: 2 }}>
        {Array.from({ length: skeletonRows }).map((_, index) => (
          <Skeleton
            key={index}
            variant="rectangular"
            height={skeletonHeight}
            animation="wave"
          />
        ))}
      </Stack>
    )
  }

  if (type === 'linear') {
    return (
      <Box sx={{ width: '100%', p: 2 }}>
        <LinearProgress />
        {message && (
          <Typography
            variant="body2"
            color="text.secondary"
            align="center"
            sx={{ mt: 2 }}
          >
            {message}
          </Typography>
        )}
      </Box>
    )
  }

  return (
    <Box sx={containerProps}>
      <CircularProgress size={40} />
      {message && (
        <Typography variant="body2" color="text.secondary">
          {message}
        </Typography>
      )}
    </Box>
  )
}