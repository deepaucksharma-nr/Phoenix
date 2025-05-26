import React, { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Container,
  Paper,
  Typography,
  Box,
  Button,
  Grid,
  Card,
  CardContent,
  CardActions,
  Chip,
  IconButton,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Tooltip,
  Skeleton,
  Alert,
  TablePagination,
  ToggleButton,
  ToggleButtonGroup,
  Menu,
  ListItemIcon,
  ListItemText,
  Divider,
} from '@mui/material'
import {
  Add,
  Science,
  PlayArrow,
  Stop,
  Delete,
  Edit,
  Refresh,
  FilterList,
  Search,
  Assessment,
  Visibility,
  VisibilityOff,
  GetApp,
  Description,
  Code,
  TableChart,
} from '@mui/icons-material'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import { fetchExperiments, deleteExperiment, updateExperimentStatus } from '@store/slices/experimentSlice'
import { ExperimentWizard } from '../components/ExperimentWizard'
import { formatDistanceToNow } from 'date-fns'

const STATUS_COLORS = {
  pending: 'default',
  running: 'primary',
  completed: 'success',
  failed: 'error',
  cancelled: 'warning',
  initializing: 'info',
  stopping: 'warning',
  stopped: 'default',
} as const

const PRIORITY_COLORS = {
  low: 'default',
  medium: 'warning',
  high: 'error',
} as const

