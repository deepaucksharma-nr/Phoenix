import { useEffect } from 'react'
import { useWebSocket } from './useWebSocket'
import { useExperimentStore } from '../store/useExperimentStore'
import { Experiment } from '../types'

interface ExperimentUpdatePayload {
  experiment: Experiment
  event: 'status_changed' | 'metrics_updated' | 'analysis_complete'
}

export const useExperimentUpdates = (experimentId?: string) => {
  const { subscribe } = useWebSocket()
  const { updateExperiment, fetchExperiment } = useExperimentStore()

  useEffect(() => {
    if (!experimentId) return

    // Subscribe to specific experiment updates
    const unsubscribeExperiment = subscribe(
      `experiment:${experimentId}`,
      (payload: ExperimentUpdatePayload) => {
        if (payload.experiment) {
          updateExperiment(payload.experiment.id, payload.experiment)
        }

        // Handle specific events
        switch (payload.event) {
          case 'status_changed':
            console.log(`Experiment ${experimentId} status changed to ${payload.experiment.status}`)
            break
          case 'metrics_updated':
            console.log(`New metrics available for experiment ${experimentId}`)
            // Refetch to get latest metrics
            fetchExperiment(experimentId)
            break
          case 'analysis_complete':
            console.log(`Analysis complete for experiment ${experimentId}`)
            // Refetch to get analysis results
            fetchExperiment(experimentId)
            break
        }
      }
    )

    // Subscribe to global experiment updates
    const unsubscribeGlobal = subscribe(
      'experiments:update',
      (payload: { experiments: Experiment[] }) => {
        // Update multiple experiments at once
        payload.experiments.forEach((exp) => {
          updateExperiment(exp.id, exp)
        })
      }
    )

    return () => {
      unsubscribeExperiment()
      unsubscribeGlobal()
    }
  }, [experimentId, subscribe, updateExperiment, fetchExperiment])
}

// Hook for real-time metrics updates
export const useMetricsUpdates = (experimentId?: string) => {
  const { subscribe } = useWebSocket()

  useEffect(() => {
    if (!experimentId) return

    const unsubscribe = subscribe(
      `metrics:${experimentId}`,
      (payload: {
        variant: 'baseline' | 'candidate'
        metric: string
        timestamp: number
        value: number
      }) => {
        // Handle real-time metric update
        console.log('Real-time metric:', payload)
        // TODO: Update metrics visualization
      }
    )

    return unsubscribe
  }, [experimentId, subscribe])
}

// Hook for system notifications
export const useSystemNotifications = () => {
  const { subscribe } = useWebSocket()

  useEffect(() => {
    const unsubscribeAlerts = subscribe('system:alert', (payload: {
      level: 'info' | 'warning' | 'error'
      title: string
      message: string
    }) => {
      console.log('System alert:', payload)
      // Use global handler if available (set by NotificationProvider)
      const handler = (window as any).__handleSystemAlert
      if (handler) {
        handler(payload)
      }
    })

    const unsubscribeNotifications = subscribe('system:notification', (payload: {
      id: string
      type: string
      title: string
      message?: string
      timestamp: string
    }) => {
      console.log('System notification:', payload)
      // Use global handler if available
      const handler = (window as any).__handleSystemAlert
      if (handler) {
        handler({
          level: 'info',
          title: payload.title,
          message: payload.message,
        })
      }
    })

    return () => {
      unsubscribeAlerts()
      unsubscribeNotifications()
    }
  }, [subscribe])
}