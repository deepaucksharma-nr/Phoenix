import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Stepper,
  Step,
  StepLabel,
  Button,
  Box,
  Typography,
  TextField,
  FormControl,
  FormControlLabel,
  Checkbox,
  Radio,
  RadioGroup,
  Card,
  CardContent,
  Stack,
  Chip,
  Grid,
  Alert,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import { Rocket, CheckCircle } from '@mui/icons-material';

interface ExperimentWizardProps {
  open: boolean;
  onClose: () => void;
  onComplete: (experiment: any) => void;
}

interface HostGroup {
  name: string;
  tags: string[];
  count: number;
  selected: boolean;
}

interface PipelineTemplate {
  id: string;
  name: string;
  description: string;
  estimatedSavings: number;
  cpuImpact: number;
  category: string;
}

const steps = ['Select Hosts', 'Choose Pipeline', 'Review & Launch'];

export const ExperimentWizard: React.FC<ExperimentWizardProps> = ({
  open,
  onClose,
  onComplete,
}) => {
  const [activeStep, setActiveStep] = useState(0);
  const [experimentName, setExperimentName] = useState('');
  const [description, setDescription] = useState('');
  const [hostGroups, setHostGroups] = useState<HostGroup[]>([]);
  const [selectedHosts, setSelectedHosts] = useState<string[]>([]);
  const [templates, setTemplates] = useState<PipelineTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<string>('');
  const [duration, setDuration] = useState(24); // hours
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  useEffect(() => {
    if (open) {
      fetchHostGroups();
      fetchPipelineTemplates();
    }
  }, [open]);
  
  const fetchHostGroups = async () => {
    try {
      const response = await fetch('/api/v1/fleet/status');
      const data = await response.json();
      
      // Group agents by tags
      const groups: Map<string, HostGroup> = new Map();
      data.agents.forEach((agent: any) => {
        const groupKey = agent.group || 'default';
        if (!groups.has(groupKey)) {
          groups.set(groupKey, {
            name: groupKey,
            tags: [],
            count: 0,
            selected: false,
          });
        }
        const group = groups.get(groupKey)!;
        group.count++;
      });
      
      setHostGroups(Array.from(groups.values()));
    } catch (error) {
      console.error('Failed to fetch host groups:', error);
    }
  };
  
  const fetchPipelineTemplates = async () => {
    try {
      const response = await fetch('/api/v1/pipelines/templates');
      const data = await response.json();
      setTemplates(data);
    } catch (error) {
      console.error('Failed to fetch templates:', error);
    }
  };
  
  const handleNext = () => {
    setActiveStep((prevStep) => prevStep + 1);
  };
  
  const handleBack = () => {
    setActiveStep((prevStep) => prevStep - 1);
  };
  
  const handleHostGroupToggle = (groupName: string) => {
    setHostGroups(prev => prev.map(group =>
      group.name === groupName ? { ...group, selected: !group.selected } : group
    ));
    
    // Update selected hosts
    const group = hostGroups.find(g => g.name === groupName);
    if (group) {
      if (group.selected) {
        setSelectedHosts(prev => prev.filter(h => !h.startsWith(groupName)));
      } else {
        setSelectedHosts(prev => [...prev, `group:${groupName}`]);
      }
    }
  };
  
  const handleSubmit = async () => {
    setIsSubmitting(true);
    
    try {
      const response = await fetch('/api/v1/experiments/wizard', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: experimentName || `Experiment ${new Date().toISOString()}`,
          description,
          host_selector: selectedHosts,
          pipeline_type: selectedTemplate,
          duration_hours: duration,
        }),
      });
      
      if (!response.ok) throw new Error('Failed to create experiment');
      
      const experiment = await response.json();
      onComplete(experiment);
      onClose();
    } catch (error) {
      console.error('Failed to create experiment:', error);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  const getStepContent = (step: number) => {
    switch (step) {
      case 0:
        return (
          <Box>
            <Typography variant="h6" gutterBottom>
              Select Target Hosts
            </Typography>
            <Typography variant="body2" color="textSecondary" gutterBottom>
              Choose which hosts to run the experiment on
            </Typography>
            
            <Grid container spacing={2} sx={{ mt: 2 }}>
              {hostGroups.map((group) => (
                <Grid item xs={12} sm={6} key={group.name}>
                  <Card
                    variant="outlined"
                    sx={{
                      cursor: 'pointer',
                      transition: 'all 0.2s',
                      borderColor: group.selected ? 'primary.main' : 'divider',
                      borderWidth: group.selected ? 2 : 1,
                      '&:hover': {
                        boxShadow: 2,
                      },
                    }}
                    onClick={() => handleHostGroupToggle(group.name)}
                  >
                    <CardContent>
                      <Stack direction="row" justifyContent="space-between" alignItems="center">
                        <Box>
                          <Typography variant="subtitle1">{group.name}</Typography>
                          <Typography variant="body2" color="textSecondary">
                            {group.count} hosts
                          </Typography>
                        </Box>
                        <Checkbox
                          checked={group.selected}
                          color="primary"
                        />
                      </Stack>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="body2" color="textSecondary">
                Selected: {selectedHosts.length} group(s)
              </Typography>
            </Box>
          </Box>
        );
        
      case 1:
        return (
          <Box>
            <Typography variant="h6" gutterBottom>
              Choose Optimization Pipeline
            </Typography>
            <Typography variant="body2" color="textSecondary" gutterBottom>
              Select a pre-built optimization strategy
            </Typography>
            
            <RadioGroup
              value={selectedTemplate}
              onChange={(e) => setSelectedTemplate(e.target.value)}
            >
              <Grid container spacing={2} sx={{ mt: 2 }}>
                {templates.map((template) => (
                  <Grid item xs={12} key={template.id}>
                    <Card
                      variant="outlined"
                      sx={{
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                        borderColor: selectedTemplate === template.id ? 'primary.main' : 'divider',
                        borderWidth: selectedTemplate === template.id ? 2 : 1,
                        '&:hover': {
                          boxShadow: 2,
                        },
                      }}
                      onClick={() => setSelectedTemplate(template.id)}
                    >
                      <CardContent>
                        <Stack direction="row" spacing={2}>
                          <Radio value={template.id} />
                          <Box flex={1}>
                            <Stack direction="row" justifyContent="space-between" alignItems="center">
                              <Typography variant="subtitle1">{template.name}</Typography>
                              <Stack direction="row" spacing={1}>
                                <Chip
                                  label={`-${template.estimatedSavings}% cost`}
                                  color="success"
                                  size="small"
                                />
                                <Chip
                                  label={`+${template.cpuImpact}% CPU`}
                                  color="warning"
                                  size="small"
                                />
                              </Stack>
                            </Stack>
                            <Typography variant="body2" color="textSecondary" sx={{ mt: 1 }}>
                              {template.description}
                            </Typography>
                          </Box>
                        </Stack>
                      </CardContent>
                    </Card>
                  </Grid>
                ))}
              </Grid>
            </RadioGroup>
          </Box>
        );
        
      case 2:
        const selectedTemplateData = templates.find(t => t.id === selectedTemplate);
        const totalHosts = hostGroups
          .filter(g => g.selected)
          .reduce((sum, g) => sum + g.count, 0);
          
        return (
          <Box>
            <Typography variant="h6" gutterBottom>
              Review & Launch
            </Typography>
            
            <TextField
              fullWidth
              label="Experiment Name (Optional)"
              value={experimentName}
              onChange={(e) => setExperimentName(e.target.value)}
              margin="normal"
              placeholder="e.g., Reduce API costs Q1"
            />
            
            <TextField
              fullWidth
              label="Description (Optional)"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              margin="normal"
              multiline
              rows={2}
              placeholder="What are you trying to achieve?"
            />
            
            <TextField
              fullWidth
              label="Duration (hours)"
              type="number"
              value={duration}
              onChange={(e) => setDuration(parseInt(e.target.value) || 24)}
              margin="normal"
              InputProps={{ inputProps: { min: 1, max: 168 } }}
            />
            
            <Alert severity="info" sx={{ mt: 3 }}>
              <Typography variant="subtitle2" gutterBottom>
                Experiment Summary:
              </Typography>
              <Stack spacing={1}>
                <Typography variant="body2">
                  • Target: {totalHosts} hosts across {selectedHosts.length} groups
                </Typography>
                <Typography variant="body2">
                  • Pipeline: {selectedTemplateData?.name}
                </Typography>
                <Typography variant="body2">
                  • Expected savings: {selectedTemplateData?.estimatedSavings}%
                </Typography>
                <Typography variant="body2">
                  • Duration: {duration} hours
                </Typography>
              </Stack>
            </Alert>
          </Box>
        );
        
      default:
        return null;
    }
  };
  
  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      PaperProps={{
        sx: { minHeight: 600 }
      }}
    >
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={2}>
          <Rocket color="primary" />
          <Typography variant="h5">Create New Experiment</Typography>
        </Stack>
      </DialogTitle>
      
      <DialogContent>
        <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
        
        <AnimatePresence mode="wait">
          <motion.div
            key={activeStep}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            transition={{ duration: 0.2 }}
          >
            {getStepContent(activeStep)}
          </motion.div>
        </AnimatePresence>
      </DialogContent>
      
      <DialogActions sx={{ p: 3 }}>
        <Button onClick={onClose}>Cancel</Button>
        <Box flex={1} />
        <Button
          disabled={activeStep === 0}
          onClick={handleBack}
        >
          Back
        </Button>
        {activeStep === steps.length - 1 ? (
          <Button
            variant="contained"
            onClick={handleSubmit}
            disabled={isSubmitting}
            startIcon={<CheckCircle />}
          >
            {isSubmitting ? 'Creating...' : 'Launch Experiment'}
          </Button>
        ) : (
          <Button
            variant="contained"
            onClick={handleNext}
            disabled={
              (activeStep === 0 && selectedHosts.length === 0) ||
              (activeStep === 1 && !selectedTemplate)
            }
          >
            Next
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};