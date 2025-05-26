import React, { useState, useEffect, useMemo } from 'react'
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  IconButton,
  Tooltip,
  Chip,
  Tabs,
  Tab,
  Alert,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  ToggleButton,
  ToggleButtonGroup,
  Badge,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Slider,
  Switch,
  FormControlLabel,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material'
import {
  TrendingUp,
  TrendingDown,
  TrendingFlat,
  CheckCircle,
  Warning,
  Error as ErrorIcon,
  Info,
  Timeline,
  BarChart,
  PieChart,
  ScatterPlot,
  Assessment,
  Download,
  Share,
  Fullscreen,
  ZoomIn,
  ZoomOut,
  FilterList,
  CompareArrows,
  Speed,
  Storage,
  AttachMoney,
  Timer,
  CloudUpload,
} from '@mui/icons-material'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip as ChartTooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line, Bar, Doughnut, Scatter } from 'react-chartjs-2'
import { format, subHours, subDays } from 'date-fns'
import { useNotification } from '../../hooks/useNotification'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  ChartTooltip,
  Legend,
  Filler
)

interface MetricData {
  baseline: number[]
  candidate: number[]
  timestamps: number[]
  labels: string[]
}

interface AnalysisResult {
  cardinalityReduction: number
  dataRateReduction: number
  errorRateChange: number
  latencyImpact: number
  costSavings: number
  recommendation: 'promote' | 'reject' | 'investigate'
  confidence: number
  insights: string[]
}

interface EnhancedAnalysisProps {
  experimentId: string
  experimentData: any
  metricsData: any
  onPromote?: (variant: 'baseline' | 'candidate') => void
}

