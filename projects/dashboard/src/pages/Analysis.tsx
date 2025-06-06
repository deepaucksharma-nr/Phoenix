import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Container,
  Box,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material'
import {
  ArrowBack,
} from '@mui/icons-material'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { fetchExperimentById, promoteVariant } from '@store/slices/experimentSlice'
import { EnhancedAnalysis } from '../components/Analysis'
import { useExperimentUpdates } from '../hooks/useExperimentUpdates'

export const Analysis: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { currentExperiment: experiment, loading, error } = useAppSelector(
    (state) => state.experiments
  )

  const [metrics, setMetrics] = useState<any[]>([])

  // Subscribe to real-time updates
  const { metrics: realtimeMetrics } = useExperimentUpdates(id)

  useEffect(() => {
    if (id) {
      dispatch(fetchExperimentById(id))
    }
  }, [id, dispatch])

  useEffect(() => {
    if (realtimeMetrics) {
      setMetrics(realtimeMetrics)
    }
  }, [realtimeMetrics])

  const handleBack = () => {
    navigate(`/experiments/${id}`)
  }

  const handlePromoteVariant = async (variant: 'baseline' | 'candidate') => {
    if (id) {
      await dispatch(promoteVariant({ id, variant }))
      navigate(`/experiments/${id}`)
    }
  }

  if (loading && !experiment) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      </Container>
    )
  }

  if (error) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4 }}>
        <Alert severity="error">{error}</Alert>
        <Button onClick={handleBack} sx={{ mt: 2 }}>
          Back to Experiment
        </Button>
      </Container>
    )
  }

  if (!experiment || !id) {
    return null
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button startIcon={<ArrowBack />} onClick={handleBack}>
          Back to Experiment
        </Button>
      </Box>

      <EnhancedAnalysis
        experimentId={id}
        experimentData={experiment}
        metricsData={metrics}
        onPromote={handlePromoteVariant}
      />
    </Container>
  )
}