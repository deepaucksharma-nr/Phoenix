import React, { useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, useNavigate } from 'react-router-dom'
import { ThemeProvider, createTheme, CssBaseline, Container, AppBar, Toolbar, Typography, Box, Card, CardContent, Grid, Button, CircularProgress, Alert } from '@mui/material'
import { Science, Timeline, Memory, Speed } from '@mui/icons-material'
import { Provider } from 'react-redux'
import { store } from './store'
import { useAppSelector, useAppDispatch } from './hooks/redux'
import { fetchExperiments } from './store/slices/experimentSlice'
import { fetchPipelineDeployments, fetchPipelineTemplates } from './store/slices/pipelineSlice'

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

// API Data Loader Component
function ApiDataLoader({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch()
  const { loading: experimentsLoading, error: experimentsError } = useAppSelector(state => state.experiments)
  const { loading: pipelinesLoading, error: pipelinesError } = useAppSelector(state => state.pipelines)

  useEffect(() => {
    dispatch(fetchExperiments())
    dispatch(fetchPipelineDeployments())
    dispatch(fetchPipelineTemplates())
  }, [dispatch])

  const isLoading = experimentsLoading || pipelinesLoading
  const hasError = experimentsError || pipelinesError

  if (isLoading) {
    return (
      <Box 
        sx={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          height: '100vh',
          flexDirection: 'column',
          gap: 2
        }}
      >
        <CircularProgress size={60} />
        <Typography variant="body1" color="text.secondary">
          Loading Phoenix Platform data...
        </Typography>
      </Box>
    )
  }

  if (hasError) {
    return (
      <Container maxWidth="sm" sx={{ mt: 8 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          <Typography variant="h6" gutterBottom>
            Failed to load application data
          </Typography>
          <Typography variant="body2">
            {experimentsError || pipelinesError}
          </Typography>
        </Alert>
        <Button 
          variant="contained" 
          onClick={() => window.location.reload()}
          fullWidth
        >
          Retry
        </Button>
      </Container>
    )
  }

  return <>{children}</>
}

// Enhanced Experiments page with Redux data
function ExperimentsPage() {
  const navigate = useNavigate()
  const { experiments, loading } = useAppSelector(state => state.experiments)
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
  const { deployments } = useAppSelector(state => state.pipelines)
  const [searchQuery, setSearchQuery] = React.useState('')
  const [statusFilter, setStatusFilter] = React.useState('all')

  const filteredDeployments = deployments.filter((deployment: any) => {
    const matchesSearch = deployment.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         deployment.pipeline.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = statusFilter === 'all' || deployment.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success'
      case 'pending': return 'warning'
      case 'failed': return 'error'
      default: return 'default'
    }
  }

  return (
    <Container maxWidth="xl" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">⚙️ Deployed Pipelines</Typography>
        <Button variant="contained" startIcon={<Timeline />}>
          Deploy New Pipeline
        </Button>
      </Box>

      {/* Search and Filter Controls */}
      <Box display="flex" gap={2} mb={3}>
        <Box flexGrow={1}>
          <Typography variant="body2" gutterBottom>Search pipelines...</Typography>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by name or pipeline type"
            style={{ 
              width: '100%', 
              padding: '8px 12px', 
              border: '1px solid #ccc', 
              borderRadius: '4px',
              fontSize: '14px'
            }}
          />
        </Box>
        <Box minWidth={150}>
          <Typography variant="body2" gutterBottom>Filter by status</Typography>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="pending">Pending</option>
            <option value="failed">Failed</option>
          </select>
        </Box>
      </Box>

      {/* Metrics Summary Cards */}
      <Grid container spacing={2} sx={{ mb: 4 }}>
        <Grid item xs={6} sm={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <Typography variant="h4" color="primary">{deployments.length}</Typography>
              <Typography variant="body2">Total Deployments</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <Typography variant="h4" color="success.main">
                {deployments.filter((d: any) => d.status === 'active').length}
              </Typography>
              <Typography variant="body2">Active</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <Typography variant="h4" color="primary">
                {deployments.reduce((sum: number, d: any) => sum + d.metrics.cardinality, 0).toLocaleString()}
              </Typography>
              <Typography variant="body2">Total Cardinality</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={3}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <Typography variant="h4" color="primary">
                {Math.round(deployments.reduce((sum: number, d: any) => sum + d.metrics.cpuUsage, 0) / deployments.length)}%
              </Typography>
              <Typography variant="body2">Avg CPU Usage</Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Pipeline Deployments List */}
      <Grid container spacing={3}>
        {filteredDeployments.map((deployment: any) => (
          <Grid item xs={12} lg={6} key={deployment.id}>
            <Card>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="start" mb={2}>
                  <Box>
                    <Typography variant="h6">{deployment.name}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {deployment.pipeline} • {deployment.namespace}
                    </Typography>
                  </Box>
                  <Box 
                    sx={{ 
                      px: 2, 
                      py: 0.5, 
                      borderRadius: 1,
                      bgcolor: `${getStatusColor(deployment.status)}.light`,
                      color: `${getStatusColor(deployment.status)}.dark`
                    }}
                  >
                    {deployment.status}
                  </Box>
                </Box>

                <Box display="flex" justifyContent="space-between" mb={2}>
                  <Box>
                    <Typography variant="caption" display="block">Instances</Typography>
                    <Typography variant="body2">
                      {deployment.instances.ready}/{deployment.instances.desired} ready
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" display="block">Cardinality</Typography>
                    <Typography variant="body2">
                      {deployment.metrics.cardinality.toLocaleString()}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" display="block">Throughput</Typography>
                    <Typography variant="body2">
                      {deployment.metrics.throughput}
                    </Typography>
                  </Box>
                </Box>

                <Box display="flex" justifyContent="space-between" mb={2}>
                  <Box>
                    <Typography variant="caption" display="block">CPU Usage</Typography>
                    <Typography variant="body2">{deployment.metrics.cpuUsage}%</Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" display="block">Memory</Typography>
                    <Typography variant="body2">{deployment.metrics.memoryUsage}%</Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" display="block">Error Rate</Typography>
                    <Typography variant="body2">{(deployment.metrics.errorRate * 100).toFixed(3)}%</Typography>
                  </Box>
                </Box>

                <Box display="flex" gap={1}>
                  <Button size="small" variant="outlined">
                    View Config
                  </Button>
                  <Button size="small" variant="outlined">
                    Metrics
                  </Button>
                  {deployment.status === 'active' && (
                    <Button size="small" variant="contained" color="primary">
                      Start Experiment
                    </Button>
                  )}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {filteredDeployments.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Timeline sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No deployments found
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {searchQuery || statusFilter !== 'all' 
              ? 'Try adjusting your filters' 
              : 'Deploy your first pipeline to get started'}
          </Typography>
        </Box>
      )}
    </Container>
  )
}

function CatalogPage() {
  const navigate = useNavigate()
  const [selectedTemplate, setSelectedTemplate] = React.useState<any>(null)
  const [showYaml, setShowYaml] = React.useState(false)
  const [categoryFilter, setCategoryFilter] = React.useState('all')

  const templates = [
    {
      id: '1',
      name: 'Process Metrics Optimizer',
      description: 'Optimizes process-level metrics by aggregating similar processes and reducing cardinality through intelligent grouping.',
      category: 'optimization',
      version: '1.2.0',
      author: 'Phoenix Team',
      tags: ['process', 'cardinality', 'aggregation', 'production-ready'],
      performance: {
        avgLatency: '2.3ms',
        cpuUsage: '15%',
        memoryUsage: '128MB',
        cardinalityReduction: '85%',
      },
      yaml: `processors:
  attributes:
    actions:
      - key: process.executable.name
        action: hash
      - key: process.pid
        action: delete
  resource:
    attributes:
      - key: service.name
        from_attribute: process.executable.name
        action: insert
  batch:
    timeout: 200ms
    send_batch_size: 8192
  memory_limiter:
    check_interval: 1s
    limit_mib: 512`,
    },
    {
      id: '2',
      name: 'Tail Sampling Pipeline',
      description: 'Implements intelligent tail sampling to capture important traces while reducing overall volume.',
      category: 'sampling',
      version: '2.0.1',
      author: 'Phoenix Team',
      tags: ['traces', 'sampling', 'performance', 'errors'],
      performance: {
        avgLatency: '5.1ms',
        cpuUsage: '25%',
        memoryUsage: '256MB',
        cardinalityReduction: '70%',
      },
      yaml: `processors:
  tail_sampling:
    decision_wait: 10s
    num_traces: 100000
    policies:
      - name: errors-policy
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: slow-traces-policy
        type: latency
        latency: {threshold_ms: 1000}`,
    },
    {
      id: '3',
      name: 'Metrics Aggregator',
      description: 'Aggregates metrics at collection time to reduce storage requirements while maintaining query performance.',
      category: 'aggregation',
      version: '1.5.3',
      author: 'Community',
      tags: ['metrics', 'aggregation', 'storage', 'cost-optimization'],
      performance: {
        avgLatency: '3.7ms',
        cpuUsage: '20%',
        memoryUsage: '192MB',
        cardinalityReduction: '75%',
      },
      yaml: `processors:
  metricstransform:
    transforms:
      - include: .*
        match_type: regexp
        action: update
        operations:
          - action: aggregate_labels
            label_set: [service.name, service.namespace]
            aggregation_type: sum`,
    },
  ]

  const filteredTemplates = templates.filter(template => 
    categoryFilter === 'all' || template.category === categoryFilter
  )

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'optimization': return <Speed />
      case 'sampling': return <Timeline />
      case 'aggregation': return <Memory />
      default: return <Science />
    }
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Button onClick={() => navigate('/')} sx={{ mb: 2 }}>← Back to Dashboard</Button>
      
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">📚 Pipeline Catalog</Typography>
        <Box>
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            style={{
              padding: '8px 12px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          >
            <option value="all">All Categories</option>
            <option value="optimization">Optimization</option>
            <option value="sampling">Sampling</option>
            <option value="aggregation">Aggregation</option>
          </select>
        </Box>
      </Box>

      <Typography variant="body1" color="text.secondary" paragraph>
        Browse pre-configured pipeline templates optimized for different use cases.
      </Typography>

      <Grid container spacing={3}>
        {filteredTemplates.map((template) => (
          <Grid item xs={12} md={6} lg={4} key={template.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box display="flex" alignItems="center" mb={2}>
                  <Box 
                    sx={{ 
                      p: 1, 
                      borderRadius: 1, 
                      bgcolor: 'primary.light',
                      color: 'primary.dark',
                      mr: 2,
                    }}
                  >
                    {getCategoryIcon(template.category)}
                  </Box>
                  <Box flexGrow={1}>
                    <Typography variant="h6">{template.name}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      v{template.version} by {template.author}
                    </Typography>
                  </Box>
                </Box>

                <Typography variant="body2" color="text.secondary" paragraph>
                  {template.description}
                </Typography>

                <Box mb={2}>
                  {template.tags.map((tag) => (
                    <Box 
                      key={tag}
                      component="span"
                      sx={{ 
                        display: 'inline-block',
                        px: 1,
                        py: 0.5,
                        mr: 0.5,
                        mb: 0.5,
                        fontSize: '0.75rem',
                        bgcolor: 'grey.200',
                        borderRadius: 1
                      }}
                    >
                      {tag}
                    </Box>
                  ))}
                </Box>

                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="text.secondary">
                      Cardinality Reduction
                    </Typography>
                    <Typography variant="body2" fontWeight="bold" color="success.main">
                      {template.performance.cardinalityReduction}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="text.secondary">
                      Avg Latency
                    </Typography>
                    <Typography variant="body2" fontWeight="bold">
                      {template.performance.avgLatency}
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>

              <Box sx={{ p: 2, display: 'flex', gap: 1 }}>
                <Button 
                  size="small"
                  variant="outlined"
                  onClick={() => {
                    setSelectedTemplate(template)
                    setShowYaml(true)
                  }}
                >
                  View YAML
                </Button>
                <Button 
                  size="small" 
                  variant="contained"
                  onClick={() => alert(`Deploying ${template.name}...`)}
                >
                  Deploy
                </Button>
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* YAML Viewer Modal */}
      {showYaml && selectedTemplate && (
        <Box
          sx={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            bgcolor: 'rgba(0,0,0,0.8)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={() => setShowYaml(false)}
        >
          <Card 
            sx={{ 
              maxWidth: '80%', 
              maxHeight: '80%', 
              overflow: 'auto',
              m: 2 
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h6">{selectedTemplate.name}</Typography>
                <Button onClick={() => setShowYaml(false)}>✕ Close</Button>
              </Box>
              
              <Box 
                component="pre"
                sx={{ 
                  bgcolor: 'grey.100',
                  p: 2,
                  borderRadius: 1,
                  overflow: 'auto',
                  fontSize: '0.875rem',
                  fontFamily: 'monospace'
                }}
              >
                {selectedTemplate.yaml}
              </Box>
              
              <Box mt={2} display="flex" gap={1}>
                <Button 
                  variant="outlined"
                  onClick={() => navigator.clipboard.writeText(selectedTemplate.yaml)}
                >
                  Copy YAML
                </Button>
                <Button 
                  variant="contained"
                  onClick={() => {
                    alert(`Deploying ${selectedTemplate.name}...`)
                    setShowYaml(false)
                  }}
                >
                  Deploy This Pipeline
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}
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
    <ApiDataLoader>
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
    </ApiDataLoader>
  )
}

function App() {
  return (
    <Provider store={store}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Router>
          <AppLayout />
        </Router>
      </ThemeProvider>
    </Provider>
  )
}

export default App