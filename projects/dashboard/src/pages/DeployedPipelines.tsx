import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  IconButton,
  Chip,
  TextField,
  InputAdornment,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid,
  Card,
  CardContent,
  Button,
  Tooltip,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Search as SearchIcon,
  Refresh as RefreshIcon,
  Visibility as VisibilityIcon,
  Science as ScienceIcon,
  Delete as DeleteIcon,
  TrendingDown as TrendingDownIcon,
  Speed as SpeedIcon,
  Storage as StorageIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { fetchPipelineDeployments } from '../store/slices/pipelineSlice';

interface PipelineDeployment {
  id: string;
  name: string;
  pipeline: string;
  namespace: string;
  status: string;
  phase: string;
  targetNodes: Record<string, string>;
  instances?: {
    desired: number;
    ready: number;
  };
  metrics?: {
    cardinality: number;
    throughput: string;
    errorRate: number;
    cpuUsage: number;
    memoryUsage: number;
  };
  createdAt: string;
  updatedAt: string;
}

export default function DeployedPipelines() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { deployments, loading, error } = useAppSelector((state) => state.pipeline);

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [pipelineFilter, setPipelineFilter] = useState('all');

  useEffect(() => {
    dispatch(fetchPipelineDeployments());
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(() => {
      dispatch(fetchPipelineDeployments());
    }, 30000);

    return () => clearInterval(interval);
  }, [dispatch]);

  const handleRefresh = () => {
    dispatch(fetchPipelineDeployments());
  };

  const handleStartExperiment = (deployment: PipelineDeployment) => {
    navigate('/experiments/create', { 
      state: { 
        baselinePipeline: deployment.pipeline,
        targetNodes: deployment.targetNodes 
      } 
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'updating':
        return 'warning';
      case 'failed':
        return 'error';
      case 'deleting':
        return 'default';
      default:
        return 'info';
    }
  };

  const filteredDeployments = deployments?.filter(deployment => {
    const matchesSearch = deployment.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         deployment.pipeline.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         Object.keys(deployment.targetNodes).some(node => 
                           node.toLowerCase().includes(searchQuery.toLowerCase())
                         );
    
    const matchesStatus = statusFilter === 'all' || deployment.status === statusFilter;
    const matchesPipeline = pipelineFilter === 'all' || deployment.pipeline === pipelineFilter;
    
    return matchesSearch && matchesStatus && matchesPipeline;
  }) || [];

  const totalCardinality = filteredDeployments.reduce((sum, d) => 
    sum + (d.metrics?.cardinality || 0), 0
  );

  const avgReduction = filteredDeployments.length > 0
    ? filteredDeployments.reduce((sum, d) => {
        const baseline = 10000; // Assumed baseline
        const reduction = d.metrics?.cardinality 
          ? ((baseline - d.metrics.cardinality) / baseline) * 100 
          : 0;
        return sum + reduction;
      }, 0) / filteredDeployments.length
    : 0;

  const totalNodes = new Set(
    filteredDeployments.flatMap(d => Object.keys(d.targetNodes))
  ).size;

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h4">Deployed Pipelines</Typography>
        <Button
          variant="contained"
          startIcon={<RefreshIcon />}
          onClick={handleRefresh}
          disabled={loading}
        >
          Refresh
        </Button>
      </Box>

      {/* Metrics Summary */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Active Deployments
                  </Typography>
                  <Typography variant="h4">
                    {filteredDeployments.filter(d => d.status === 'active').length}
                  </Typography>
                </Box>
                <SpeedIcon color="primary" sx={{ fontSize: 40, opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Total Cardinality
                  </Typography>
                  <Typography variant="h4">
                    {totalCardinality.toLocaleString()}
                  </Typography>
                </Box>
                <StorageIcon color="primary" sx={{ fontSize: 40, opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Avg. Reduction
                  </Typography>
                  <Typography variant="h4">
                    {avgReduction.toFixed(1)}%
                  </Typography>
                </Box>
                <TrendingDownIcon color="success" sx={{ fontSize: 40, opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Target Nodes
                  </Typography>
                  <Typography variant="h4">
                    {totalNodes}
                  </Typography>
                </Box>
                <StorageIcon color="primary" sx={{ fontSize: 40, opacity: 0.3 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={4}>
            <TextField
              fullWidth
              variant="outlined"
              placeholder="Search by name, pipeline, or host..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
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
              <InputLabel>Status</InputLabel>
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                label="Status"
              >
                <MenuItem value="all">All Status</MenuItem>
                <MenuItem value="active">Active</MenuItem>
                <MenuItem value="updating">Updating</MenuItem>
                <MenuItem value="failed">Failed</MenuItem>
                <MenuItem value="deleting">Deleting</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Pipeline Type</InputLabel>
              <Select
                value={pipelineFilter}
                onChange={(e) => setPipelineFilter(e.target.value)}
                label="Pipeline Type"
              >
                <MenuItem value="all">All Pipelines</MenuItem>
                <MenuItem value="process-baseline-v1">Baseline</MenuItem>
                <MenuItem value="process-sampling-v1">Sampling</MenuItem>
                <MenuItem value="process-topk-v1">Top-K</MenuItem>
                <MenuItem value="process-adaptive-filter-v1">Adaptive Filter</MenuItem>
                <MenuItem value="process-anomaly-v1">Anomaly Detection</MenuItem>
              </Select>
            </FormControl>
          </Grid>
        </Grid>
      </Paper>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Deployments Table */}
      <TableContainer component={Paper}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Host</TableCell>
                  <TableCell>Pipeline</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell align="right">Instances</TableCell>
                  <TableCell align="right">Cardinality</TableCell>
                  <TableCell align="right">Throughput</TableCell>
                  <TableCell align="right">Error Rate</TableCell>
                  <TableCell align="center">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredDeployments
                  .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                  .map((deployment) => (
                    <TableRow key={deployment.id} hover>
                      <TableCell>
                        <Box>
                          <Typography variant="body2" fontWeight="medium">
                            {deployment.name}
                          </Typography>
                          <Typography variant="caption" color="textSecondary">
                            {Object.keys(deployment.targetNodes).join(', ')}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {deployment.pipeline}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={deployment.status}
                          color={getStatusColor(deployment.status)}
                          size="small"
                        />
                      </TableCell>
                      <TableCell align="right">
                        {deployment.instances ? (
                          <Typography variant="body2">
                            {deployment.instances.ready}/{deployment.instances.desired}
                          </Typography>
                        ) : (
                          '-'
                        )}
                      </TableCell>
                      <TableCell align="right">
                        {deployment.metrics?.cardinality?.toLocaleString() || '-'}
                      </TableCell>
                      <TableCell align="right">
                        {deployment.metrics?.throughput || '-'}
                      </TableCell>
                      <TableCell align="right">
                        {deployment.metrics?.errorRate 
                          ? `${deployment.metrics.errorRate.toFixed(2)}%`
                          : '-'
                        }
                      </TableCell>
                      <TableCell align="center">
                        <Tooltip title="View Details">
                          <IconButton
                            size="small"
                            onClick={() => navigate(`/pipelines/deployments/${deployment.id}`)}
                          >
                            <VisibilityIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Start Experiment">
                          <IconButton
                            size="small"
                            onClick={() => handleStartExperiment(deployment)}
                            disabled={deployment.status !== 'active'}
                          >
                            <ScienceIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete Deployment">
                          <IconButton
                            size="small"
                            color="error"
                            disabled={deployment.status === 'deleting'}
                          >
                            <DeleteIcon />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))}
                {filteredDeployments.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={8} align="center" sx={{ py: 4 }}>
                      <Typography color="textSecondary">
                        No deployments found
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
            <TablePagination
              rowsPerPageOptions={[5, 10, 25]}
              component="div"
              count={filteredDeployments.length}
              rowsPerPage={rowsPerPage}
              page={page}
              onPageChange={(e, newPage) => setPage(newPage)}
              onRowsPerPageChange={(e) => {
                setRowsPerPage(parseInt(e.target.value, 10));
                setPage(0);
              }}
            />
          </>
        )}
      </TableContainer>
    </Box>
  );
}