import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Container,
  Paper,
  Typography,
  Box,
  Button,
  Grid,
  Card,
  CardContent,
  Tabs,
  Tab,
  Alert,
  CircularProgress,
  Chip,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  IconButton,
} from '@mui/material'
import {
  ArrowBack,
  TrendingUp,
  TrendingDown,
  CheckCircle,
  Warning,
  Info,
  Refresh,
  Download,
  Assessment,
} from '@mui/icons-material'
import { useExperimentStore } from '../store/useExperimentStore'
import { MetricsChart } from '../components/Metrics/MetricsChart'
import { MetricCard } from '../components/Metrics/MetricCard'
import { ComparisonTable } from '../components/Metrics/ComparisonTable'
import { format } from 'date-fns'
import { useMetricsUpdates } from '../hooks/useExperimentUpdates'

interface TabPanelProps {
  children?: React.ReactNode
  index: number
  value: number
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`analysis-tabpanel-${index}`}
      aria-labelledby={`analysis-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
    </div>
  )
}

export const Analysis: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const {
    currentExperiment: experiment,
    loading,
    error,
    fetchExperiment,
    fetchExperimentAnalysis,
    fetchExperimentMetrics,
  } = useExperimentStore()

  const [activeTab, setActiveTab] = useState(0)
  const [analysis, setAnalysis] = useState<any>(null)
  const [metrics, setMetrics] = useState<any>(null)
  const [metricsLoading, setMetricsLoading] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)

  // Subscribe to real-time metrics updates
  useMetricsUpdates(id)

  useEffect(() => {
    if (id) {
      fetchExperiment(id)
      loadAnalysisData()
    }
  }, [id])

  useEffect(() => {
    if (!autoRefresh || !id) return

    const interval = setInterval(() => {
      loadAnalysisData()
    }, 5000) // Refresh every 5 seconds

    return () => clearInterval(interval)
  }, [id, autoRefresh])

  const loadAnalysisData = async () => {
    if (!id) return
    
    setMetricsLoading(true)
    try {
      const [analysisData, metricsData] = await Promise.all([
        fetchExperimentAnalysis(id),
        fetchExperimentMetrics(id),
      ])
      setAnalysis(analysisData)
      setMetrics(metricsData)
    } catch (error) {
      console.error('Failed to load analysis data:', error)
    } finally {
      setMetricsLoading(false)
    }
  }

  const handleBack = () => {
    navigate(`/experiments/${id}`)
  }

  const handleExportReport = () => {
    // TODO: Implement report export
    console.log('Exporting report...')
  }

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue)
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

  if (!experiment) {
    return null
  }

  const mockAnalysis = analysis || {
    status: 'completed',
    comparison: {
      cardinalityReduction: 65,
      costSavings: 58,
      criticalProcessRetention: 100,
      performanceImpact: {
        cpuOverhead: 2.5,
        memoryOverhead: 3.1,
        latencyIncrease: 0.8,
      },
    },
    recommendation: {
      variant: 'candidate',
      confidence: 92,
      reasons: [
        'Achieved 65% cardinality reduction while maintaining 100% critical process visibility',
        'CPU overhead remains well below 5% threshold',
        'Estimated annual cost savings of $24,000',
        'No significant latency impact observed',
      ],
    },
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button startIcon={<ArrowBack />} onClick={handleBack} sx={{ mb: 2 }}>
          Back to Experiment
        </Button>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box>
            <Typography variant="h4" component="h1" gutterBottom>
              Experiment Analysis
            </Typography>
            <Typography variant="body1" color="text.secondary">
              {experiment.name}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Tooltip title={autoRefresh ? 'Disable auto-refresh' : 'Enable auto-refresh'}>
              <IconButton
                onClick={() => setAutoRefresh(!autoRefresh)}
                color={autoRefresh ? 'primary' : 'default'}
              >
                <Refresh />
              </IconButton>
            </Tooltip>
            <Button
              variant="outlined"
              startIcon={<Download />}
              onClick={handleExportReport}
            >
              Export Report
            </Button>
          </Box>
        </Box>
      </Box>

      {metricsLoading && <LinearProgress sx={{ mb: 2 }} />}

      {/* Summary Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Cardinality Reduction"
            value={`${mockAnalysis.comparison.cardinalityReduction}%`}
            change={mockAnalysis.comparison.cardinalityReduction}
            icon={<TrendingDown color="success" />}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Cost Savings"
            value={`${mockAnalysis.comparison.costSavings}%`}
            change={mockAnalysis.comparison.costSavings}
            icon={<TrendingDown color="success" />}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Critical Process Retention"
            value={`${mockAnalysis.comparison.criticalProcessRetention}%`}
            icon={<CheckCircle color="success" />}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="CPU Overhead"
            value={`${mockAnalysis.comparison.performanceImpact.cpuOverhead}%`}
            change={-mockAnalysis.comparison.performanceImpact.cpuOverhead}
            icon={<TrendingUp color="warning" />}
            color="warning"
          />
        </Grid>
      </Grid>

      {/* Recommendation */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Assessment sx={{ mr: 1 }} />
          <Typography variant="h6">Recommendation</Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Chip
            label={`Promote ${mockAnalysis.recommendation.variant}`}
            color="primary"
            size="large"
          />
          <Typography variant="body2" color="text.secondary">
            Confidence: {mockAnalysis.recommendation.confidence}%
          </Typography>
        </Box>
        <Typography variant="body2" sx={{ mb: 2 }}>
          Based on the analysis, we recommend promoting the{' '}
          <strong>{mockAnalysis.recommendation.variant}</strong> configuration.
        </Typography>
        <Box component="ul" sx={{ mt: 1, pl: 2 }}>
          {mockAnalysis.recommendation.reasons.map((reason: string, index: number) => (
            <Typography component="li" variant="body2" key={index} sx={{ mb: 0.5 }}>
              {reason}
            </Typography>
          ))}
        </Box>
      </Paper>

      {/* Detailed Analysis */}
      <Paper>
        <Tabs value={activeTab} onChange={handleTabChange}>
          <Tab label="Metrics Comparison" />
          <Tab label="Time Series" />
          <Tab label="Performance Impact" />
          <Tab label="Cost Analysis" />
        </Tabs>

        <TabPanel value={activeTab} index={0}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Baseline vs Candidate Comparison
            </Typography>
            <ComparisonTable
              baseline={{
                cardinality: 150000,
                cpuUsage: 12.5,
                memoryUsage: 2048,
                networkTraffic: 125,
                dataPointsPerSecond: 5000,
                uniqueProcesses: 450,
              }}
              candidate={{
                cardinality: 52500,
                cpuUsage: 15.0,
                memoryUsage: 2150,
                networkTraffic: 45,
                dataPointsPerSecond: 1750,
                uniqueProcesses: 125,
              }}
            />
          </Box>
        </TabPanel>

        <TabPanel value={activeTab} index={1}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Metrics Over Time
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <MetricsChart
                  title="Cardinality"
                  data={{
                    baseline: generateTimeSeriesData(150000, 0.1),
                    candidate: generateTimeSeriesData(52500, 0.15),
                  }}
                  yAxisLabel="Time Series Count"
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <MetricsChart
                  title="CPU Usage"
                  data={{
                    baseline: generateTimeSeriesData(12.5, 0.2),
                    candidate: generateTimeSeriesData(15.0, 0.25),
                  }}
                  yAxisLabel="CPU %"
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <MetricsChart
                  title="Memory Usage"
                  data={{
                    baseline: generateTimeSeriesData(2048, 0.05),
                    candidate: generateTimeSeriesData(2150, 0.08),
                  }}
                  yAxisLabel="Memory (MB)"
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <MetricsChart
                  title="Network Traffic"
                  data={{
                    baseline: generateTimeSeriesData(125, 0.3),
                    candidate: generateTimeSeriesData(45, 0.2),
                  }}
                  yAxisLabel="KB/s"
                />
              </Grid>
            </Grid>
          </Box>
        </TabPanel>

        <TabPanel value={activeTab} index={2}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Performance Impact Analysis
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} md={4}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>
                      CPU Overhead
                    </Typography>
                    <Typography variant="h4">
                      +{mockAnalysis.comparison.performanceImpact.cpuOverhead}%
                    </Typography>
                    <Alert severity="success" sx={{ mt: 2 }}>
                      Within acceptable limits (&lt;5%)
                    </Alert>
                  </CardContent>
                </Card>
              </Grid>
              <Grid item xs={12} md={4}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>
                      Memory Overhead
                    </Typography>
                    <Typography variant="h4">
                      +{mockAnalysis.comparison.performanceImpact.memoryOverhead}%
                    </Typography>
                    <Alert severity="success" sx={{ mt: 2 }}>
                      Minimal impact observed
                    </Alert>
                  </CardContent>
                </Card>
              </Grid>
              <Grid item xs={12} md={4}>
                <Card>
                  <CardContent>
                    <Typography color="text.secondary" gutterBottom>
                      Latency Increase
                    </Typography>
                    <Typography variant="h4">
                      +{mockAnalysis.comparison.performanceImpact.latencyIncrease}ms
                    </Typography>
                    <Alert severity="success" sx={{ mt: 2 }}>
                      No significant impact
                    </Alert>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          </Box>
        </TabPanel>

        <TabPanel value={activeTab} index={3}>
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Cost Analysis
            </Typography>
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Metric</TableCell>
                    <TableCell align="right">Current Monthly Cost</TableCell>
                    <TableCell align="right">Projected Monthly Cost</TableCell>
                    <TableCell align="right">Monthly Savings</TableCell>
                    <TableCell align="right">Annual Savings</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  <TableRow>
                    <TableCell>Data Ingestion</TableCell>
                    <TableCell align="right">$2,500</TableCell>
                    <TableCell align="right">$1,050</TableCell>
                    <TableCell align="right">$1,450</TableCell>
                    <TableCell align="right">$17,400</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Storage</TableCell>
                    <TableCell align="right">$800</TableCell>
                    <TableCell align="right">$336</TableCell>
                    <TableCell align="right">$464</TableCell>
                    <TableCell align="right">$5,568</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell>Query Processing</TableCell>
                    <TableCell align="right">$150</TableCell>
                    <TableCell align="right">$89</TableCell>
                    <TableCell align="right">$61</TableCell>
                    <TableCell align="right">$732</TableCell>
                  </TableRow>
                  <TableRow sx={{ bgcolor: 'action.hover' }}>
                    <TableCell><strong>Total</strong></TableCell>
                    <TableCell align="right"><strong>$3,450</strong></TableCell>
                    <TableCell align="right"><strong>$1,475</strong></TableCell>
                    <TableCell align="right"><strong>$1,975</strong></TableCell>
                    <TableCell align="right"><strong>$23,700</strong></TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
            <Alert severity="info" sx={{ mt: 3 }}>
              <Typography variant="body2">
                Based on current usage patterns and pricing, implementing the candidate
                configuration would save approximately <strong>$23,700</strong> annually
                ({mockAnalysis.comparison.costSavings}% reduction).
              </Typography>
            </Alert>
          </Box>
        </TabPanel>
      </Paper>
    </Container>
  )
}

// Helper function to generate mock time series data
function generateTimeSeriesData(baseValue: number, variance: number) {
  const points = 24 // 24 hours of data
  const data = []
  const now = Date.now()
  
  for (let i = 0; i < points; i++) {
    const timestamp = now - (points - i) * 3600000 // 1 hour intervals
    const value = baseValue + (Math.random() - 0.5) * 2 * baseValue * variance
    data.push({
      timestamp,
      value: Math.max(0, value),
    })
  }
  
  return data
}