export const Experiments: React.FC = () => {
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { experiments, loading, error } = useAppSelector(state => state.experiments)

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(10)
  const [showFilters, setShowFilters] = useState(false)
  const [wizardOpen, setWizardOpen] = useState(false)
  const [watchMode, setWatchMode] = useState(false)
  const [exportMenuAnchor, setExportMenuAnchor] = useState<null | HTMLElement>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    dispatch(fetchExperiments())
  }, [dispatch])

  // Watch mode effect
  useEffect(() => {
    if (watchMode) {
      intervalRef.current = setInterval(() => {
        dispatch(fetchExperiments())
      }, 5000) // Refresh every 5 seconds
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [watchMode, dispatch])

  const handleCreateExperiment = () => {
    setWizardOpen(true)
  }

  const handleViewExperiment = (id: string) => {
    navigate(`/experiments/${id}`)
  }

  const handleAnalyzeExperiment = (id: string) => {
    navigate(`/experiments/${id}/analysis`)
  }

  const handleStartExperiment = async (id: string) => {
    try {
      await dispatch(updateExperimentStatus({ id, status: 'running' })).unwrap()
    } catch (error) {
      console.error('Failed to start experiment:', error)
    }
  }

  const handleStopExperiment = async (id: string) => {
    try {
      await dispatch(updateExperimentStatus({ id, status: 'stopped' })).unwrap()
    } catch (error) {
      console.error('Failed to stop experiment:', error)
    }
  }

  const handleDeleteExperiment = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this experiment?')) {
      try {
        await dispatch(deleteExperiment(id)).unwrap()
      } catch (error) {
        console.error('Failed to delete experiment:', error)
      }
    }
  }

  const filteredExperiments = experiments.filter((exp) => {
    const matchesSearch =
      exp.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      exp.description?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === 'all' || exp.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const paginatedExperiments = filteredExperiments.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  )

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage)
  }

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10))
    setPage(0)
  }

  const handleExportClick = (event: React.MouseEvent<HTMLElement>) => {
    setExportMenuAnchor(event.currentTarget)
  }

  const handleExportClose = () => {
    setExportMenuAnchor(null)
  }

  const exportToJSON = () => {
    const dataStr = JSON.stringify(filteredExperiments, null, 2)
    const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr)
    
    const exportFileDefaultName = `experiments_${new Date().toISOString().split('T')[0]}.json`
    
    const linkElement = document.createElement('a')
    linkElement.setAttribute('href', dataUri)
    linkElement.setAttribute('download', exportFileDefaultName)
    linkElement.click()
    handleExportClose()
  }

  const exportToCSV = () => {
    const headers = ['ID', 'Name', 'Description', 'Status', 'Created At', 'Duration', 'Target Hosts']
    const rows = filteredExperiments.map(exp => [
      exp.id,
      exp.name,
      exp.description || '',
      exp.status,
      new Date(exp.createdAt).toISOString(),
      exp.spec?.duration || '',
      exp.spec?.targetHosts?.length || 0
    ])
    
    const csvContent = [headers, ...rows]
      .map(row => row.map(cell => `"${cell}"`).join(','))
      .join('\n')
    
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `experiments_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    handleExportClose()
  }

  const exportToYAML = () => {
    // Simple YAML export (in production, use a proper YAML library)
    const yamlContent = filteredExperiments.map(exp => {
      return `- id: ${exp.id}
  name: ${exp.name}
  description: ${exp.description || ''}
  status: ${exp.status}
  created_at: ${exp.createdAt}
  spec:
    duration: ${exp.spec?.duration || ''}
    target_hosts: ${exp.spec?.targetHosts?.length || 0}`
    }).join('\n\n')
    
    const blob = new Blob([yamlContent], { type: 'text/yaml;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `experiments_${new Date().toISOString().split('T')[0]}.yaml`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    handleExportClose()
  }

  if (loading && experiments.length === 0) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4 }}>
        <Grid container spacing={3}>
          {[1, 2, 3, 4].map((i) => (
            <Grid item xs={12} sm={6} md={4} key={i}>
              <Skeleton variant="rectangular" height={200} />
            </Grid>
          ))}
        </Grid>
      </Container>
    )
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Experiments
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Manage and monitor your A/B testing experiments
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <TextField
            placeholder="Search experiments..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: <Search sx={{ mr: 1, color: 'action.active' }} />,
            }}
            sx={{ flexGrow: 1 }}
          />
          <FormControl sx={{ minWidth: 150 }}>
            <InputLabel>Status</InputLabel>
            <Select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              label="Status"
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="pending">Pending</MenuItem>
              <MenuItem value="running">Running</MenuItem>
              <MenuItem value="completed">Completed</MenuItem>
              <MenuItem value="failed">Failed</MenuItem>
              <MenuItem value="cancelled">Cancelled</MenuItem>
            </Select>
          </FormControl>
          <Tooltip title="Toggle filters">
            <IconButton onClick={() => setShowFilters(!showFilters)}>
              <FilterList />
            </IconButton>
          </Tooltip>
          <Tooltip title="Refresh">
            <IconButton onClick={() => dispatch(fetchExperiments())}>
              <Refresh />
            </IconButton>
          </Tooltip>
          <ToggleButtonGroup
            value={watchMode}
            exclusive
            onChange={(_, newValue) => setWatchMode(newValue === true)}
            size="small"
          >
            <ToggleButton value={true} aria-label="watch mode">
              <Tooltip title={watchMode ? "Watch mode on (5s refresh)" : "Enable watch mode"}>
                {watchMode ? <Visibility /> : <VisibilityOff />}
              </Tooltip>
            </ToggleButton>
          </ToggleButtonGroup>
          <Button
            variant="outlined"
            startIcon={<GetApp />}
            onClick={handleExportClick}
          >
            Export
          </Button>
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={handleCreateExperiment}
          >
            New Experiment
          </Button>
        </Box>
        {watchMode && (
          <Alert severity="info" sx={{ mt: 2 }}>
            Watch mode enabled - experiments will refresh every 5 seconds
          </Alert>
        )}
      </Paper>

      {/* Export Menu */}
      <Menu
        anchorEl={exportMenuAnchor}
        open={Boolean(exportMenuAnchor)}
        onClose={handleExportClose}
      >
        <MenuItem onClick={exportToJSON}>
          <ListItemIcon>
            <Code fontSize="small" />
          </ListItemIcon>
          <ListItemText>Export as JSON</ListItemText>
        </MenuItem>
        <MenuItem onClick={exportToCSV}>
          <ListItemIcon>
            <TableChart fontSize="small" />
          </ListItemIcon>
          <ListItemText>Export as CSV</ListItemText>
        </MenuItem>
        <MenuItem onClick={exportToYAML}>
          <ListItemIcon>
            <Description fontSize="small" />
          </ListItemIcon>
          <ListItemText>Export as YAML</ListItemText>
        </MenuItem>
      </Menu>

      <Grid container spacing={3}>
        {paginatedExperiments.map((experiment) => (
          <Grid item xs={12} sm={6} md={4} key={experiment.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Science sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="h6" component="h2" noWrap>
                    {experiment.name}
                  </Typography>
                </Box>
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ mb: 2, minHeight: 40 }}
                >
                  {experiment.description || 'No description'}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mb: 2, flexWrap: 'wrap' }}>
                  <Chip
                    label={experiment.status}
                    size="small"
                    color={STATUS_COLORS[experiment.status] || 'default'}
                  />
                  {experiment.spec?.targetHosts && (
                    <Chip
                      label={`${experiment.spec.targetHosts.length} hosts`}
                      size="small"
                      variant="outlined"
                    />
                  )}
                  {experiment.spec?.duration && (
                    <Chip
                      label={experiment.spec.duration}
                      size="small"
                      variant="outlined"
                    />
                  )}
                </Box>
                <Typography variant="caption" color="text.secondary">
                  Created {formatDistanceToNow(new Date(experiment.createdAt))} ago
                </Typography>
              </CardContent>
              <CardActions sx={{ justifyContent: 'space-between' }}>
                <Box>
                  {experiment.status === 'pending' && (
                    <Tooltip title="Start experiment">
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => handleStartExperiment(experiment.id)}
                      >
                        <PlayArrow />
                      </IconButton>
                    </Tooltip>
                  )}
                  {experiment.status === 'running' && (
                    <Tooltip title="Stop experiment">
                      <IconButton
                        size="small"
                        color="warning"
                        onClick={() => handleStopExperiment(experiment.id)}
                      >
                        <Stop />
                      </IconButton>
                    </Tooltip>
                  )}
                  {['completed', 'failed'].includes(experiment.status) && (
                    <Tooltip title="View analysis">
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => handleAnalyzeExperiment(experiment.id)}
                      >
                        <Assessment />
                      </IconButton>
                    </Tooltip>
                  )}
                </Box>
                <Box>
                  <Tooltip title="View details">
                    <IconButton
                      size="small"
                      onClick={() => handleViewExperiment(experiment.id)}
                    >
                      <Edit />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete experiment">
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() => handleDeleteExperiment(experiment.id)}
                      disabled={experiment.status === 'running'}
                    >
                      <Delete />
                    </IconButton>
                  </Tooltip>
                </Box>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {filteredExperiments.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Science sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No experiments found
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            {searchTerm || statusFilter !== 'all'
              ? 'Try adjusting your filters'
              : 'Create your first experiment to get started'}
          </Typography>
          {searchTerm === '' && statusFilter === 'all' && (
            <Button
              variant="contained"
              startIcon={<Add />}
              onClick={handleCreateExperiment}
            >
              Create Experiment
            </Button>
          )}
        </Box>
      )}

      <TablePagination
        component="div"
        count={filteredExperiments.length}
        page={page}
        onPageChange={handleChangePage}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        sx={{ mt: 3 }}
      />
      
      {/* Experiment Creation Wizard */}
      <ExperimentWizard
        open={wizardOpen}
        onClose={() => setWizardOpen(false)}
      />
    </Container>
  )
}