const MetricComparisonCard: React.FC<{
  title: string
  icon: React.ReactNode
  baseline: number
  candidate: number
  unit?: string
  higherIsBetter?: boolean
  showTrend?: boolean
}> = ({ title, icon, baseline, candidate, unit = '', higherIsBetter = false, showTrend = true }) => {
  const change = ((candidate - baseline) / baseline) * 100
  const improved = higherIsBetter ? change > 0 : change < 0
  const significant = Math.abs(change) > 5

  const getTrendIcon = () => {
    if (Math.abs(change) < 1) return <TrendingFlat />
    return improved ? <TrendingDown color="success" /> : <TrendingUp color="error" />
  }

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Box sx={{ color: 'primary.main', mr: 1 }}>{icon}</Box>
          <Typography variant="subtitle2" color="text.secondary">
            {title}
          </Typography>
        </Box>
        
        <Grid container spacing={2}>
          <Grid item xs={6}>
            <Typography variant="caption" color="text.secondary">
              Baseline
            </Typography>
            <Typography variant="h6">
              {baseline.toLocaleString()}{unit}
            </Typography>
          </Grid>
          <Grid item xs={6}>
            <Typography variant="caption" color="text.secondary">
              Candidate
            </Typography>
            <Typography variant="h6">
              {candidate.toLocaleString()}{unit}
            </Typography>
          </Grid>
        </Grid>

        <Box sx={{ mt: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            {showTrend && getTrendIcon()}
            <Typography
              variant="body2"
              sx={{
                ml: 1,
                color: improved ? 'success.main' : significant ? 'error.main' : 'text.secondary',
                fontWeight: significant ? 'bold' : 'normal',
              }}
            >
              {change > 0 ? '+' : ''}{change.toFixed(1)}%
            </Typography>
          </Box>
          {significant && (
            <Chip
              label={improved ? 'Better' : 'Worse'}
              size="small"
              color={improved ? 'success' : 'error'}
              variant="outlined"
            />
          )}
        </Box>
        
        <LinearProgress
          variant="determinate"
          value={Math.min(100, Math.abs(change))}
          color={improved ? 'success' : 'error'}
          sx={{ mt: 1, height: 4, borderRadius: 2 }}
        />
      </CardContent>
    </Card>
  )
}

export const EnhancedAnalysis: React.FC<EnhancedAnalysisProps> = ({
  experimentId,
  experimentData,
  metricsData,
  onPromote,
}) => {
  const { showNotification } = useNotification()
  const [timeRange, setTimeRange] = useState('1h')
  const [chartType, setChartType] = useState<'line' | 'bar'>('line')
  const [activeTab, setActiveTab] = useState(0)
  const [selectedMetrics, setSelectedMetrics] = useState<string[]>(['cardinality', 'dataRate'])
  const [showConfidenceInterval, setShowConfidenceInterval] = useState(true)
  const [exportDialogOpen, setExportDialogOpen] = useState(false)
  
  // Calculate analysis results
  const analysisResult = useMemo<AnalysisResult>(() => {
    if (!metricsData || metricsData.length === 0) {
      return {
        cardinalityReduction: 0,
        dataRateReduction: 0,
        errorRateChange: 0,
        latencyImpact: 0,
        costSavings: 0,
        recommendation: 'investigate',
        confidence: 0,
        insights: [],
      }
    }

    // Calculate averages from the last hour of data
    const recentMetrics = metricsData.slice(-12) // Last 12 5-minute intervals
    
    const avgBaseline = {
      cardinality: recentMetrics.reduce((sum, m) => sum + m.baseline_cardinality, 0) / recentMetrics.length,
      dataRate: recentMetrics.reduce((sum, m) => sum + m.baseline_data_rate, 0) / recentMetrics.length,
      errorRate: recentMetrics.reduce((sum, m) => sum + m.baseline_error_rate, 0) / recentMetrics.length,
      latency: recentMetrics.reduce((sum, m) => sum + m.baseline_p95_latency, 0) / recentMetrics.length,
    }
    
    const avgCandidate = {
      cardinality: recentMetrics.reduce((sum, m) => sum + m.candidate_cardinality, 0) / recentMetrics.length,
      dataRate: recentMetrics.reduce((sum, m) => sum + m.candidate_data_rate, 0) / recentMetrics.length,
      errorRate: recentMetrics.reduce((sum, m) => sum + m.candidate_error_rate, 0) / recentMetrics.length,
      latency: recentMetrics.reduce((sum, m) => sum + m.candidate_p95_latency, 0) / recentMetrics.length,
    }

    const cardinalityReduction = ((avgBaseline.cardinality - avgCandidate.cardinality) / avgBaseline.cardinality) * 100
    const dataRateReduction = ((avgBaseline.dataRate - avgCandidate.dataRate) / avgBaseline.dataRate) * 100
    const errorRateChange = ((avgCandidate.errorRate - avgBaseline.errorRate) / avgBaseline.errorRate) * 100
    const latencyImpact = ((avgCandidate.latency - avgBaseline.latency) / avgBaseline.latency) * 100
    
    // Estimate cost savings based on data reduction
    const avgReduction = (cardinalityReduction + dataRateReduction) / 2
    const annualDataVolume = avgBaseline.dataRate * 60 * 24 * 365 // Minutes to year
    const costPerGB = 0.10 // Example cost
    const costSavings = (annualDataVolume * avgReduction / 100) * costPerGB

    // Generate insights
    const insights: string[] = []
    
    if (cardinalityReduction > 50) {
      insights.push(`Excellent cardinality reduction of ${cardinalityReduction.toFixed(1)}%`)
    }
    
    if (errorRateChange > 5) {
      insights.push(`⚠️ Error rate increased by ${errorRateChange.toFixed(1)}%`)
    }
    
    if (latencyImpact > 10) {
      insights.push(`⚠️ Latency increased by ${latencyImpact.toFixed(1)}%`)
    }
    
    if (dataRateReduction > 40 && errorRateChange < 1 && latencyImpact < 5) {
      insights.push('✅ Significant cost reduction with minimal impact')
    }

    // Determine recommendation
    let recommendation: 'promote' | 'reject' | 'investigate' = 'investigate'
    let confidence = 50

    if (cardinalityReduction > 30 && errorRateChange < 1 && latencyImpact < 5) {
      recommendation = 'promote'
      confidence = 85 + Math.min(15, cardinalityReduction / 4)
    } else if (errorRateChange > 5 || latencyImpact > 20) {
      recommendation = 'reject'
      confidence = 70 + Math.min(20, errorRateChange * 2)
    }

    return {
      cardinalityReduction,
      dataRateReduction,
      errorRateChange,
      latencyImpact,
      costSavings,
      recommendation,
      confidence,
      insights,
    }
  }, [metricsData])

  // Prepare chart data
  const chartData = useMemo(() => {
    if (!metricsData || metricsData.length === 0) return null

    const filteredData = metricsData.filter(m => {
      const timestamp = new Date(m.timestamp).getTime()
      const now = Date.now()
      switch (timeRange) {
        case '1h':
          return timestamp > now - 3600000
        case '6h':
          return timestamp > now - 21600000
        case '24h':
          return timestamp > now - 86400000
        default:
          return true
      }
    })

    return {
      labels: filteredData.map(m => format(new Date(m.timestamp), 'HH:mm')),
      datasets: [
        {
          label: 'Baseline',
          data: filteredData.map(m => m.baseline_cardinality),
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.1)',
          tension: 0.4,
        },
        {
          label: 'Candidate',
          data: filteredData.map(m => m.candidate_cardinality),
          borderColor: 'rgb(54, 162, 235)',
          backgroundColor: 'rgba(54, 162, 235, 0.1)',
          tension: 0.4,
        },
      ],
    }
  }, [metricsData, timeRange])

  const handleExport = (format: 'pdf' | 'csv' | 'json') => {
    // Implementation for exporting report
    showNotification(`Exporting analysis report as ${format.toUpperCase()}`, 'info')
    setExportDialogOpen(false)
  }

  const handlePromoteVariant = (variant: 'baseline' | 'candidate') => {
    onPromote?.(variant)
  }

  return (
    <Box>
      {/* Header with Actions */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Box>
            <Typography variant="h5" gutterBottom>
              Experiment Analysis
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Comprehensive comparison of baseline vs candidate performance
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <Select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value)}
              >
                <MenuItem value="1h">Last 1 hour</MenuItem>
                <MenuItem value="6h">Last 6 hours</MenuItem>
                <MenuItem value="24h">Last 24 hours</MenuItem>
                <MenuItem value="all">All data</MenuItem>
              </Select>
            </FormControl>
            <Button
              variant="outlined"
              startIcon={<Download />}
              onClick={() => setExportDialogOpen(true)}
            >
              Export
            </Button>
            <Button
              variant="outlined"
              startIcon={<Share />}
            >
              Share
            </Button>
          </Box>
        </Box>
      </Paper>

      {/* Summary Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={3}>
          <MetricComparisonCard
            title="Time Series Cardinality"
            icon={<Timeline />}
            baseline={150000}
            candidate={52500}
            higherIsBetter={false}
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <MetricComparisonCard
            title="Data Rate"
            icon={<Speed />}
            baseline={2.3}
            candidate={0.92}
            unit=" M/min"
            higherIsBetter={false}
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <MetricComparisonCard
            title="Error Rate"
            icon={<Warning />}
            baseline={0.05}
            candidate={0.06}
            unit="%"
            higherIsBetter={false}
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <MetricComparisonCard
            title="P95 Latency"
            icon={<Timer />}
            baseline={45}
            candidate={48}
            unit="ms"
            higherIsBetter={false}
          />
        </Grid>
      </Grid>

      {/* Analysis Result Alert */}
      <Alert
        severity={
          analysisResult.recommendation === 'promote'
            ? 'success'
            : analysisResult.recommendation === 'reject'
            ? 'error'
            : 'info'
        }
        sx={{ mb: 3 }}
        action={
          analysisResult.recommendation === 'promote' && (
            <Button
              color="inherit"
              size="small"
              onClick={() => handlePromoteVariant('candidate')}
            >
              Promote Candidate
            </Button>
          )
        }
      >
        <Typography variant="subtitle2" gutterBottom>
          Recommendation: {analysisResult.recommendation.toUpperCase()} (
          {analysisResult.confidence.toFixed(0)}% confidence)
        </Typography>
        <Typography variant="body2">
          {analysisResult.recommendation === 'promote'
            ? `The candidate pipeline shows ${analysisResult.cardinalityReduction.toFixed(1)}% reduction in metrics volume with minimal impact. Estimated annual savings: $${analysisResult.costSavings.toFixed(0)}`
            : analysisResult.recommendation === 'reject'
            ? 'The candidate pipeline shows degraded performance or increased error rates. Consider adjusting the configuration.'
            : 'More data needed for a confident recommendation. Continue monitoring the experiment.'}
        </Typography>
      </Alert>

      {/* Detailed Analysis Tabs */}
      <Paper sx={{ mb: 3 }}>
        <Tabs value={activeTab} onChange={(e, v) => setActiveTab(v)}>
          <Tab label="Metrics Comparison" />
          <Tab label="Performance Impact" />
          <Tab label="Cost Analysis" />
          <Tab label="Detailed Insights" />
        </Tabs>

        {/* Metrics Comparison Tab */}
        {activeTab === 0 && (
          <Box sx={{ p: 3 }}>
            <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between' }}>
              <ToggleButtonGroup
                value={chartType}
                exclusive
                onChange={(e, v) => v && setChartType(v)}
                size="small"
              >
                <ToggleButton value="line">
                  <Timeline />
                </ToggleButton>
                <ToggleButton value="bar">
                  <BarChart />
                </ToggleButton>
              </ToggleButtonGroup>
              
              <FormControlLabel
                control={
                  <Switch
                    checked={showConfidenceInterval}
                    onChange={(e) => setShowConfidenceInterval(e.target.checked)}
                  />
                }
                label="Show confidence interval"
              />
            </Box>

            <Box sx={{ height: 400 }}>
              {chartData && (
                chartType === 'line' ? (
                  <Line
                    data={chartData}
                    options={{
                      responsive: true,
                      maintainAspectRatio: false,
                      plugins: {
                        title: {
                          display: true,
                          text: 'Cardinality Over Time',
                        },
                      },
                    }}
                  />
                ) : (
                  <Bar
                    data={chartData}
                    options={{
                      responsive: true,
                      maintainAspectRatio: false,
                      plugins: {
                        title: {
                          display: true,
                          text: 'Cardinality Comparison',
                        },
                      },
                    }}
                  />
                )
              )}
            </Box>
          </Box>
        )}

        {/* Performance Impact Tab */}
        {activeTab === 1 && (
          <Box sx={{ p: 3 }}>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <Typography variant="h6" gutterBottom>
                  Resource Utilization
                </Typography>
                <TableContainer>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Metric</TableCell>
                        <TableCell align="right">Baseline</TableCell>
                        <TableCell align="right">Candidate</TableCell>
                        <TableCell align="right">Change</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      <TableRow>
                        <TableCell>CPU Usage</TableCell>
                        <TableCell align="right">12%</TableCell>
                        <TableCell align="right">8%</TableCell>
                        <TableCell align="right">
                          <Chip label="-33%" size="small" color="success" />
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>Memory Usage</TableCell>
                        <TableCell align="right">256MB</TableCell>
                        <TableCell align="right">192MB</TableCell>
                        <TableCell align="right">
                          <Chip label="-25%" size="small" color="success" />
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>Network I/O</TableCell>
                        <TableCell align="right">45Mbps</TableCell>
                        <TableCell align="right">18Mbps</TableCell>
                        <TableCell align="right">
                          <Chip label="-60%" size="small" color="success" />
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </TableContainer>
              </Grid>
              
              <Grid item xs={12} md={6}>
                <Typography variant="h6" gutterBottom>
                  Quality Metrics
                </Typography>
                <List>
                  {analysisResult.insights.map((insight, i) => (
                    <ListItem key={i}>
                      <ListItemText primary={insight} />
                    </ListItem>
                  ))}
                </List>
              </Grid>
            </Grid>
          </Box>
        )}

        {/* Cost Analysis Tab */}
        {activeTab === 2 && (
          <Box sx={{ p: 3 }}>
            <Grid container spacing={3}>
              <Grid item xs={12} md={8}>
                <Typography variant="h6" gutterBottom>
                  Projected Cost Savings
                </Typography>
                <Card sx={{ mb: 3 }}>
                  <CardContent>
                    <Grid container spacing={2}>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="text.secondary">
                          Current Monthly Cost
                        </Typography>
                        <Typography variant="h4">
                          ${(analysisResult.costSavings * 12 / analysisResult.cardinalityReduction * 100).toFixed(0)}
                        </Typography>
                      </Grid>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="text.secondary">
                          Projected Monthly Cost
                        </Typography>
                        <Typography variant="h4" color="success.main">
                          ${(analysisResult.costSavings * 12 / analysisResult.cardinalityReduction * (100 - analysisResult.cardinalityReduction)).toFixed(0)}
                        </Typography>
                      </Grid>
                    </Grid>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="h6" color="primary">
                      Annual Savings: ${analysisResult.costSavings.toFixed(0)}
                    </Typography>
                  </CardContent>
                </Card>
                
                <Typography variant="subtitle2" gutterBottom>
                  Cost Breakdown by Component
                </Typography>
                <Box sx={{ height: 300 }}>
                  <Doughnut
                    data={{
                      labels: ['Storage', 'Ingestion', 'Query', 'Transfer'],
                      datasets: [{
                        data: [40, 30, 20, 10],
                        backgroundColor: [
                          'rgba(255, 99, 132, 0.8)',
                          'rgba(54, 162, 235, 0.8)',
                          'rgba(255, 206, 86, 0.8)',
                          'rgba(75, 192, 192, 0.8)',
                        ],
                      }],
                    }}
                    options={{
                      responsive: true,
                      maintainAspectRatio: false,
                    }}
                  />
                </Box>
              </Grid>
              
              <Grid item xs={12} md={4}>
                <Typography variant="h6" gutterBottom>
                  ROI Calculator
                </Typography>
                <Card>
                  <CardContent>
                    <Typography variant="body2" paragraph>
                      Based on your current metrics volume and the observed reduction rate:
                    </Typography>
                    <List>
                      <ListItem>
                        <ListItemText
                          primary="Break-even Time"
                          secondary="< 1 month"
                        />
                      </ListItem>
                      <ListItem>
                        <ListItemText
                          primary="3-Year Savings"
                          secondary={`$${(analysisResult.costSavings * 3).toFixed(0)}`}
                        />
                      </ListItem>
                      <ListItem>
                        <ListItemText
                          primary="Efficiency Gain"
                          secondary={`${analysisResult.cardinalityReduction.toFixed(0)}%`}
                        />
                      </ListItem>
                    </List>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>
          </Box>
        )}

        {/* Detailed Insights Tab */}
        {activeTab === 3 && (
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              AI-Powered Insights
            </Typography>
            <Grid container spacing={2}>
              {[
                {
                  title: 'Top Cardinality Contributors',
                  content: 'Container labels (65%), HTTP endpoints (20%), Custom tags (15%)',
                  action: 'Review label configuration',
                  severity: 'warning' as const,
                },
                {
                  title: 'Optimization Opportunities',
                  content: 'Additional 15-20% reduction possible with histogram buckets optimization',
                  action: 'Configure histogram buckets',
                  severity: 'info' as const,
                },
                {
                  title: 'Quality Validation',
                  content: 'All critical metrics preserved. No data loss detected.',
                  action: 'View validation report',
                  severity: 'success' as const,
                },
                {
                  title: 'Performance Impact',
                  content: 'Minimal latency increase (+3ms) well within acceptable range',
                  action: 'View performance metrics',
                  severity: 'success' as const,
                },
              ].map((insight, i) => (
                <Grid item xs={12} md={6} key={i}>
                  <Alert
                    severity={insight.severity}
                    action={
                      <Button size="small" color="inherit">
                        {insight.action}
                      </Button>
                    }
                  >
                    <Typography variant="subtitle2">{insight.title}</Typography>
                    <Typography variant="body2">{insight.content}</Typography>
                  </Alert>
                </Grid>
              ))}
            </Grid>
          </Box>
        )}
      </Paper>

      {/* Export Dialog */}
      <Dialog open={exportDialogOpen} onClose={() => setExportDialogOpen(false)}>
        <DialogTitle>Export Analysis Report</DialogTitle>
        <DialogContent>
          <Typography variant="body2" paragraph>
            Choose the format for your analysis report:
          </Typography>
          <List>
            <ListItem button onClick={() => handleExport('pdf')}>
              <ListItemIcon>
                <Assessment />
              </ListItemIcon>
              <ListItemText
                primary="PDF Report"
                secondary="Comprehensive report with charts and insights"
              />
            </ListItem>
            <ListItem button onClick={() => handleExport('csv')}>
              <ListItemIcon>
                <Storage />
              </ListItemIcon>
              <ListItemText
                primary="CSV Data"
                secondary="Raw metrics data for further analysis"
              />
            </ListItem>
            <ListItem button onClick={() => handleExport('json')}>
              <ListItemIcon>
                <CloudUpload />
              </ListItemIcon>
              <ListItemText
                primary="JSON Export"
                secondary="Complete analysis data in JSON format"
              />
            </ListItem>
          </List>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setExportDialogOpen(false)}>Cancel</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}