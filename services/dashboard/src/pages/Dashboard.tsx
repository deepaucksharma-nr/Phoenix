import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Container,
  Grid,
  Paper,
  Typography,
  Box,
  Card,
  CardContent,
  Button,
  LinearProgress,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  ListItemButton,
  Divider,
  Fab,
} from '@mui/material'
import {
  TrendingDown,
  Science,
  CheckCircle,
  Schedule,
  Warning,
  ArrowForward,
  CloudUpload,
  Storage,
  Speed,
  AttachMoney,
  Settings,
  Add as AddIcon,
} from '@mui/icons-material'
import { MetricCard } from '../components/Metrics/MetricCard'
import { MetricsChart } from '../components/Metrics/MetricsChart'
import { ExperimentWizard } from '../components/ExperimentWizard'
import { WelcomeGuide } from '../components/Onboarding'
import { useExperimentStore } from '../store/useExperimentStore'
import { formatDistanceToNow } from 'date-fns'

export const Dashboard: React.FC = () => {
  const navigate = useNavigate()
  const { experiments, fetchExperiments, loading } = useExperimentStore()
  const [wizardOpen, setWizardOpen] = useState(false)
  const [welcomeOpen, setWelcomeOpen] = useState(false)
  const [systemMetrics, setSystemMetrics] = useState({
    totalHosts: 1247,
    activeExperiments: 3,
    avgCardinalityReduction: 62,
    totalCostSavings: 156000,
  })

  useEffect(() => {
    fetchExperiments()
    
    // Check if user should see onboarding
    const hasCompletedOnboarding = localStorage.getItem('phoenix_onboarding_completed')
    const hasExperiments = experiments.length > 0
    if (!hasCompletedOnboarding && !hasExperiments) {
      setWelcomeOpen(true)
    }
  }, [fetchExperiments])
  
  useEffect(() => {
    // Listen for experiment wizard event from onboarding
    const handleOpenWizard = () => setWizardOpen(true)
    window.addEventListener('openExperimentWizard', handleOpenWizard)
    return () => window.removeEventListener('openExperimentWizard', handleOpenWizard)
  }, [])

  const recentExperiments = experiments
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5)

  const runningExperiments = experiments.filter((exp) => exp.status === 'running')
  const completedExperiments = experiments.filter((exp) => exp.status === 'completed')

  const handleNavigateToExperiments = () => {
    navigate('/experiments')
  }

  const handleNavigateToExperiment = (id: string) => {
    navigate(`/experiments/${id}`)
  }

  const handleCreateExperiment = () => {
    setWizardOpen(true)
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            Dashboard
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Monitor your process metrics optimization performance
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleCreateExperiment}
          size="large"
        >
          Create Experiment
        </Button>
      </Box>

      {loading && <LinearProgress sx={{ mb: 3 }} />}

      {/* Summary Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Total Hosts"
            value={systemMetrics.totalHosts.toLocaleString()}
            icon={<Storage />}
            color="primary"
            subtitle="Monitored infrastructure"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Active Experiments"
            value={systemMetrics.activeExperiments}
            icon={<Science />}
            color="info"
            subtitle={`${runningExperiments.length} running now`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Avg. Cardinality Reduction"
            value={`${systemMetrics.avgCardinalityReduction}%`}
            icon={<TrendingDown />}
            color="success"
            change={systemMetrics.avgCardinalityReduction}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Annual Savings"
            value={`$${(systemMetrics.totalCostSavings / 1000).toFixed(0)}K`}
            icon={<AttachMoney />}
            color="success"
            subtitle="Projected for this year"
          />
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        {/* Recent Activity */}
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3, mb: 3 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="h6">Recent Experiments</Typography>
              <Button
                size="small"
                endIcon={<ArrowForward />}
                onClick={handleNavigateToExperiments}
              >
                View All
              </Button>
            </Box>
            <List>
              {recentExperiments.map((experiment, index) => (
                <React.Fragment key={experiment.id}>
                  <ListItemButton onClick={() => handleNavigateToExperiment(experiment.id)}>
                    <ListItemIcon>
                      {experiment.status === 'completed' && <CheckCircle color="success" />}
                      {experiment.status === 'running' && <Schedule color="primary" />}
                      {experiment.status === 'failed' && <Warning color="error" />}
                      {experiment.status === 'pending' && <Science color="disabled" />}
                    </ListItemIcon>
                    <ListItemText
                      primary={experiment.name}
                      secondary={`${experiment.status} • ${formatDistanceToNow(new Date(experiment.createdAt))} ago`}
                    />
                    <Chip
                      label={experiment.status}
                      size="small"
                      color={
                        experiment.status === 'completed'
                          ? 'success'
                          : experiment.status === 'running'
                          ? 'primary'
                          : experiment.status === 'failed'
                          ? 'error'
                          : 'default'
                      }
                    />
                  </ListItemButton>
                  {index < recentExperiments.length - 1 && <Divider />}
                </React.Fragment>
              ))}
            </List>
            {recentExperiments.length === 0 && (
              <Box sx={{ textAlign: 'center', py: 4 }}>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  No experiments yet
                </Typography>
                <Button
                  variant="contained"
                  size="small"
                  onClick={handleCreateExperiment}
                  sx={{ mt: 1 }}
                >
                  Create First Experiment
                </Button>
              </Box>
            )}
          </Paper>

          {/* System Performance */}
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <MetricsChart
                title="Cardinality Reduction Trend"
                data={{
                  baseline: generateTrendData(150000, 7),
                  candidate: generateTrendData(52500, 7),
                }}
                yAxisLabel="Time Series"
                height={250}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <MetricsChart
                title="Cost Savings Trend"
                data={{
                  baseline: generateTrendData(3500, 7),
                  candidate: generateTrendData(1400, 7),
                }}
                yAxisLabel="Monthly Cost ($)"
                height={250}
              />
            </Grid>
          </Grid>
        </Grid>

        {/* Quick Actions & Stats */}
        <Grid item xs={12} md={4}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Quick Actions
              </Typography>
              <List>
                <ListItemButton onClick={handleCreateExperiment}>
                  <ListItemIcon>
                    <Science color="primary" />
                  </ListItemIcon>
                  <ListItemText primary="Create New Experiment" />
                </ListItemButton>
                <ListItemButton onClick={() => navigate('/pipeline-builder')}>
                  <ListItemIcon>
                    <CloudUpload color="primary" />
                  </ListItemIcon>
                  <ListItemText primary="Build Pipeline" />
                </ListItemButton>
                <ListItemButton onClick={() => navigate('/settings')}>
                  <ListItemIcon>
                    <Settings color="primary" />
                  </ListItemIcon>
                  <ListItemText primary="Configure Settings" />
                </ListItemButton>
              </List>
            </CardContent>
          </Card>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                System Status
              </Typography>
              <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Platform Health</Typography>
                  <Chip label="Healthy" color="success" size="small" />
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">API Latency</Typography>
                  <Typography variant="body2" color="success.main">
                    45ms
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Active Collectors</Typography>
                  <Typography variant="body2">{systemMetrics.totalHosts}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">Data Ingestion Rate</Typography>
                  <Typography variant="body2">2.3M/min</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Experiment Stats
              </Typography>
              <Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Total Experiments</Typography>
                  <Typography variant="body2" fontWeight="medium">
                    {experiments.length}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Completed</Typography>
                  <Typography variant="body2" color="success.main">
                    {completedExperiments.length}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Running</Typography>
                  <Typography variant="body2" color="primary.main">
                    {runningExperiments.length}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">Success Rate</Typography>
                  <Typography variant="body2" color="success.main">
                    {completedExperiments.length > 0
                      ? `${Math.round((completedExperiments.length / experiments.length) * 100)}%`
                      : 'N/A'}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
      
      {/* Floating Action Button */}
      <Fab
        color="primary"
        aria-label="create experiment"
        onClick={handleCreateExperiment}
        sx={{
          position: 'fixed',
          bottom: 24,
          right: 24,
        }}
      >
        <AddIcon />
      </Fab>
      
      {/* Experiment Creation Wizard */}
      <ExperimentWizard
        open={wizardOpen}
        onClose={() => setWizardOpen(false)}
      />
      
      {/* Welcome Guide for New Users */}
      <WelcomeGuide
        open={welcomeOpen}
        onClose={() => setWelcomeOpen(false)}
      />
    </Container>
  )
}

// Helper function to generate trend data
function generateTrendData(baseValue: number, days: number) {
  const data = []
  const now = Date.now()
  
  for (let i = 0; i < days * 24; i++) {
    const timestamp = now - (days * 24 - i) * 3600000
    const dayProgress = (i % 24) / 24
    const dailyVariation = Math.sin(dayProgress * Math.PI * 2) * 0.1
    const trend = i / (days * 24) * -0.3 // Downward trend
    const noise = (Math.random() - 0.5) * 0.1
    const value = baseValue * (1 + dailyVariation + trend + noise)
    
    data.push({
      timestamp,
      value: Math.max(0, value),
    })
  }
  
  return data
}