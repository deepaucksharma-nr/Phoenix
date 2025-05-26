import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  CardActions,
  Button,
  TextField,
  InputAdornment,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tabs,
  Tab,
  Alert,
  CircularProgress,
  Tooltip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Collapse,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Divider,
} from '@mui/material';
import {
  Search as SearchIcon,
  FilterList as FilterListIcon,
  Code as CodeIcon,
  Description as DescriptionIcon,
  Category as CategoryIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  CloudUpload as CloudUploadIcon,
  ContentCopy as ContentCopyIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { tomorrow } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface PipelineTemplate {
  id: string;
  name: string;
  description: string;
  category: 'optimization' | 'filtering' | 'sampling' | 'aggregation' | 'custom';
  version: string;
  author: string;
  tags: string[];
  performance: {
    avgLatency: string;
    cpuUsage: string;
    memoryUsage: string;
    cardinalityReduction: string;
  };
  processors: string[];
  configuration: {
    yaml: string;
    parameters: Array<{
      name: string;
      type: string;
      default: any;
      description: string;
      required: boolean;
    }>;
  };
  examples: Array<{
    name: string;
    description: string;
    config: string;
  }>;
  compatibility: {
    otelVersion: string;
    platforms: string[];
  };
  lastUpdated: string;
}

export const PipelineCatalog: React.FC = () => {
  const dispatch = useAppDispatch();
  const [templates, setTemplates] = useState<PipelineTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedTemplate, setSelectedTemplate] = useState<PipelineTemplate | null>(null);
  const [showYamlDialog, setShowYamlDialog] = useState(false);
  const [activeTab, setActiveTab] = useState(0);
  const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set());
  const [copiedYaml, setCopiedYaml] = useState(false);

  // Mock data - in production, this would come from the API
  useEffect(() => {
    const mockTemplates: PipelineTemplate[] = [
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
        processors: ['attributes', 'resource', 'batch', 'memory_limiter'],
        configuration: {
          yaml: `processors:
  attributes:
    actions:
      - key: process.executable.name
        action: hash
      - key: process.pid
        action: delete
      - key: process.command_line
        pattern: ^(.{0,100}).*
        action: extract
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
          parameters: [
            {
              name: 'batch_timeout',
              type: 'duration',
              default: '200ms',
              description: 'Maximum time to wait before sending a batch',
              required: true,
            },
            {
              name: 'memory_limit',
              type: 'integer',
              default: 512,
              description: 'Memory limit in MiB',
              required: true,
            },
          ],
        },
        examples: [
          {
            name: 'High-cardinality reduction',
            description: 'Aggressive cardinality reduction for high-volume environments',
            config: 'batch.timeout: 100ms\nmemory_limiter.limit_mib: 1024',
          },
        ],
        compatibility: {
          otelVersion: '>=0.88.0',
          platforms: ['linux', 'windows', 'darwin'],
        },
        lastUpdated: '2024-03-15T10:30:00Z',
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
        processors: ['tail_sampling', 'attributes', 'batch'],
        configuration: {
          yaml: `processors:
  tail_sampling:
    decision_wait: 10s
    num_traces: 100000
    expected_new_traces_per_sec: 10000
    policies:
      - name: errors-policy
        type: status_code
        status_code: {status_codes: [ERROR]}
      - name: slow-traces-policy
        type: latency
        latency: {threshold_ms: 1000}
      - name: probabilistic-policy
        type: probabilistic
        probabilistic: {sampling_percentage: 10}`,
          parameters: [
            {
              name: 'decision_wait',
              type: 'duration',
              default: '10s',
              description: 'Time to wait for trace completion',
              required: true,
            },
            {
              name: 'sampling_percentage',
              type: 'float',
              default: 10,
              description: 'Percentage of traces to sample',
              required: false,
            },
          ],
        },
        examples: [
          {
            name: 'Error-focused sampling',
            description: 'Prioritize error traces with minimal normal traffic',
            config: 'policies.probabilistic.sampling_percentage: 1',
          },
        ],
        compatibility: {
          otelVersion: '>=0.90.0',
          platforms: ['linux', 'windows', 'darwin'],
        },
        lastUpdated: '2024-03-20T14:45:00Z',
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
        processors: ['metricstransform', 'filter', 'batch'],
        configuration: {
          yaml: `processors:
  metricstransform:
    transforms:
      - include: .*
        match_type: regexp
        action: update
        operations:
          - action: aggregate_labels
            label_set: [service.name, service.namespace]
            aggregation_type: sum
  filter:
    metrics:
      exclude:
        match_type: regexp
        metric_names:
          - .*_temp$
          - .*_debug$`,
          parameters: [
            {
              name: 'aggregation_interval',
              type: 'duration',
              default: '60s',
              description: 'Interval for metric aggregation',
              required: true,
            },
          ],
        },
        examples: [
          {
            name: 'Service-level aggregation',
            description: 'Aggregate all metrics at service level',
            config: 'label_set: [service.name]',
          },
        ],
        compatibility: {
          otelVersion: '>=0.85.0',
          platforms: ['linux', 'windows'],
        },
        lastUpdated: '2024-03-10T09:15:00Z',
      },
    ];

    setTimeout(() => {
      setTemplates(mockTemplates);
      setLoading(false);
    }, 1000);
  }, []);

  const filteredTemplates = templates.filter(template => {
    const matchesSearch = searchTerm === '' || 
      template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      template.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()));
    
    const matchesCategory = selectedCategory === 'all' || template.category === selectedCategory;
    
    return matchesSearch && matchesCategory;
  });

  const handleViewDetails = (template: PipelineTemplate) => {
    setSelectedTemplate(template);
    setShowYamlDialog(true);
    setActiveTab(0);
  };

  const handleCopyYaml = () => {
    if (selectedTemplate) {
      navigator.clipboard.writeText(selectedTemplate.configuration.yaml);
      setCopiedYaml(true);
      setTimeout(() => setCopiedYaml(false), 2000);
    }
  };

  const handleDeploy = (template: PipelineTemplate) => {
    // In production, this would navigate to deployment flow or open deployment dialog
    console.log('Deploy template:', template.id);
  };

  const toggleCardExpansion = (templateId: string) => {
    const newExpanded = new Set(expandedCards);
    if (newExpanded.has(templateId)) {
      newExpanded.delete(templateId);
    } else {
      newExpanded.add(templateId);
    }
    setExpandedCards(newExpanded);
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'optimization':
        return <SpeedIcon />;
      case 'filtering':
        return <FilterListIcon />;
      case 'sampling':
        return <CategoryIcon />;
      case 'aggregation':
        return <MemoryIcon />;
      default:
        return <CodeIcon />;
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'optimization':
        return 'primary';
      case 'filtering':
        return 'secondary';
      case 'sampling':
        return 'info';
      case 'aggregation':
        return 'success';
      default:
        return 'default';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Pipeline Catalog
      </Typography>
      <Typography variant="body1" color="text.secondary" paragraph>
        Browse and deploy pre-configured OpenTelemetry pipeline templates optimized for different use cases.
      </Typography>

      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              variant="outlined"
              placeholder="Search templates..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon />
                  </InputAdornment>
                ),
              }}
            />
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Category</InputLabel>
              <Select
                value={selectedCategory}
                label="Category"
                onChange={(e) => setSelectedCategory(e.target.value)}
              >
                <MenuItem value="all">All Categories</MenuItem>
                <MenuItem value="optimization">Optimization</MenuItem>
                <MenuItem value="filtering">Filtering</MenuItem>
                <MenuItem value="sampling">Sampling</MenuItem>
                <MenuItem value="aggregation">Aggregation</MenuItem>
                <MenuItem value="custom">Custom</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <Typography variant="body2" color="text.secondary" textAlign="right">
              {filteredTemplates.length} templates available
            </Typography>
          </Grid>
        </Grid>
      </Paper>

      {loading ? (
        <Box display="flex" justifyContent="center" py={4}>
          <CircularProgress />
        </Box>
      ) : (
        <Grid container spacing={3}>
          {filteredTemplates.map((template) => (
            <Grid item xs={12} md={6} lg={4} key={template.id}>
              <Card 
                sx={{ 
                  height: '100%', 
                  display: 'flex', 
                  flexDirection: 'column',
                  transition: 'transform 0.2s',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    boxShadow: 3,
                  },
                }}
              >
                <CardContent sx={{ flexGrow: 1 }}>
                  <Box display="flex" alignItems="center" mb={2}>
                    <Box 
                      sx={{ 
                        p: 1, 
                        borderRadius: 1, 
                        bgcolor: `${getCategoryColor(template.category)}.light`,
                        color: `${getCategoryColor(template.category)}.dark`,
                        mr: 2,
                      }}
                    >
                      {getCategoryIcon(template.category)}
                    </Box>
                    <Box flexGrow={1}>
                      <Typography variant="h6" component="div">
                        {template.name}
                      </Typography>
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
                      <Chip
                        key={tag}
                        label={tag}
                        size="small"
                        sx={{ mr: 0.5, mb: 0.5 }}
                      />
                    ))}
                  </Box>

                  <Grid container spacing={2} sx={{ mb: 2 }}>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Cardinality Reduction
                        </Typography>
                        <Typography variant="body2" fontWeight="bold" color="success.main">
                          {template.performance.cardinalityReduction}
                        </Typography>
                      </Box>
                    </Grid>
                    <Grid item xs={6}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Avg Latency
                        </Typography>
                        <Typography variant="body2" fontWeight="bold">
                          {template.performance.avgLatency}
                        </Typography>
                      </Box>
                    </Grid>
                  </Grid>

                  <Collapse in={expandedCards.has(template.id)}>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="subtitle2" gutterBottom>
                      Processors
                    </Typography>
                    <List dense>
                      {template.processors.map((processor) => (
                        <ListItem key={processor} sx={{ py: 0 }}>
                          <ListItemIcon sx={{ minWidth: 30 }}>
                            <CheckCircleIcon fontSize="small" color="success" />
                          </ListItemIcon>
                          <ListItemText primary={processor} />
                        </ListItem>
                      ))}
                    </List>
                    <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                      Compatible with OTel {template.compatibility.otelVersion}
                    </Typography>
                  </Collapse>
                </CardContent>

                <CardActions>
                  <Button 
                    size="small"
                    startIcon={expandedCards.has(template.id) ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                    onClick={() => toggleCardExpansion(template.id)}
                  >
                    {expandedCards.has(template.id) ? 'Less' : 'More'}
                  </Button>
                  <Button 
                    size="small" 
                    startIcon={<CodeIcon />}
                    onClick={() => handleViewDetails(template)}
                  >
                    View YAML
                  </Button>
                  <Button 
                    size="small" 
                    variant="contained"
                    startIcon={<CloudUploadIcon />}
                    onClick={() => handleDeploy(template)}
                  >
                    Deploy
                  </Button>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      <Dialog
        open={showYamlDialog}
        onClose={() => setShowYamlDialog(false)}
        maxWidth="md"
        fullWidth
      >
        {selectedTemplate && (
          <>
            <DialogTitle>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Typography variant="h6">{selectedTemplate.name}</Typography>
                <IconButton
                  onClick={handleCopyYaml}
                  disabled={activeTab !== 0}
                  color={copiedYaml ? 'success' : 'default'}
                >
                  <ContentCopyIcon />
                </IconButton>
              </Box>
            </DialogTitle>
            <DialogContent>
              <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 2 }}>
                <Tab label="Configuration" />
                <Tab label="Parameters" />
                <Tab label="Examples" />
              </Tabs>

              {activeTab === 0 && (
                <Box>
                  <Alert severity="info" sx={{ mb: 2 }}>
                    Copy this configuration to use in your pipeline deployment.
                  </Alert>
                  <Paper sx={{ p: 0, overflow: 'hidden' }}>
                    <SyntaxHighlighter
                      language="yaml"
                      style={tomorrow}
                      customStyle={{
                        margin: 0,
                        maxHeight: '400px',
                      }}
                    >
                      {selectedTemplate.configuration.yaml}
                    </SyntaxHighlighter>
                  </Paper>
                </Box>
              )}

              {activeTab === 1 && (
                <Box>
                  <Typography variant="body2" color="text.secondary" paragraph>
                    Configure these parameters when deploying the pipeline.
                  </Typography>
                  <List>
                    {selectedTemplate.configuration.parameters.map((param) => (
                      <ListItem key={param.name} sx={{ px: 0 }}>
                        <ListItemIcon>
                          {param.required ? (
                            <WarningIcon color="warning" />
                          ) : (
                            <InfoIcon color="info" />
                          )}
                        </ListItemIcon>
                        <ListItemText
                          primary={
                            <Box display="flex" alignItems="center" gap={1}>
                              <Typography variant="subtitle2">
                                {param.name}
                              </Typography>
                              <Chip 
                                label={param.type} 
                                size="small" 
                                variant="outlined"
                              />
                              {param.required && (
                                <Chip 
                                  label="Required" 
                                  size="small" 
                                  color="warning"
                                />
                              )}
                            </Box>
                          }
                          secondary={
                            <>
                              <Typography variant="body2" color="text.secondary">
                                {param.description}
                              </Typography>
                              <Typography variant="caption" color="text.secondary">
                                Default: {JSON.stringify(param.default)}
                              </Typography>
                            </>
                          }
                        />
                      </ListItem>
                    ))}
                  </List>
                </Box>
              )}

              {activeTab === 2 && (
                <Box>
                  {selectedTemplate.examples.map((example, index) => (
                    <Box key={index} mb={3}>
                      <Typography variant="subtitle2" gutterBottom>
                        {example.name}
                      </Typography>
                      <Typography variant="body2" color="text.secondary" paragraph>
                        {example.description}
                      </Typography>
                      <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
                        <Typography variant="body2" component="pre" sx={{ fontFamily: 'monospace' }}>
                          {example.config}
                        </Typography>
                      </Paper>
                    </Box>
                  ))}
                </Box>
              )}
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setShowYamlDialog(false)}>Close</Button>
              <Button 
                variant="contained" 
                startIcon={<CloudUploadIcon />}
                onClick={() => {
                  handleDeploy(selectedTemplate);
                  setShowYamlDialog(false);
                }}
              >
                Deploy This Pipeline
              </Button>
            </DialogActions>
          </>
        )}
      </Dialog>
    </Box>
  );
};