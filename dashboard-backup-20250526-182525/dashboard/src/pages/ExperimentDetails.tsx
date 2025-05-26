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
  Chip,
  IconButton,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Alert,
  Skeleton,
  Tooltip,
  LinearProgress,
} from '@mui/material'
import {
  ArrowBack,
  PlayArrow,
  Stop,
  Delete,
  Assessment,
  Settings,
  Timeline,
  Computer,
  Schedule,
  CheckCircle,
  Error,
  Warning,
  CloudUpload,
  Visibility,
} from '@mui/icons-material'
import { useAppSelector, useAppDispatch } from '@hooks/redux'
import {
  fetchExperimentById,
  startExperiment,
  stopExperiment,
  deleteExperiment,
  promoteVariant,
} from '@store/slices/experimentSlice'
import { format, formatDistanceToNow } from 'date-fns'
import { PipelineViewer } from '../components/PipelineBuilder/PipelineViewer'
import { ExperimentMonitor } from '../components/ExperimentMonitor'
import { useExperimentUpdates } from '../hooks/useExperimentUpdates'

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
      id={`experiment-tabpanel-${index}`}
      aria-labelledby={`experiment-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
    </div>
  )
}

export const ExperimentDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { currentExperiment: experiment, loading, error } = useAppSelector(
    (state) => state.experiments
  )

  const [activeTab, setActiveTab] = useState(0)
  const [promotingVariant, setPromotingVariant] = useState<string | null>(null)
  const [monitorOpen, setMonitorOpen] = useState(false)

  // Subscribe to real-time updates for this experiment
  useExperimentUpdates(id)

  useEffect(() => {
    if (id) {
      dispatch(fetchExperimentById(id))
    }
  }, [id, dispatch])

  const handleBack = () => {
    navigate('/experiments')
  }

  const handleStart = async () => {
    if (id) {
      await dispatch(startExperiment(id))
    }
  }

  const handleStop = async () => {
    if (id) {
      await dispatch(stopExperiment(id))
    }
  }

  const handleDelete = async () => {
    if (id && window.confirm('Are you sure you want to delete this experiment?')) {
      await dispatch(deleteExperiment(id))
      navigate('/experiments')
    }
  }

  const handlePromote = async (variant: 'baseline' | 'candidate') => {
    if (id) {
      setPromotingVariant(variant)
      try {
        await dispatch(promoteVariant({ id, variant }))
        setPromotingVariant(null)
      } catch (error) {
        setPromotingVariant(null)
      }
    }
  }

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue)
  }

  if (loading && !experiment) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4 }}>
        <Skeleton variant="rectangular" height={200} sx={{ mb: 3 }} />
        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Skeleton variant="rectangular" height={400} />
          </Grid>
          <Grid item xs={12} md={4}>
            <Skeleton variant="rectangular" height={400} />
          </Grid>
        </Grid>
      </Container>
    )
  }

  if (error) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4 }}>
        <Alert severity="error">{error}</Alert>
        <Button onClick={handleBack} sx={{ mt: 2 }}>
          Back to Experiments
        </Button>
      </Container>
    )
  }

  if (!experiment) {
    return null
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <Schedule color="primary" />
      case 'completed':
        return <CheckCircle color="success" />
      case 'failed':
        return <Error color="error" />
      case 'cancelled':
        return <Warning color="warning" />
      default:
        return <Schedule color="disabled" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'primary'
      case 'completed':
        return 'success'
      case 'failed':
        return 'error'
      case 'cancelled':
        return 'warning'
      default:
        return 'default'
    }
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4 }}>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBack />}
          onClick={handleBack}
          sx={{ mb: 2 }}
        >
          Back to Experiments
        </Button>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box>
            <Typography variant="h4" component="h1" gutterBottom>
              {experiment.name}
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Chip
                icon={getStatusIcon(experiment.status)}
                label={experiment.status.toUpperCase()}
                color={getStatusColor(experiment.status)}
              />
              <Typography variant="body2" color="text.secondary">
                Created {formatDistanceToNow(new Date(experiment.createdAt))} ago
              </Typography>
            </Box>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            {experiment.status === 'pending' && (
              <Button
                variant="contained"
                startIcon={<PlayArrow />}
                onClick={handleStart}
                disabled={loading}
              >
                Start
              </Button>
            )}
            {experiment.status === 'running' && (
              <>
                <Button
                  variant="contained"
                  startIcon={<Visibility />}
                  onClick={() => setMonitorOpen(true)}
                >
                  Monitor Live
                </Button>
                <Button
                  variant="contained"
                  color="warning"
                  startIcon={<Stop />}
                  onClick={handleStop}
                  disabled={loading}
                >
                  Stop
                </Button>
              </>
            )}
            {['completed', 'failed'].includes(experiment.status) && (
              <Button
                variant="contained"
                startIcon={<Assessment />}
                onClick={() => navigate(`/experiments/${id}/analysis`)}
              >
                View Analysis
              </Button>
            )}
            <IconButton
              color="error"
              onClick={handleDelete}
              disabled={experiment.status === 'running' || loading}
            >
              <Delete />
            </IconButton>
          </Box>
        </Box>
      </Box>

      {experiment.status === 'running' && (
        <LinearProgress sx={{ mb: 3 }} />
      )}

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ mb: 3 }}>
            <Tabs value={activeTab} onChange={handleTabChange}>
              <Tab label="Overview" />
              <Tab label="Pipeline Configuration" />
              <Tab label="Target Hosts" />
              <Tab label="Events" />
              {experiment.status === 'running' && <Tab label="Real-time Monitor" />}
            </Tabs>
            <Divider />
            
            <TabPanel value={activeTab} index={0}>
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Experiment Details
                </Typography>
                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <Typography variant="body2" color="text.secondary">
                      Description
                    </Typography>
                    <Typography variant="body1">
                      {experiment.description || 'No description provided'}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      Duration
                    </Typography>
                    <Typography variant="body1">
                      {experiment.spec?.duration || 'Not specified'}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      Load Profile
                    </Typography>
                    <Typography variant="body1">
                      {experiment.spec?.loadProfile || 'realistic'}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      Started At
                    </Typography>
                    <Typography variant="body1">
                      {experiment.startedAt
                        ? format(new Date(experiment.startedAt), 'PPp')
                        : 'Not started'}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      Completed At
                    </Typography>
                    <Typography variant="body1">
                      {experiment.completedAt
                        ? format(new Date(experiment.completedAt), 'PPp')
                        : 'Not completed'}
                    </Typography>
                  </Grid>
                </Grid>
              </Box>
            </TabPanel>

            <TabPanel value={activeTab} index={1}>
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Pipeline Configuration
                </Typography>
                {experiment.spec?.baseline && (
                  <Box sx={{ mb: 3 }}>
                    <Typography variant="subtitle1" gutterBottom>
                      Baseline (Control)
                    </Typography>
                    <PipelineViewer pipeline={experiment.spec.baseline} />
                  </Box>
                )}
                {experiment.spec?.candidate && (
                  <Box>
                    <Typography variant="subtitle1" gutterBottom>
                      Candidate (Test)
                    </Typography>
                    <PipelineViewer pipeline={experiment.spec.candidate} />
                  </Box>
                )}
              </Box>
            </TabPanel>

            <TabPanel value={activeTab} index={2}>
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Target Hosts ({experiment.spec?.targetHosts?.length || 0})
                </Typography>
                <List>
                  {experiment.spec?.targetHosts?.map((host, index) => (
                    <ListItem key={index}>
                      <ListItemIcon>
                        <Computer />
                      </ListItemIcon>
                      <ListItemText
                        primary={host}
                        secondary={`Host ${index + 1}`}
                      />
                    </ListItem>
                  )) || (
                    <ListItem>
                      <ListItemText
                        primary="No target hosts specified"
                        secondary="Experiment will run on all available hosts"
                      />
                    </ListItem>
                  )}
                </List>
              </Box>
            </TabPanel>

            <TabPanel value={activeTab} index={3}>
              <Box sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Experiment Events
                </Typography>
                <List>
                  <ListItem>
                    <ListItemIcon>
                      <Schedule />
                    </ListItemIcon>
                    <ListItemText
                      primary="Experiment created"
                      secondary={format(new Date(experiment.createdAt), 'PPp')}
                    />
                  </ListItem>
                  {experiment.startedAt && (
                    <ListItem>
                      <ListItemIcon>
                        <PlayArrow color="primary" />
                      </ListItemIcon>
                      <ListItemText
                        primary="Experiment started"
                        secondary={format(new Date(experiment.startedAt), 'PPp')}
                      />
                    </ListItem>
                  )}
                  {experiment.completedAt && (
                    <ListItem>
                      <ListItemIcon>
                        <CheckCircle color="success" />
                      </ListItemIcon>
                      <ListItemText
                        primary="Experiment completed"
                        secondary={format(new Date(experiment.completedAt), 'PPp')}
                      />
                    </ListItem>
                  )}
                </List>
              </Box>
            </TabPanel>
            
            {experiment.status === 'running' && (
              <TabPanel value={activeTab} index={4}>
                <Box sx={{ p: 3 }}>
                  <ExperimentMonitor experimentId={id!} embedded />
                </Box>
              </TabPanel>
            )}
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Actions
              </Typography>
              <List>
                {experiment.status === 'completed' && (
                  <>
                    <ListItem>
                      <Button
                        fullWidth
                        variant="outlined"
                        color="primary"
                        startIcon={<CloudUpload />}
                        onClick={() => handlePromote('baseline')}
                        disabled={promotingVariant !== null}
                      >
                        {promotingVariant === 'baseline' ? 'Promoting...' : 'Promote Baseline'}
                      </Button>
                    </ListItem>
                    <ListItem>
                      <Button
                        fullWidth
                        variant="outlined"
                        color="secondary"
                        startIcon={<CloudUpload />}
                        onClick={() => handlePromote('candidate')}
                        disabled={promotingVariant !== null}
                      >
                        {promotingVariant === 'candidate' ? 'Promoting...' : 'Promote Candidate'}
                      </Button>
                    </ListItem>
                  </>
                )}
                <ListItem>
                  <Button
                    fullWidth
                    variant="outlined"
                    startIcon={<Assessment />}
                    onClick={() => navigate(`/experiments/${id}/analysis`)}
                    disabled={!['running', 'completed'].includes(experiment.status)}
                  >
                    View Metrics
                  </Button>
                </ListItem>
                <ListItem>
                  <Button
                    fullWidth
                    variant="outlined"
                    color="error"
                    startIcon={<Delete />}
                    onClick={handleDelete}
                    disabled={experiment.status === 'running'}
                  >
                    Delete Experiment
                  </Button>
                </ListItem>
              </List>
            </CardContent>
          </Card>

          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Metadata
              </Typography>
              <List dense>
                <ListItem>
                  <ListItemText
                    primary="ID"
                    secondary={experiment.id}
                    secondaryTypographyProps={{ sx: { wordBreak: 'break-all' } }}
                  />
                </ListItem>
                <ListItem>
                  <ListItemText
                    primary="Owner"
                    secondary={experiment.owner || 'Unknown'}
                  />
                </ListItem>
                <ListItem>
                  <ListItemText
                    primary="Created"
                    secondary={format(new Date(experiment.createdAt), 'PPp')}
                  />
                </ListItem>
                <ListItem>
                  <ListItemText
                    primary="Last Updated"
                    secondary={format(new Date(experiment.updatedAt), 'PPp')}
                  />
                </ListItem>
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
      
      {/* Real-time Monitor Dialog */}
      {monitorOpen && (
        <ExperimentMonitor
          experimentId={id!}
          onClose={() => setMonitorOpen(false)}
        />
      )}
    </Container>
  )
}