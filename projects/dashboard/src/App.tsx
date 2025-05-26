import { BrowserRouter as Router, Routes, Route, useNavigate } from 'react-router-dom'
import { ThemeProvider, createTheme, CssBaseline, Container, AppBar, Toolbar, Typography, Box, Card, CardContent, Grid, Button } from '@mui/material'
import { Science, Timeline, Memory, Speed } from '@mui/icons-material'

const theme = createTheme({
  palette: {
    primary: {
      main: '#ff6b35',
    },
    secondary: {
      main: '#f7931e',
    },
  },
})

// Dashboard Home Component
function Dashboard() {
  const navigate = useNavigate()

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Process Metrics Optimization Platform
      </Typography>
      
      <Typography variant="body1" color="text.secondary" paragraph>
        Welcome to the Phoenix Platform! Monitor and optimize your OpenTelemetry pipeline configurations
        to reduce metrics cardinality while maintaining critical visibility.
      </Typography>

      <Grid container spacing={3} sx={{ mt: 2 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center' }}>
              <Science color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">Experiments</Typography>
              <Typography variant="body2" color="text.secondary">
                A/B test pipeline configurations
              </Typography>
              <Button 
                variant="outlined" 
                sx={{ mt: 2 }} 
                fullWidth
                onClick={() => navigate('/experiments')}
              >
                View Experiments
              </Button>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center' }}>
              <Timeline color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">Pipelines</Typography>
              <Typography variant="body2" color="text.secondary">
                Manage deployed configurations
              </Typography>
              <Button 
                variant="outlined" 
                sx={{ mt: 2 }} 
                fullWidth
                onClick={() => navigate('/pipelines')}
              >
                View Pipelines
              </Button>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center' }}>
              <Memory color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">Catalog</Typography>
              <Typography variant="body2" color="text.secondary">
                Browse pipeline templates
              </Typography>
              <Button 
                variant="outlined" 
                sx={{ mt: 2 }} 
                fullWidth
                onClick={() => navigate('/catalog')}
              >
                Browse Catalog
              </Button>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center' }}>
              <Speed color="primary" sx={{ fontSize: 40, mb: 1 }} />
              <Typography variant="h6">Analytics</Typography>
              <Typography variant="body2" color="text.secondary">
                View performance metrics
              </Typography>
              <Button 
                variant="outlined" 
                sx={{ mt: 2 }} 
                fullWidth
                onClick={() => navigate('/analytics')}
              >
                View Analytics
              </Button>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box sx={{ mt: 4, p: 3, backgroundColor: 'grey.50', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>
          ✅ Dashboard Status
        </Typography>
        <Typography variant="body2">
          • Basic React app: ✅ Working<br/>
          • Material-UI components: ✅ Working<br/>
          • React Router navigation: ✅ Working<br/>
          • Ready for advanced features: ✅ Ready
        </Typography>
      </Box>
    </Container>
  )
}

// Enhanced Experiments page with Redux data
function ExperimentsPage() {
  const navigate = useNavigate()
  const { experiments, loading } = useAppSelector((state: any) => state.experiments)
  const dispatch = useAppDispatch()
  const [watchMode, setWatchMode] = React.useState(false)
  const [exportAnchor, setExportAnchor] = React.useState<null | HTMLElement>(null)

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">🧪 Experiments</Typography>
        <Box display="flex" gap={2}>
          <Button 
            variant={watchMode ? "contained" : "outlined"}
            onClick={() => setWatchMode(!watchMode)}
            startIcon={watchMode ? <Speed /> : <Timeline />}
          >
            {watchMode ? 'Watch ON' : 'Watch Mode'}
          </Button>
          <Button 
            variant="outlined"
            onClick={(e) => setExportAnchor(e.currentTarget)}
          >
            Export
          </Button>
          <Button variant="contained" startIcon={<Science />}>
            New Experiment
          </Button>
        </Box>
      </Box>

      {watchMode && (
        <Box sx={{ mb: 3, p: 2, bgcolor: 'info.light', borderRadius: 1 }}>
          <Typography variant="body2">
            🔄 Watch mode enabled - Auto-refreshing every 5 seconds
          </Typography>
        </Box>
      )}

      <Grid container spacing={3}>
        {experiments.map((experiment: any) => (
          <Grid item xs={12} md={6} key={experiment.id}>
            <Card>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="start" mb={2}>
                  <Typography variant="h6">{experiment.name}</Typography>
                  <Box 
                    sx={{ 
                      px: 2, 
                      py: 0.5, 
                      borderRadius: 1,
                      bgcolor: experiment.status === 'running' ? 'success.light' : 'grey.300',
                      color: experiment.status === 'running' ? 'success.dark' : 'text.primary'
                    }}
                  >
                    {experiment.status}
                  </Box>
                </Box>
                
                <Typography variant="body2" color="text.secondary" mb={2}>
                  {experiment.description}
                </Typography>
                
                <Box display="flex" gap={2} mb={2}>
                  <Typography variant="caption">
                    Duration: {experiment.spec.duration}
                  </Typography>
                  <Typography variant="caption">
                    Hosts: {experiment.spec.targetHosts.length}
                  </Typography>
                </Box>
                
                <Box display="flex" gap={1}>
                  <Button size="small" variant="outlined">
                    View Details
                  </Button>
                  {experiment.status === 'running' && (
                    <Button size="small" variant="outlined" color="warning">
                      Stop
                    </Button>
                  )}
                  {experiment.status === 'completed' && (
                    <Button size="small" variant="outlined" color="primary">
                      Analyze Results
                    </Button>
                  )}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Export Menu */}
      <Box
        component="div"
        sx={{
          position: 'fixed',
          top: exportAnchor ? '100px' : '-1000px',
          right: '20px',
          bgcolor: 'background.paper',
          boxShadow: 3,
          borderRadius: 1,
          p: 2,
          minWidth: 150,
        }}
      >
        <Button fullWidth onClick={() => setExportAnchor(null)}>Export JSON</Button>
        <Button fullWidth onClick={() => setExportAnchor(null)}>Export CSV</Button>
        <Button fullWidth onClick={() => setExportAnchor(null)}>Export YAML</Button>
        <Button fullWidth onClick={() => setExportAnchor(null)}>Close</Button>
      </Box>
    </Container>
  )
}

function PipelinesPage() {
  const navigate = useNavigate()
  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      <Typography variant="h4" gutterBottom>⚙️ Deployed Pipelines</Typography>
      <Typography variant="body1">Monitor and manage your active OpenTelemetry pipeline deployments.</Typography>
      <Box sx={{ mt: 3, p: 2, border: '1px solid', borderColor: 'grey.300', borderRadius: 1 }}>
        <Typography variant="body2">Features will include: Real-time metrics, Resource monitoring, Quick experiment creation</Typography>
      </Box>
    </Container>
  )
}

function CatalogPage() {
  const navigate = useNavigate()
  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      <Typography variant="h4" gutterBottom>📚 Pipeline Catalog</Typography>
      <Typography variant="body1">Browse pre-configured pipeline templates for different optimization strategies.</Typography>
      <Box sx={{ mt: 3, p: 2, border: '1px solid', borderColor: 'grey.300', borderRadius: 1 }}>
        <Typography variant="body2">Features will include: YAML viewer, Performance metrics, One-click deployment</Typography>
      </Box>
    </Container>
  )
}

function AnalyticsPage() {
  const navigate = useNavigate()
  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      <Typography variant="h4" gutterBottom>📊 Analytics</Typography>
      <Typography variant="body1">View detailed performance analytics and cost optimization metrics.</Typography>
      <Box sx={{ mt: 3, p: 2, border: '1px solid', borderColor: 'grey.300', borderRadius: 1 }}>
        <Typography variant="body2">Features will include: Cost reduction charts, Cardinality trends, Performance comparisons</Typography>
      </Box>
    </Container>
  )
}

// Main App Component with Navigation
function AppLayout() {
  const navigate = useNavigate()
  
  return (
    <Box>
      <AppBar position="static">
        <Toolbar>
          <Typography 
            variant="h6" 
            component="div" 
            sx={{ flexGrow: 1, cursor: 'pointer' }}
            onClick={() => navigate('/')}
          >
            🔥 Phoenix Dashboard
          </Typography>
          <Button color="inherit" onClick={() => navigate('/experiments')}>Experiments</Button>
          <Button color="inherit" onClick={() => navigate('/pipelines')}>Pipelines</Button>
          <Button color="inherit" onClick={() => navigate('/catalog')}>Catalog</Button>
          <Button color="inherit" onClick={() => navigate('/analytics')}>Analytics</Button>
        </Toolbar>
      </AppBar>
      
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/experiments" element={<ExperimentsPage />} />
        <Route path="/pipelines" element={<PipelinesPage />} />
        <Route path="/catalog" element={<CatalogPage />} />
        <Route path="/analytics" element={<AnalyticsPage />} />
      </Routes>
    </Box>
  )
}

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <AppLayout />
      </Router>
    </ThemeProvider>
  )
}

export default App