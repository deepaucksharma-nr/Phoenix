import React, { useEffect, useState } from 'react'
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
} from '@mui/icons-material'
import { useExperimentStore } from '../store/useExperimentStore'
import { ExperimentWizard } from '../components/ExperimentWizard'
import { formatDistanceToNow } from 'date-fns'

const STATUS_COLORS = {
  pending: 'default',
  running: 'primary',
  completed: 'success',
  failed: 'error',
  cancelled: 'warning',
} as const

const PRIORITY_COLORS = {
  low: 'default',
  medium: 'warning',
  high: 'error',
} as const

export const Experiments: React.FC = () => {
  const navigate = useNavigate()
  const {
    experiments,
    loading,
    error,
    fetchExperiments,
    deleteExperiment,
    startExperiment,
    stopExperiment,
  } = useExperimentStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(10)
  const [showFilters, setShowFilters] = useState(false)
  const [wizardOpen, setWizardOpen] = useState(false)

  useEffect(() => {
    fetchExperiments()
  }, [fetchExperiments])

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
      await startExperiment(id)
    } catch (error) {
      console.error('Failed to start experiment:', error)
    }
  }

  const handleStopExperiment = async (id: string) => {
    try {
      await stopExperiment(id)
    } catch (error) {
      console.error('Failed to stop experiment:', error)
    }
  }

  const handleDeleteExperiment = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this experiment?')) {
      try {
        await deleteExperiment(id)
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
            <IconButton onClick={() => fetchExperiments()}>
              <Refresh />
            </IconButton>
          </Tooltip>
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={handleCreateExperiment}
          >
            New Experiment
          </Button>
        </Box>
      </Paper>

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