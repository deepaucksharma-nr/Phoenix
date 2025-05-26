import React, { useState } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Box,
  TextField,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  FormHelperText,
  Alert,
  Grid,
  Card,
  CardContent,
  CardActionArea,
  Radio,
  RadioGroup,
  FormControlLabel,
  Slider,
  Switch,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  IconButton,
  Tooltip,
  Paper,
  Collapse,
} from '@mui/material'
import {
  Close as CloseIcon,
  Info as InfoIcon,
  Science as ScienceIcon,
  Timeline as TimelineIcon,
  Speed as SpeedIcon,
  AttachMoney as MoneyIcon,
  Security as SecurityIcon,
  CheckCircle as CheckIcon,
  Warning as WarningIcon,
  Add as AddIcon,
  Delete as DeleteIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  Lightbulb as LightbulbIcon,
  AutoAwesome as AutoAwesomeIcon,
  Build as BuildIcon,
} from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch } from '@hooks/redux'
import { createExperiment } from '@store/slices/experimentSlice'
import { useNotification } from '../../hooks/useNotification'
import { CreateExperimentData } from '../../types/experiment'
import { EnhancedPipelineBuilder } from '../PipelineBuilder'

const steps = ['Basic Information', 'Pipeline Configuration', 'Target Hosts', 'Review & Launch']

const pipelineTemplates = [
  {
    id: 'process-baseline-v1',
    name: 'Baseline',
    description: 'Standard metrics collection without optimization',
    icon: <TimelineIcon />,
    recommended: false,
    savings: '0%',
  },
  {
    id: 'process-aggregated-v1',
    name: 'Aggregated',
    description: 'Time-based aggregation for high-volume metrics',
    icon: <SpeedIcon />,
    recommended: true,
    savings: '30-50%',
  },
  {
    id: 'process-intelligent-v1',
    name: 'Intelligent',
    description: 'ML-based dynamic sampling and filtering',
    icon: <AutoAwesomeIcon />,
    recommended: true,
    savings: '50-70%',
  },
  {
    id: 'process-topk-v1',
    name: 'Top-K',
    description: 'Keep only top K processes by resource usage',
    icon: <SecurityIcon />,
    recommended: false,
    savings: '60-80%',
  },
]

interface ExperimentWizardProps {
  open: boolean
  onClose: () => void
}

