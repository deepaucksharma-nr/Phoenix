import React, { useState, useEffect } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Stepper,
  Step,
  StepLabel,
  Box,
  Typography,
  IconButton,
  Grid,
  Card,
  CardContent,
  CardActionArea,
  Chip,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Alert,
  LinearProgress,
  Fade,
  Grow,
} from '@mui/material'
import {
  Close as CloseIcon,
  Science as ScienceIcon,
  Timeline as TimelineIcon,
  AttachMoney as MoneyIcon,
  Speed as SpeedIcon,
  CheckCircle as CheckIcon,
  RadioButtonUnchecked as UncheckedIcon,
  ArrowForward as ArrowForwardIcon,
  Lightbulb as LightbulbIcon,
  School as SchoolIcon,
  EmojiObjects as IdeaIcon,
  TrendingDown as TrendingDownIcon,
  AutoAwesome as AutoAwesomeIcon,
} from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store/useAuthStore'

const steps = [
  'Welcome to Phoenix',
  'Understanding the Platform',
  'Your First Experiment',
  'Best Practices',
]

interface WelcomeGuideProps {
  open: boolean
  onClose: () => void
}

export const WelcomeGuide: React.FC<WelcomeGuideProps> = ({ open, onClose }) => {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [activeStep, setActiveStep] = useState(0)
  const [checkedItems, setCheckedItems] = useState<Set<string>>(new Set())

  useEffect(() => {
    // Load saved progress
    const saved = localStorage.getItem('phoenix_onboarding_progress')
    if (saved) {
      setCheckedItems(new Set(JSON.parse(saved)))
    }
  }, [])

  const handleNext = () => {
    setActiveStep((prev) => prev + 1)
  }

  const handleBack = () => {
    setActiveStep((prev) => prev - 1)
  }

  const handleComplete = () => {
    localStorage.setItem('phoenix_onboarding_completed', 'true')
    onClose()
  }

  const handleCheckItem = (item: string) => {
    const newChecked = new Set(checkedItems)
    if (newChecked.has(item)) {
      newChecked.delete(item)
    } else {
      newChecked.add(item)
    }
    setCheckedItems(newChecked)
    localStorage.setItem('phoenix_onboarding_progress', JSON.stringify(Array.from(newChecked)))
  }

  const handleStartExperiment = () => {
    handleComplete()
    // Open experiment wizard directly
    const event = new CustomEvent('openExperimentWizard')
    window.dispatchEvent(event)
  }

  const renderStepContent = (step: number) => {
    switch (step) {
      case 0:
        return (
          <Fade in timeout={500}>
            <Box>
              <Box sx={{ textAlign: 'center', mb: 4 }}>
                <ScienceIcon sx={{ fontSize: 80, color: 'primary.main', mb: 2 }} />
                <Typography variant="h4" gutterBottom>
                  Welcome to Phoenix, {user?.name || 'Explorer'}!
                </Typography>
                <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
                  Your journey to reduce observability costs by up to 80% starts here.
                </Typography>
              </Box>

              <Grid container spacing={3}>
                <Grid item xs={12} md={4}>
                  <Card sx={{ height: '100%', textAlign: 'center' }}>
                    <CardContent>
                      <TrendingDownIcon sx={{ fontSize: 48, color: 'success.main', mb: 1 }} />
                      <Typography variant="h6" gutterBottom>
                        50-80% Cost Reduction
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Intelligently reduce metrics volume without losing critical insights
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Card sx={{ height: '100%', textAlign: 'center' }}>
                    <CardContent>
                      <SpeedIcon sx={{ fontSize: 48, color: 'info.main', mb: 1 }} />
                      <Typography variant="h6" gutterBottom>
                        Real-time A/B Testing
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Compare baseline and optimized pipelines side-by-side
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Card sx={{ height: '100%', textAlign: 'center' }}>
                    <CardContent>
                      <AutoAwesomeIcon sx={{ fontSize: 48, color: 'secondary.main', mb: 1 }} />
                      <Typography variant="h6" gutterBottom>
                        No Service Mesh Required
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Simple dual-collector pattern for easy deployment
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>

              <Alert severity="info" sx={{ mt: 3 }}>
                <Typography variant="subtitle2">
                  This guide will help you create your first experiment in under 5 minutes!
                </Typography>
              </Alert>
            </Box>
          </Fade>
        )

      case 1:
        return (
          <Grow in timeout={500}>
            <Box>
              <Typography variant="h5" gutterBottom sx={{ display: 'flex', alignItems: 'center' }}>
                <SchoolIcon sx={{ mr: 1 }} />
                How Phoenix Works
              </Typography>

              <Box sx={{ mb: 3 }}>
                <Typography variant="body1" paragraph>
                  Phoenix optimizes your OpenTelemetry pipelines through intelligent A/B testing:
                </Typography>

                <List>
                  <ListItem>
                    <ListItemIcon>
                      <Typography variant="h6" color="primary">1</Typography>
                    </ListItemIcon>
                    <ListItemText
                      primary="Deploy Dual Collectors"
                      secondary="Run baseline and candidate pipelines simultaneously on a subset of hosts"
                    />
                  </ListItem>
                  <ListItem>
                    <ListItemIcon>
                      <Typography variant="h6" color="primary">2</Typography>
                    </ListItemIcon>
                    <ListItemText
                      primary="Compare Metrics"
                      secondary="Analyze volume reduction, data quality, and performance impact in real-time"
                    />
                  </ListItem>
                  <ListItem>
                    <ListItemIcon>
                      <Typography variant="h6" color="primary">3</Typography>
                    </ListItemIcon>
                    <ListItemText
                      primary="Validate Results"
                      secondary="Ensure critical metrics are preserved while reducing unnecessary data"
                    />
                  </ListItem>
                  <ListItem>
                    <ListItemIcon>
                      <Typography variant="h6" color="primary">4</Typography>
                    </ListItemIcon>
                    <ListItemText
                      primary="Promote Winners"
                      secondary="Roll out successful configurations to your entire infrastructure"
                    />
                  </ListItem>
                </List>
              </Box>

              <Card sx={{ bgcolor: 'primary.light', p: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'flex-start' }}>
                  <IdeaIcon sx={{ mr: 1, color: 'primary.main' }} />
                  <Box>
                    <Typography variant="subtitle2" gutterBottom>
                      Pro Tip: Start Small
                    </Typography>
                    <Typography variant="body2">
                      Begin with 10% of your hosts to validate optimizations before wider rollout.
                      Phoenix automatically handles traffic splitting and metric collection.
                    </Typography>
                  </Box>
                </Box>
              </Card>
            </Box>
          </Grow>
        )

      case 2:
        return (
          <Box>
            <Typography variant="h5" gutterBottom sx={{ display: 'flex', alignItems: 'center' }}>
              <ScienceIcon sx={{ mr: 1 }} />
              Creating Your First Experiment
            </Typography>

            <Typography variant="body1" paragraph>
              Let's set up your first cost optimization experiment. Check off each item as you go:
            </Typography>

            <List sx={{ mb: 3 }}>
              <ListItem>
                <ListItemIcon>
                  <IconButton
                    size="small"
                    onClick={() => handleCheckItem('name')}
                  >
                    {checkedItems.has('name') ? <CheckIcon color="success" /> : <UncheckedIcon />}
                  </IconButton>
                </ListItemIcon>
                <ListItemText
                  primary="Choose a descriptive name"
                  secondary="e.g., 'Reduce container metrics cardinality'"
                />
              </ListItem>
              <ListItem>
                <ListItemIcon>
                  <IconButton
                    size="small"
                    onClick={() => handleCheckItem('pipeline')}
                  >
                    {checkedItems.has('pipeline') ? <CheckIcon color="success" /> : <UncheckedIcon />}
                  </IconButton>
                </ListItemIcon>
                <ListItemText
                  primary="Select an optimization strategy"
                  secondary="Start with 'Aggregated' for predictable workloads"
                />
              </ListItem>
              <ListItem>
                <ListItemIcon>
                  <IconButton
                    size="small"
                    onClick={() => handleCheckItem('hosts')}
                  >
                    {checkedItems.has('hosts') ? <CheckIcon color="success" /> : <UncheckedIcon />}
                  </IconButton>
                </ListItemIcon>
                <ListItemText
                  primary="Target 10% of hosts"
                  secondary="Minimize risk while gathering meaningful data"
                />
              </ListItem>
              <ListItem>
                <ListItemIcon>
                  <IconButton
                    size="small"
                    onClick={() => handleCheckItem('duration')}
                  >
                    {checkedItems.has('duration') ? <CheckIcon color="success" /> : <UncheckedIcon />}
                  </IconButton>
                </ListItemIcon>
                <ListItemText
                  primary="Run for at least 1 hour"
                  secondary="Capture full workload patterns and peak times"
                />
              </ListItem>
            </List>

            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button
                variant="contained"
                color="primary"
                size="large"
                startIcon={<ScienceIcon />}
                onClick={handleStartExperiment}
                fullWidth
              >
                Create My First Experiment
              </Button>
            </Box>

            <Alert severity="success" sx={{ mt: 2 }}>
              <Typography variant="subtitle2">
                Ready to start? Phoenix will guide you through each step!
              </Typography>
            </Alert>
          </Box>
        )

      case 3:
        return (
          <Box>
            <Typography variant="h5" gutterBottom sx={{ display: 'flex', alignItems: 'center' }}>
              <LightbulbIcon sx={{ mr: 1 }} />
              Best Practices & Tips
            </Typography>

            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Card sx={{ height: '100%' }}>
                  <CardContent>
                    <Typography variant="h6" color="primary" gutterBottom>
                      Do's ✅
                    </Typography>
                    <List dense>
                      <ListItem>
                        <ListItemText primary="Start with small host percentages (5-10%)" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Run experiments during peak traffic" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Monitor error rates closely" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Test one optimization at a time" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Document your experiments" />
                      </ListItem>
                    </List>
                  </CardContent>
                </Card>
              </Grid>
              <Grid item xs={12} md={6}>
                <Card sx={{ height: '100%' }}>
                  <CardContent>
                    <Typography variant="h6" color="error" gutterBottom>
                      Don'ts ❌
                    </Typography>
                    <List dense>
                      <ListItem>
                        <ListItemText primary="Don't start with 100% of hosts" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Don't ignore validation errors" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Don't skip the analysis phase" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Don't promote without validation" />
                      </ListItem>
                      <ListItem>
                        <ListItemText primary="Don't run too many experiments at once" />
                      </ListItem>
                    </List>
                  </CardContent>
                </Card>
              </Grid>
            </Grid>

            <Box sx={{ mt: 3, p: 2, bgcolor: 'background.paper', borderRadius: 1 }}>
              <Typography variant="h6" gutterBottom>
                Optimization Strategies
              </Typography>
              <Grid container spacing={1}>
                <Grid item xs={12} sm={6}>
                  <Chip label="Aggregated" color="primary" sx={{ mr: 1 }} />
                  <Typography variant="caption">Best for stable workloads</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Chip label="Intelligent" color="secondary" sx={{ mr: 1 }} />
                  <Typography variant="caption">Best for dynamic environments</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Chip label="Top-K" color="info" sx={{ mr: 1 }} />
                  <Typography variant="caption">Best for high cardinality</Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Chip label="Adaptive" color="success" sx={{ mr: 1 }} />
                  <Typography variant="caption">Best for mixed workloads</Typography>
                </Grid>
              </Grid>
            </Box>

            <Alert severity="info" sx={{ mt: 3 }}>
              <Typography variant="subtitle2">
                Join our community Slack for tips and support from other Phoenix users!
              </Typography>
            </Alert>
          </Box>
        )

      default:
        return null
    }
  }

  const progress = ((activeStep + 1) / steps.length) * 100

  return (
    <Dialog
      open={open}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { minHeight: '500px' }
      }}
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h5">Getting Started with Phoenix</Typography>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
        <LinearProgress variant="determinate" value={progress} sx={{ mt: 2 }} />
      </DialogTitle>
      
      <DialogContent>
        <Stepper activeStep={activeStep} sx={{ mb: 3 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        
        <Box sx={{ minHeight: 400 }}>
          {renderStepContent(activeStep)}
        </Box>
      </DialogContent>
      
      <DialogActions>
        <Button
          disabled={activeStep === 0}
          onClick={handleBack}
        >
          Back
        </Button>
        <Box sx={{ flex: '1 1 auto' }} />
        <Button
          onClick={onClose}
          color="inherit"
        >
          Skip Tour
        </Button>
        <Button
          variant="contained"
          onClick={activeStep === steps.length - 1 ? handleComplete : handleNext}
          endIcon={activeStep < steps.length - 1 ? <ArrowForwardIcon /> : <CheckIcon />}
        >
          {activeStep === steps.length - 1 ? 'Get Started' : 'Next'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}