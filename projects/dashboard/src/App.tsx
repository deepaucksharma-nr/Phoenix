import React from 'react'
import { Box, Typography, Container } from '@mui/material'

function App() {
  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ textAlign: 'center' }}>
        <Typography variant="h3" component="h1" gutterBottom>
          🔥 Phoenix Dashboard
        </Typography>
        <Typography variant="h5" color="text.secondary">
          Process Metrics Optimization Platform
        </Typography>
        <Typography variant="body1" sx={{ mt: 3 }}>
          Dashboard is loading successfully! 🎉
        </Typography>
      </Box>
    </Container>
  )
}

export default App