export const ExperimentWizard: React.FC<ExperimentWizardProps> = ({ open, onClose }) => {
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const { showNotification } = useNotification()
  const [activeStep, setActiveStep] = useState(0)
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const [useVisualBuilder, setUseVisualBuilder] = useState(false)
  const [loading, setLoading] = useState(false)

  const [formData, setFormData] = useState<Partial<CreateExperimentData>>({
    name: '',
    description: '',
    spec: {
      baseline: { name: 'process-baseline-v1' },
      candidate: { name: 'process-aggregated-v1' },
      targetHosts: [],
      duration: '1h',
      loadProfile: 'realistic',
      successCriteria: {
        minCardinalityReduction: 30,
        maxCostIncrease: 10,
        maxLatencyIncrease: 5,
        minCriticalProcessRetention: 95,
      },
    },
  })

  const [hostInput, setHostInput] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleNext = () => {
    if (validateStep(activeStep)) {
      setActiveStep((prev) => prev + 1)
    }
  }

  const handleBack = () => {
    setActiveStep((prev) => prev - 1)
  }

  const validateStep = (step: number): boolean => {
    const newErrors: Record<string, string> = {}

    switch (step) {
      case 0:
        if (!formData.name?.trim()) {
          newErrors.name = 'Experiment name is required'
        }
        if (formData.name && formData.name.length < 3) {
          newErrors.name = 'Name must be at least 3 characters'
        }
        break
      case 1:
        if (!formData.spec?.candidate?.name) {
          newErrors.pipeline = 'Please select a candidate pipeline'
        }
        if (formData.spec?.candidate?.name === formData.spec?.baseline?.name) {
          newErrors.pipeline = 'Candidate pipeline must be different from baseline'
        }
        break
      case 2:
        if (!formData.spec?.targetHosts || formData.spec.targetHosts.length === 0) {
          newErrors.hosts = 'Please specify target hosts'
        }
        break
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleCreateExperiment = async () => {
    if (!validateStep(activeStep)) return

    setLoading(true)
    try {
      const result = await dispatch(createExperiment(formData as CreateExperimentData))
      if (createExperiment.fulfilled.match(result)) {
        const experiment = result.payload
        showNotification(`Experiment "${experiment.name}" created successfully`, 'success')
        onClose()
        navigate(`/experiments/${experiment.id}`)
      }
    } catch (error) {
      showNotification('Failed to create experiment', 'error')
    } finally {
      setLoading(false)
    }
  }

  const addHost = () => {
    if (hostInput.trim()) {
      setFormData({
        ...formData,
        spec: {
          ...formData.spec!,
          targetHosts: [...(formData.spec?.targetHosts || []), hostInput.trim()],
        },
      })
      setHostInput('')
    }
  }

  const removeHost = (host: string) => {
    setFormData({
      ...formData,
      spec: {
        ...formData.spec!,
        targetHosts: formData.spec?.targetHosts?.filter((h: string) => h !== host) || [],
      },
    })
  }

  const renderStepContent = (step: number) => {
    switch (step) {
      case 0:
        return (
          <Box>
            <TextField
              fullWidth
              label="Experiment Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              error={!!errors.name}
              helperText={errors.name || 'Give your experiment a meaningful name'}
              margin="normal"
              autoFocus
            />
            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              multiline
              rows={3}
              margin="normal"
              helperText="Describe the goal and expected outcome of this experiment"
            />
            
            <Box sx={{ mt: 3 }}>
              <Button
                variant="text"
                onClick={() => setAdvancedOpen(!advancedOpen)}
                endIcon={advancedOpen ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              >
                Advanced Settings
              </Button>
              <Collapse in={advancedOpen}>
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Experiment Duration
                  </Typography>
                  <Slider
                    value={parseInt(formData.spec?.duration?.replace(/[^0-9]/g, '') || '1') * 3600}
                    onChange={(e, value) => setFormData({ 
                      ...formData, 
                      spec: { 
                        ...formData.spec!, 
                        duration: `${Math.floor((value as number) / 3600)}h` 
                      } 
                    })}
                    min={1800}
                    max={86400}
                    step={1800}
                    marks={[
                      { value: 1800, label: '30m' },
                      { value: 3600, label: '1h' },
                      { value: 14400, label: '4h' },
                      { value: 86400, label: '24h' },
                    ]}
                    valueLabelDisplay="auto"
                    valueLabelFormat={(value) => `${value / 3600}h`}
                  />
                  
                  <FormControlLabel
                    control={
                      <Switch
                        checked={true}
                        onChange={(e) => console.log('Auto rollback:', e.target.checked)}
                      />
                    }
                    label="Enable automatic rollback on failure"
                    sx={{ mt: 2 }}
                  />
                  
                  <Grid container spacing={2} sx={{ mt: 1 }}>
                    <Grid item xs={6}>
                      <TextField
                        fullWidth
                        type="number"
                        label="Min Metric Reduction %"
                        value={formData.spec?.successCriteria?.minCardinalityReduction || 30}
                        onChange={(e) => setFormData({
                          ...formData,
                          spec: {
                            ...formData.spec!,
                            successCriteria: {
                              ...formData.spec?.successCriteria!,
                              minCardinalityReduction: parseInt(e.target.value),
                            },
                          },
                        })}
                        InputProps={{ inputProps: { min: 0, max: 100 } }}
                        helperText="Minimum required metric reduction"
                      />
                    </Grid>
                    <Grid item xs={6}>
                      <TextField
                        fullWidth
                        type="number"
                        label="Max Error Rate %"
                        value={formData.spec?.successCriteria?.maxLatencyIncrease || 5}
                        onChange={(e) => setFormData({
                          ...formData,
                          spec: {
                            ...formData.spec!,
                            successCriteria: {
                              ...formData.spec?.successCriteria!,
                              maxLatencyIncrease: parseFloat(e.target.value),
                            },
                          },
                        })}
                        InputProps={{ inputProps: { min: 0, max: 100, step: 0.1 } }}
                        helperText="Maximum allowed error rate"
                      />
                    </Grid>
                  </Grid>
                </Box>
              </Collapse>
            </Box>
          </Box>
        )

      case 1:
        return (
          <Box>
            <Alert severity="info" sx={{ mb: 3 }}>
              <Typography variant="subtitle2">How it works:</Typography>
              We'll run your baseline and candidate pipelines side-by-side on a subset of hosts,
              comparing metrics volume and quality to determine the best configuration.
            </Alert>

            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Box>
                <Typography variant="h6" gutterBottom>
                  Select Candidate Pipeline
                </Typography>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  Choose an optimized pipeline to test against your baseline
                </Typography>
              </Box>
              <Button
                variant="outlined"
                startIcon={<BuildIcon />}
                onClick={() => setUseVisualBuilder(!useVisualBuilder)}
              >
                {useVisualBuilder ? 'Use Templates' : 'Visual Builder'}
              </Button>
            </Box>

            {errors.pipeline && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {errors.pipeline}
              </Alert>
            )}

            {useVisualBuilder ? (
              <Box sx={{ height: 400, border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
                <EnhancedPipelineBuilder
                  embedded
                  experimentMode
                  onSave={(config) => {
                    setFormData({
                      ...formData,
                      spec: {
                        ...formData.spec!,
                        candidate: { name: 'custom', ...config },
                      },
                    })
                    showNotification('Pipeline configuration saved', 'success')
                  }}
                  onValidate={(isValid, errors) => {
                    if (!isValid) {
                      setErrors({ ...errors, pipeline: errors.join(', ') })
                    } else {
                      setErrors({ ...errors, pipeline: '' })
                    }
                  }}
                />
              </Box>
            ) : (
            <Grid container spacing={2} sx={{ mt: 1 }}>
              {pipelineTemplates.map((template) => (
                <Grid item xs={12} sm={6} key={template.id}>
                  <Card
                    sx={{
                      border: 2,
                      borderColor: formData.spec?.candidate?.name === template.id
                        ? 'primary.main'
                        : 'transparent',
                      position: 'relative',
                    }}
                  >
                    {template.recommended && (
                      <Chip
                        label="Recommended"
                        color="success"
                        size="small"
                        sx={{ position: 'absolute', top: 8, right: 8 }}
                      />
                    )}
                    <CardActionArea
                      onClick={() => setFormData({
                        ...formData,
                        spec: {
                          ...formData.spec!,
                          candidate: { name: template.id },
                        },
                      })}
                    >
                      <CardContent>
                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                          {template.icon}
                          <Typography variant="h6" sx={{ ml: 1 }}>
                            {template.name}
                          </Typography>
                        </Box>
                        <Typography variant="body2" color="text.secondary" gutterBottom>
                          {template.description}
                        </Typography>
                        <Box sx={{ display: 'flex', alignItems: 'center', mt: 2 }}>
                          <MoneyIcon fontSize="small" color="success" />
                          <Typography variant="subtitle2" color="success.main" sx={{ ml: 0.5 }}>
                            Est. Savings: {template.savings}
                          </Typography>
                        </Box>
                      </CardContent>
                    </CardActionArea>
                  </Card>
                </Grid>
              ))}
            </Grid>
            )}

            <Box sx={{ mt: 3, p: 2, bgcolor: 'background.paper', borderRadius: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <LightbulbIcon color="primary" />
                <Typography variant="subtitle2" sx={{ ml: 1 }}>
                  Pro Tip
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary">
                {useVisualBuilder 
                  ? 'Build custom pipelines by dragging processors onto the canvas. Connect them to create your data flow.'
                  : 'Start with "Aggregated" for predictable workloads or "Intelligent" for dynamic environments.'}
              </Typography>
            </Box>
          </Box>
        )

      case 2:
        return (
          <Box>
            <Typography variant="h6" gutterBottom>
              Target Hosts
            </Typography>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Choose which hosts will participate in the experiment
            </Typography>

            {errors.hosts && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {errors.hosts}
              </Alert>
            )}

            <RadioGroup
              value={formData.spec?.targetHosts?.length ? 'specific' : 'percentage'}
              onChange={(e) => {
                if (e.target.value === 'percentage') {
                  setFormData({
                    ...formData,
                    spec: {
                      ...formData.spec!,
                      targetHosts: [],
                    },
                  })
                }
              }}
            >
              <FormControlLabel
                value="percentage"
                control={<Radio />}
                label="Percentage of hosts"
              />
              {formData.spec?.targetHosts?.length === 0 && (
                <Box sx={{ ml: 4, mb: 2 }}>
                  <Typography variant="body2" gutterBottom>
                    Randomly select a percentage of all hosts
                  </Typography>
                  <Slider
                    value={10}
                    onChange={(e, value) => console.log('Percentage:', value)}
                    min={5}
                    max={50}
                    step={5}
                    marks={[
                      { value: 5, label: '5%' },
                      { value: 10, label: '10%' },
                      { value: 25, label: '25%' },
                      { value: 50, label: '50%' },
                    ]}
                    valueLabelDisplay="auto"
                    valueLabelFormat={(value) => `${value}%`}
                  />
                  <Alert severity="warning" sx={{ mt: 2 }}>
                    <Typography variant="body2">
                      Starting with 10% is recommended for initial experiments
                    </Typography>
                  </Alert>
                </Box>
              )}

              <FormControlLabel
                value="specific"
                control={<Radio />}
                label="Specific hosts"
              />
            </RadioGroup>

            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" gutterBottom>
                Add specific hosts to the experiment
              </Typography>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <TextField
                  fullWidth
                  size="small"
                  placeholder="Enter hostname or pattern (e.g., web-server-*)"
                  value={hostInput}
                  onChange={(e) => setHostInput(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && addHost()}
                />
                <Button
                  variant="contained"
                  onClick={addHost}
                  disabled={!hostInput.trim()}
                >
                  Add
                </Button>
              </Box>
              
              {formData.spec?.targetHosts && formData.spec.targetHosts.length > 0 && (
                <Paper variant="outlined" sx={{ mt: 2, p: 1 }}>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                    {formData.spec.targetHosts.map((host: string) => (
                      <Chip
                        key={host}
                        label={host}
                        onDelete={() => removeHost(host)}
                        size="small"
                      />
                    ))}
                  </Box>
                </Paper>
              )}
            </Box>
          </Box>
        )

      case 3:
        return (
          <Box>
            <Alert severity="success" sx={{ mb: 3 }}>
              <Typography variant="subtitle2">Ready to launch!</Typography>
              Review your experiment configuration before starting.
            </Alert>

            <List>
              <ListItem>
                <ListItemIcon>
                  <ScienceIcon />
                </ListItemIcon>
                <ListItemText
                  primary="Experiment Name"
                  secondary={formData.name}
                />
              </ListItem>
              
              <Divider />
              
              <ListItem>
                <ListItemIcon>
                  <TimelineIcon />
                </ListItemIcon>
                <ListItemText
                  primary="Pipeline Configuration"
                  secondary={
                    <>
                      Baseline: {formData.spec?.baseline?.name}
                      <br />
                      Candidate: {formData.spec?.candidate?.name}
                    </>
                  }
                />
              </ListItem>
              
              <Divider />
              
              <ListItem>
                <ListItemIcon>
                  <SecurityIcon />
                </ListItemIcon>
                <ListItemText
                  primary="Target Hosts"
                  secondary={
                    formData.spec?.targetHosts?.length
                      ? `${formData.spec.targetHosts.length} specific hosts`
                      : `10% of all hosts`
                  }
                />
              </ListItem>
              
              <Divider />
              
              <ListItem>
                <ListItemIcon>
                  <SpeedIcon />
                </ListItemIcon>
                <ListItemText
                  primary="Duration"
                  secondary={formData.spec?.duration || '1h'}
                />
              </ListItem>
              
              <Divider />
              
              <ListItem>
                <ListItemIcon>
                  <CheckIcon />
                </ListItemIcon>
                <ListItemText
                  primary="Success Criteria"
                  secondary={
                    <>
                      Min reduction: {formData.spec?.successCriteria?.minCardinalityReduction}%
                      <br />
                      Max latency increase: {formData.spec?.successCriteria?.maxLatencyIncrease}%
                    </>
                  }
                />
              </ListItem>
            </List>

            <Paper sx={{ p: 2, bgcolor: 'primary.light', mt: 3 }}>
              <Typography variant="subtitle2" gutterBottom>
                What happens next?
              </Typography>
              <Typography variant="body2">
                1. Phoenix will deploy dual collectors to your target hosts
                <br />
                2. Metrics will be collected using both pipelines simultaneously
                <br />
                3. Real-time analysis will compare volume reduction and data quality
                <br />
                4. You'll receive notifications on experiment progress
                <br />
                5. Results will be available for review after completion
              </Typography>
            </Paper>
          </Box>
        )

      default:
        return null
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { minHeight: '600px' }
      }}
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h5">Create New Experiment</Typography>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>
      
      <DialogContent>
        <Stepper activeStep={activeStep} orientation="vertical">
          {steps.map((label, index) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
              <StepContent>
                {renderStepContent(index)}
                <Box sx={{ mt: 3 }}>
                  <Button
                    variant="contained"
                    onClick={index === steps.length - 1 ? handleCreateExperiment : handleNext}
                    sx={{ mr: 1 }}
                    disabled={loading}
                  >
                    {index === steps.length - 1 ? 'Launch Experiment' : 'Continue'}
                  </Button>
                  <Button
                    disabled={index === 0 || loading}
                    onClick={handleBack}
                  >
                    Back
                  </Button>
                </Box>
              </StepContent>
            </Step>
          ))}
        </Stepper>
      </DialogContent>
    </Dialog>
  )
}