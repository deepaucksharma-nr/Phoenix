import React, { useState, useCallback } from 'react'
import { Box } from '@mui/material'
import { PipelineCanvas } from '../ExperimentBuilder/PipelineCanvas'
import { ProcessorLibrary } from './ProcessorLibrary'
import { PipelineConfig } from '@/types/pipeline'

interface PipelineBuilderProps {
  initialConfig?: PipelineConfig
  onSave?: (config: PipelineConfig) => void
  onRun?: (config: PipelineConfig) => void
}

export const PipelineBuilder: React.FC<PipelineBuilderProps> = ({
  initialConfig,
  onSave,
  onRun,
}) => {
  const [selectedProcessor, setSelectedProcessor] = useState<string | null>(null)

  const handleProcessorDrag = useCallback((processorType: string) => {
    setSelectedProcessor(processorType)
  }, [])

  const handleSave = useCallback((config: PipelineConfig) => {
    console.log('Saving pipeline configuration:', config)
    onSave?.(config)
  }, [onSave])

  const handleRun = useCallback((config: PipelineConfig) => {
    console.log('Running pipeline configuration:', config)
    onRun?.(config)
  }, [onRun])

  return (
    <Box sx={{ display: 'flex', height: '100vh', width: '100%' }}>
      {/* Processor Library Sidebar */}
      <ProcessorLibrary onDragStart={handleProcessorDrag} />
      
      {/* Main Canvas Area */}
      <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <PipelineCanvas
          initialConfig={initialConfig}
          onSave={handleSave}
          onRun={handleRun}
        />
      </Box>
    </Box>
  )
}