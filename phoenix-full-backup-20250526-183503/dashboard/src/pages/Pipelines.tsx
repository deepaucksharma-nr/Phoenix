import React from 'react'
import { Typography, Box, Paper, Container } from '@mui/material'
import { useParams } from 'react-router-dom'
import { PipelineViewer } from '@/components/Pipeline'
import { useAppSelector } from '@/store/hooks'
import { selectPipelineById } from '@/store/slices/pipelineSlice'

export default function Pipelines() {
  const { id } = useParams<{ id: string }>()
  const pipeline = useAppSelector(state => id ? selectPipelineById(state, id) : null)

  return (
    <Container maxWidth="xl">
      <Box py={3}>
        <Typography variant="h4" gutterBottom>
          Pipeline Viewer
        </Typography>
        
        {pipeline ? (
          <Paper sx={{ p: 3, mt: 3 }}>
            <PipelineViewer pipeline={pipeline} />
          </Paper>
        ) : (
          <Typography variant="body1" color="text.secondary">
            Select a pipeline to view its configuration
          </Typography>
        )}
      </Box>
    </Container>
  )
}
