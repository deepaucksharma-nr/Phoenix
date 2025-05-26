# Phoenix Platform - PRD Implementation Examples

## Missing Component Implementation Examples

This guide provides concrete implementation examples for the critical missing components identified in the PRD alignment report.

## 1. LoadSim Operator Controller Implementation

### Location: `/projects/loadsim-operator/controllers/loadsimulationjob_controller.go`

```go
package controllers

import (
    "context"
    "fmt"
    "time"

    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    phoenixv1alpha1 "github.com/phoenix/platform/operators/loadsim/api/v1alpha1"
)

// LoadSimulationJobReconciler reconciles a LoadSimulationJob object
type LoadSimulationJobReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=phoenix.io,resources=loadsimulationjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=phoenix.io,resources=loadsimulationjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *LoadSimulationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the LoadSimulationJob instance
    var loadSimJob phoenixv1alpha1.LoadSimulationJob
    if err := r.Get(ctx, req.NamespacedName, &loadSimJob); err != nil {
        if errors.IsNotFound(err) {
            // Object deleted
            return ctrl.Result{}, nil
        }
        log.Error(err, "Failed to get LoadSimulationJob")
        return ctrl.Result{}, err
    }

    // Check if job already exists
    job := &batchv1.Job{}
    jobName := fmt.Sprintf("loadsim-%s", loadSimJob.Name)
    err := r.Get(ctx, client.ObjectKey{Name: jobName, Namespace: loadSimJob.Namespace}, job)
    
    if err != nil && errors.IsNotFound(err) {
        // Create new Job
        job = r.constructJobForLoadSim(&loadSimJob)
        if err := ctrl.SetControllerReference(&loadSimJob, job, r.Scheme); err != nil {
            return ctrl.Result{}, err
        }
        
        log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
        if err := r.Create(ctx, job); err != nil {
            log.Error(err, "Failed to create new Job")
            return ctrl.Result{}, err
        }
        
        // Update status to Running
        loadSimJob.Status.Phase = "Running"
        loadSimJob.Status.StartTime = &metav1.Time{Time: time.Now()}
        if err := r.Status().Update(ctx, &loadSimJob); err != nil {
            log.Error(err, "Failed to update LoadSimulationJob status")
            return ctrl.Result{}, err
        }
        
        return ctrl.Result{RequeueAfter: time.Second * 30}, nil
    } else if err != nil {
        log.Error(err, "Failed to get Job")
        return ctrl.Result{}, err
    }

    // Check Job status and update LoadSimulationJob accordingly
    if job.Status.Succeeded > 0 {
        loadSimJob.Status.Phase = "Completed"
        loadSimJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
    } else if job.Status.Failed > 0 {
        loadSimJob.Status.Phase = "Failed"
        loadSimJob.Status.Message = "Load simulation job failed"
    } else if job.Status.Active > 0 {
        // Count active processes
        pods := &corev1.PodList{}
        if err := r.List(ctx, pods, client.InNamespace(loadSimJob.Namespace), 
            client.MatchingLabels{"job-name": jobName}); err == nil {
            loadSimJob.Status.ActiveProcesses = int32(len(pods.Items))
        }
    }

    // Update status
    if err := r.Status().Update(ctx, &loadSimJob); err != nil {
        log.Error(err, "Failed to update LoadSimulationJob status")
        return ctrl.Result{}, err
    }

    // Requeue if still running
    if loadSimJob.Status.Phase == "Running" {
        return ctrl.Result{RequeueAfter: time.Second * 30}, nil
    }

    return ctrl.Result{}, nil
}

func (r *LoadSimulationJobReconciler) constructJobForLoadSim(loadSimJob *phoenixv1alpha1.LoadSimulationJob) *batchv1.Job {
    // Parse duration
    duration := loadSimJob.Spec.Duration
    
    // Construct environment variables based on profile
    env := []corev1.EnvVar{
        {Name: "EXPERIMENT_ID", Value: loadSimJob.Spec.ExperimentID},
        {Name: "PROFILE", Value: loadSimJob.Spec.Profile},
        {Name: "DURATION", Value: duration},
        {Name: "PROCESS_COUNT", Value: fmt.Sprintf("%d", loadSimJob.Spec.ProcessCount)},
    }

    // Add custom profile parameters if specified
    if loadSimJob.Spec.CustomProfile != nil {
        env = append(env, corev1.EnvVar{
            Name:  "CHURN_RATE",
            Value: fmt.Sprintf("%f", loadSimJob.Spec.CustomProfile.ChurnRate),
        })
    }

    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("loadsim-%s", loadSimJob.Name),
            Namespace: loadSimJob.Namespace,
            Labels: map[string]string{
                "phoenix.io/component": "load-simulator",
                "phoenix.io/experiment": loadSimJob.Spec.ExperimentID,
            },
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "phoenix.io/component": "load-simulator",
                        "phoenix.io/experiment": loadSimJob.Spec.ExperimentID,
                    },
                },
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    HostPID:       true, // Important: Access host processes
                    NodeSelector:  loadSimJob.Spec.NodeSelector,
                    Containers: []corev1.Container{
                        {
                            Name:            "load-generator",
                            Image:           "phoenix/load-generator:latest",
                            ImagePullPolicy: corev1.PullIfNotPresent,
                            Env:             env,
                            SecurityContext: &corev1.SecurityContext{
                                Privileged: &[]bool{true}[0],
                            },
                        },
                    },
                },
            },
        },
    }
    
    return job
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoadSimulationJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&phoenixv1alpha1.LoadSimulationJob{}).
        Owns(&batchv1.Job{}).
        Complete(r)
}
```

## 2. Load Generator Implementation

### Location: `/projects/loadsim-operator/internal/generator/generator.go`

```go
package generator

import (
    "context"
    "fmt"
    "math/rand"
    "os"
    "os/exec"
    "sync"
    "time"
)

// ProcessGenerator interface for different load profiles
type ProcessGenerator interface {
    Start(ctx context.Context) error
    Stop() error
    GetActiveCount() int
}

// RealisticProfile simulates typical server workload
type RealisticProfile struct {
    ProcessCount int
    processes    []*exec.Cmd
    mu           sync.Mutex
}

func (r *RealisticProfile) Start(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // Long-running processes (databases, web servers)
    for i := 0; i < r.ProcessCount/3; i++ {
        cmd := exec.CommandContext(ctx, "sleep", "infinity")
        cmd.Env = append(os.Environ(), fmt.Sprintf("PHOENIX_PROCESS_TYPE=long-running-%d", i))
        if err := cmd.Start(); err != nil {
            return fmt.Errorf("failed to start long-running process: %w", err)
        }
        r.processes = append(r.processes, cmd)
    }

    // Medium-lived processes (application workers)
    for i := 0; i < r.ProcessCount/3; i++ {
        go func(idx int) {
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    cmd := exec.CommandContext(ctx, "bash", "-c", 
                        fmt.Sprintf("sleep %d", 30+rand.Intn(120)))
                    cmd.Env = append(os.Environ(), 
                        fmt.Sprintf("PHOENIX_PROCESS_TYPE=medium-lived-%d", idx))
                    cmd.Start()
                    cmd.Wait()
                }
            }
        }(i)
    }

    // Short-lived processes (cron jobs, scripts)
    for i := 0; i < r.ProcessCount/3; i++ {
        go func(idx int) {
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    cmd := exec.CommandContext(ctx, "bash", "-c", 
                        fmt.Sprintf("sleep %d", 1+rand.Intn(10)))
                    cmd.Env = append(os.Environ(), 
                        fmt.Sprintf("PHOENIX_PROCESS_TYPE=short-lived-%d", idx))
                    cmd.Start()
                    cmd.Wait()
                    time.Sleep(time.Second * time.Duration(rand.Intn(5)))
                }
            }
        }(i)
    }

    return nil
}

// HighCardinalityProfile generates many unique process names
type HighCardinalityProfile struct {
    ProcessCount int
    ChurnRate    float64
}

func (h *HighCardinalityProfile) Start(ctx context.Context) error {
    for i := 0; i < h.ProcessCount; i++ {
        go func(idx int) {
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    // Generate unique process name
                    processName := fmt.Sprintf("phoenix-hc-proc-%d-%d", idx, time.Now().Unix())
                    cmd := exec.CommandContext(ctx, "bash", "-c", 
                        fmt.Sprintf("exec -a %s sleep %d", processName, rand.Intn(60)))
                    cmd.Start()
                    cmd.Wait()
                }
            }
        }(i)
    }
    return nil
}

// ProcessChurnProfile rapidly creates and destroys processes
type ProcessChurnProfile struct {
    ProcessCount    int
    ChurnPerSecond  int
}

func (p *ProcessChurnProfile) Start(ctx context.Context) error {
    ticker := time.NewTicker(time.Second / time.Duration(p.ChurnPerSecond))
    defer ticker.Stop()

    processID := 0
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            // Start a new process
            go func(id int) {
                cmd := exec.CommandContext(ctx, "bash", "-c", 
                    fmt.Sprintf("exec -a churn-proc-%d sleep 1", id))
                cmd.Start()
                cmd.Wait()
            }(processID)
            processID++
        }
    }
}
```

## 3. Missing CLI Commands Implementation

### Location: `/projects/phoenix-cli/cmd/pipeline_show.go`

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/phoenix/platform/projects/phoenix-cli/internal/client"
    "github.com/phoenix/platform/projects/phoenix-cli/internal/output"
)

func NewPipelineShowCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "show <pipeline-name>",
        Short: "Display the OTel YAML configuration of a catalog pipeline",
        Long:  `Show the fully resolved OpenTelemetry Collector YAML configuration for a catalog pipeline template.`,
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            pipelineName := args[0]
            
            // Get API client
            apiClient, err := client.NewAPIClient()
            if err != nil {
                return fmt.Errorf("failed to create API client: %w", err)
            }

            // Fetch pipeline config
            config, err := apiClient.GetPipelineConfig(cmd.Context(), pipelineName)
            if err != nil {
                return fmt.Errorf("failed to get pipeline config: %w", err)
            }

            // Output based on format flag
            outputFormat, _ := cmd.Flags().GetString("output")
            switch outputFormat {
            case "yaml":
                fmt.Println(config.YAML)
            case "json":
                return output.PrintJSON(config)
            default:
                // Pretty print with syntax highlighting
                fmt.Printf("Pipeline: %s\n", output.Bold(pipelineName))
                fmt.Printf("Version: %s\n", config.Version)
                fmt.Printf("Description: %s\n\n", config.Description)
                fmt.Println("Configuration:")
                fmt.Println(output.SyntaxHighlightYAML(config.YAML))
            }

            return nil
        },
    }

    cmd.Flags().StringP("output", "o", "yaml", "Output format (yaml|json|pretty)")
    
    return cmd
}
```

### Location: `/projects/phoenix-cli/cmd/loadsim.go`

```go
package cmd

import (
    "github.com/spf13/cobra"
)

func NewLoadSimCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "loadsim",
        Short: "Manage load simulation for experiments",
        Long:  `Start, stop, and monitor load simulations that generate synthetic process activity on target hosts.`,
    }

    cmd.AddCommand(
        NewLoadSimStartCommand(),
        NewLoadSimStopCommand(),
        NewLoadSimStatusCommand(),
        NewLoadSimListProfilesCommand(),
    )

    return cmd
}

func NewLoadSimStartCommand() *cobra.Command {
    var (
        profile    string
        targetHost string
        duration   string
        simJobName string
        params     map[string]string
    )

    cmd := &cobra.Command{
        Use:   "start",
        Short: "Start a load simulation",
        Long:  `Start a load simulation on a target Kubernetes node to test pipeline effectiveness.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            apiClient, err := client.NewAPIClient()
            if err != nil {
                return fmt.Errorf("failed to create API client: %w", err)
            }

            request := &client.LoadSimulationRequest{
                Profile:    profile,
                TargetHost: targetHost,
                Duration:   duration,
                JobName:    simJobName,
                Parameters: params,
            }

            job, err := apiClient.StartLoadSimulation(cmd.Context(), request)
            if err != nil {
                return fmt.Errorf("failed to start load simulation: %w", err)
            }

            fmt.Printf("Load simulation started: %s\n", output.Bold(job.Name))
            fmt.Printf("Profile: %s\n", job.Profile)
            fmt.Printf("Target: %s\n", job.TargetHost)
            fmt.Printf("Duration: %s\n", job.Duration)
            
            return nil
        },
    }

    cmd.Flags().StringVar(&profile, "profile", "realistic", 
        "Load profile (realistic|high-cardinality|process-churn|custom)")
    cmd.Flags().StringVar(&targetHost, "target-host", "", 
        "Target Kubernetes node name")
    cmd.Flags().StringVar(&duration, "duration", "10m", 
        "Simulation duration (e.g., 30s, 5m, 1h)")
    cmd.Flags().StringVar(&simJobName, "sim-job-name", "", 
        "Custom name for the simulation job")
    cmd.Flags().StringToStringVar(&params, "params", nil, 
        "Additional parameters (key=value)")

    cmd.MarkFlagRequired("target-host")

    return cmd
}
```

## 4. Web Console Missing Views

### Location: `/projects/dashboard/src/pages/DeployedPipelines.tsx`

```tsx
import React, { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Button,
  TextField,
  InputAdornment,
  IconButton,
  Tooltip,
  LinearProgress,
} from '@mui/material';
import {
  Search as SearchIcon,
  Refresh as RefreshIcon,
  Science as ExperimentIcon,
  TrendingDown as SavingsIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api.service';

interface DeployedPipeline {
  id: string;
  hostname: string;
  pipelineName: string;
  pipelineVersion: string;
  status: 'Running' | 'Error' | 'Pending';
  inputProcessCount: number;
  outputSeriesCount: number;
  cardinalityReduction: number;
  criticalProcessRetention: number;
  cpuUsage: number;
  memoryUsage: number;
  lastUpdated: string;
  error?: string;
}

export const DeployedPipelines: React.FC = () => {
  const [pipelines, setPipelines] = useState<DeployedPipeline[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [refreshing, setRefreshing] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetchPipelines();
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchPipelines, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchPipelines = async () => {
    try {
      setLoading(true);
      const response = await api.get('/pipelines/deployed');
      setPipelines(response.data);
    } catch (error) {
      console.error('Failed to fetch deployed pipelines:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleRefresh = () => {
    setRefreshing(true);
    fetchPipelines();
  };

  const handleStartExperiment = (hostname: string) => {
    navigate('/experiments/new', { state: { targetHost: hostname } });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Running':
        return 'success';
      case 'Error':
        return 'error';
      case 'Pending':
        return 'warning';
      default:
        return 'default';
    }
  };

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  const formatCostSavings = (reduction: number, inputCount: number) => {
    // Simplified cost calculation
    const estimatedMonthlyCost = inputCount * 0.0001 * 30 * 24; // $0.0001 per series per hour
    const savings = estimatedMonthlyCost * reduction;
    return `$${savings.toFixed(2)}/mo`;
  };

  const filteredPipelines = pipelines.filter(
    (p) =>
      p.hostname.toLowerCase().includes(searchTerm.toLowerCase()) ||
      p.pipelineName.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Deployed Process Pipelines
        </Typography>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <TextField
            size="small"
            placeholder="Search hosts or pipelines..."
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
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={handleRefresh}
            disabled={refreshing}
          >
            Refresh
          </Button>
        </Box>
      </Box>

      {/* Summary Cards */}
      <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 2, mb: 3 }}>
        <Card>
          <CardContent>
            <Typography color="textSecondary" gutterBottom>
              Total Hosts
            </Typography>
            <Typography variant="h4">{pipelines.length}</Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography color="textSecondary" gutterBottom>
              Active Pipelines
            </Typography>
            <Typography variant="h4">
              {pipelines.filter((p) => p.status === 'Running').length}
            </Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography color="textSecondary" gutterBottom>
              Avg. Cardinality Reduction
            </Typography>
            <Typography variant="h4">
              {formatPercentage(
                pipelines.reduce((acc, p) => acc + p.cardinalityReduction, 0) / pipelines.length || 0
              )}
            </Typography>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <Typography color="textSecondary" gutterBottom>
              Est. Monthly Savings
            </Typography>
            <Typography variant="h4" color="success.main">
              $
              {pipelines
                .reduce(
                  (acc, p) =>
                    acc + parseFloat(formatCostSavings(p.cardinalityReduction, p.inputProcessCount).slice(1, -3)),
                  0
                )
                .toFixed(2)}
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Pipelines Table */}
      {loading && <LinearProgress />}
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Hostname</TableCell>
              <TableCell>Pipeline</TableCell>
              <TableCell>Status</TableCell>
              <TableCell align="right">Input Processes</TableCell>
              <TableCell align="right">Output Series</TableCell>
              <TableCell align="right">Reduction</TableCell>
              <TableCell align="right">Critical Retention</TableCell>
              <TableCell align="right">CPU/Memory</TableCell>
              <TableCell align="right">Est. Savings</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredPipelines.map((pipeline) => (
              <TableRow key={pipeline.id}>
                <TableCell>
                  <Typography variant="body2" fontWeight="medium">
                    {pipeline.hostname}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Box>
                    <Typography variant="body2">{pipeline.pipelineName}</Typography>
                    <Typography variant="caption" color="textSecondary">
                      v{pipeline.pipelineVersion}
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell>
                  <Tooltip title={pipeline.error || ''}>
                    <Chip
                      label={pipeline.status}
                      color={getStatusColor(pipeline.status) as any}
                      size="small"
                    />
                  </Tooltip>
                </TableCell>
                <TableCell align="right">{pipeline.inputProcessCount.toLocaleString()}</TableCell>
                <TableCell align="right">{pipeline.outputSeriesCount.toLocaleString()}</TableCell>
                <TableCell align="right">
                  <Chip
                    icon={<TrendingDown />}
                    label={formatPercentage(pipeline.cardinalityReduction)}
                    color="success"
                    size="small"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell align="right">
                  {formatPercentage(pipeline.criticalProcessRetention)}
                </TableCell>
                <TableCell align="right">
                  <Typography variant="caption">
                    {pipeline.cpuUsage.toFixed(1)}% / {pipeline.memoryUsage}MB
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" color="success.main" fontWeight="medium">
                    {formatCostSavings(pipeline.cardinalityReduction, pipeline.inputProcessCount)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Button
                    size="small"
                    startIcon={<ExperimentIcon />}
                    onClick={() => handleStartExperiment(pipeline.hostname)}
                  >
                    New Experiment
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};
```

### Location: `/projects/dashboard/src/pages/PipelineCatalog.tsx`

```tsx
import React, { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  IconButton,
  Tabs,
  Tab,
  Paper,
} from '@mui/material';
import {
  Close as CloseIcon,
  ContentCopy as CopyIcon,
  TrendingDown,
  Speed,
  FilterList,
  Functions,
  AutoAwesome,
} from '@mui/icons-material';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { api } from '../services/api.service';

interface PipelineTemplate {
  name: string;
  version: string;
  description: string;
  strategy: string;
  expectedReduction: number;
  defaultParameters: Record<string, any>;
  yamlConfig: string;
  icon: JSX.Element;
}

const pipelineIcons: Record<string, JSX.Element> = {
  'process-baseline-v1': <Speed />,
  'process-priority-based-v1': <FilterList />,
  'process-topk-v1': <Functions />,
  'process-aggregated-v1': <TrendingDown />,
  'process-adaptive-filter-v1': <AutoAwesome />,
};

export const PipelineCatalog: React.FC = () => {
  const [pipelines, setPipelines] = useState<PipelineTemplate[]>([]);
  const [selectedPipeline, setSelectedPipeline] = useState<PipelineTemplate | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPipelineCatalog();
  }, []);

  const fetchPipelineCatalog = async () => {
    try {
      setLoading(true);
      const response = await api.get('/pipelines/process/catalog');
      const pipelinesWithIcons = response.data.map((p: PipelineTemplate) => ({
        ...p,
        icon: pipelineIcons[p.name] || <Speed />,
      }));
      setPipelines(pipelinesWithIcons);
    } catch (error) {
      console.error('Failed to fetch pipeline catalog:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = (pipeline: PipelineTemplate) => {
    setSelectedPipeline(pipeline);
    setDialogOpen(true);
    setActiveTab(0);
  };

  const handleCopyCommand = (pipeline: PipelineTemplate) => {
    const command = `phoenix pipeline deploy ${pipeline.name} --target-host <hostname>`;
    navigator.clipboard.writeText(command);
    // Show toast notification
  };

  const handleCopyYAML = () => {
    if (selectedPipeline) {
      navigator.clipboard.writeText(selectedPipeline.yamlConfig);
      // Show toast notification
    }
  };

  const getStrategyColor = (strategy: string) => {
    const colors: Record<string, string> = {
      'Priority-Based': 'primary',
      'Top-K': 'secondary',
      'Aggregation': 'success',
      'Adaptive': 'warning',
      'Baseline': 'default',
    };
    return colors[strategy] || 'default';
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Process Pipeline Catalog
        </Typography>
        <Typography variant="body1" color="textSecondary">
          Pre-validated OpenTelemetry pipeline configurations optimized for process metrics
        </Typography>
      </Box>

      <Grid container spacing={3}>
        {pipelines.map((pipeline) => (
          <Grid item xs={12} md={6} lg={4} key={pipeline.name}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      width: 48,
                      height: 48,
                      borderRadius: 1,
                      bgcolor: 'primary.light',
                      color: 'primary.main',
                      mr: 2,
                    }}
                  >
                    {pipeline.icon}
                  </Box>
                  <Box>
                    <Typography variant="h6" component="h2">
                      {pipeline.name}
                    </Typography>
                    <Typography variant="caption" color="textSecondary">
                      Version {pipeline.version}
                    </Typography>
                  </Box>
                </Box>

                <Typography variant="body2" color="textSecondary" paragraph>
                  {pipeline.description}
                </Typography>

                <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
                  <Chip
                    label={pipeline.strategy}
                    size="small"
                    color={getStrategyColor(pipeline.strategy) as any}
                  />
                  <Chip
                    icon={<TrendingDown />}
                    label={`~${pipeline.expectedReduction}% reduction`}
                    size="small"
                    variant="outlined"
                    color="success"
                  />
                </Box>

                {Object.keys(pipeline.defaultParameters).length > 0 && (
                  <Box>
                    <Typography variant="caption" color="textSecondary">
                      Configurable Parameters:
                    </Typography>
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mt: 0.5 }}>
                      {Object.keys(pipeline.defaultParameters).map((param) => (
                        <Chip key={param} label={param} size="small" variant="outlined" />
                      ))}
                    </Box>
                  </Box>
                )}
              </CardContent>

              <Box sx={{ p: 2, pt: 0, display: 'flex', gap: 1 }}>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => handleViewDetails(pipeline)}
                  fullWidth
                >
                  View Config
                </Button>
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<CopyIcon />}
                  onClick={() => handleCopyCommand(pipeline)}
                  fullWidth
                >
                  Copy CLI
                </Button>
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Pipeline Details Dialog */}
      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        maxWidth="lg"
        fullWidth
      >
        {selectedPipeline && (
          <>
            <DialogTitle>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Typography variant="h6">{selectedPipeline.name}</Typography>
                <IconButton onClick={() => setDialogOpen(false)}>
                  <CloseIcon />
                </IconButton>
              </Box>
            </DialogTitle>
            <DialogContent>
              <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 2 }}>
                <Tab label="Overview" />
                <Tab label="Configuration" />
                <Tab label="Usage" />
              </Tabs>

              {activeTab === 0 && (
                <Box>
                  <Typography variant="body1" paragraph>
                    {selectedPipeline.description}
                  </Typography>
                  <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
                    <Typography variant="subtitle2" gutterBottom>
                      Key Features
                    </Typography>
                    <ul>
                      <li>Strategy: {selectedPipeline.strategy}</li>
                      <li>Expected Cardinality Reduction: ~{selectedPipeline.expectedReduction}%</li>
                      <li>Maintains 100% critical process visibility</li>
                      <li>Optimized for New Relic Infrastructure</li>
                    </ul>
                  </Paper>
                </Box>
              )}

              {activeTab === 1 && (
                <Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                    <Typography variant="subtitle2">OpenTelemetry Configuration</Typography>
                    <Button
                      size="small"
                      startIcon={<CopyIcon />}
                      onClick={handleCopyYAML}
                    >
                      Copy YAML
                    </Button>
                  </Box>
                  <Box sx={{ maxHeight: 500, overflow: 'auto' }}>
                    <SyntaxHighlighter
                      language="yaml"
                      style={vscDarkPlus}
                      customStyle={{ margin: 0 }}
                    >
                      {selectedPipeline.yamlConfig}
                    </SyntaxHighlighter>
                  </Box>
                </Box>
              )}

              {activeTab === 2 && (
                <Box>
                  <Typography variant="subtitle2" gutterBottom>
                    Deployment Command
                  </Typography>
                  <Paper variant="outlined" sx={{ p: 2, mb: 2, bgcolor: 'grey.100' }}>
                    <code>
                      phoenix pipeline deploy {selectedPipeline.name} --target-host {'<hostname>'} --env-vars "NR_API_KEY=secret:nr-secret:api-key"
                    </code>
                  </Paper>
                  <Typography variant="subtitle2" gutterBottom>
                    Parameters
                  </Typography>
                  {Object.entries(selectedPipeline.defaultParameters).map(([key, value]) => (
                    <Box key={key} sx={{ mb: 1 }}>
                      <Typography variant="body2">
                        <strong>{key}:</strong> {String(value)} (default)
                      </Typography>
                    </Box>
                  ))}
                </Box>
              )}
            </DialogContent>
          </>
        )}
      </Dialog>
    </Box>
  );
};
```

## Testing Recommendations

### 1. Acceptance Test Implementation

```go
// tests/acceptance/process_metrics_test.go
package acceptance

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProcessPipelineDeployment(t *testing.T) {
    // AT-P01: Deploy process-baseline-v1
    t.Run("Deploy baseline pipeline", func(t *testing.T) {
        start := time.Now()
        
        // Deploy pipeline
        err := phoenixCLI.Run("pipeline", "deploy", "process-baseline-v1", 
            "--target-host", testNode,
            "--crd-name", "test-baseline")
        require.NoError(t, err)
        
        // Wait for deployment
        require.Eventually(t, func() bool {
            status, _ := phoenixCLI.GetPipelineStatus("test-baseline")
            return status.Phase == "Running"
        }, 10*time.Minute, 10*time.Second)
        
        // Verify deployment time
        assert.Less(t, time.Since(start), 10*time.Minute, "Deployment took too long")
        
        // Verify metrics flow
        metrics := queryPrometheus("otelcol_receiver_accepted_metric_points")
        assert.Greater(t, metrics["hostmetrics"], 0)
    })
}

func TestExperimentPromotion(t *testing.T) {
    // AT-P10: Experiment promotion
    t.Run("Promote winning variant", func(t *testing.T) {
        // Run experiment first
        err := phoenixCLI.Run("experiment", "create", "--scenario", "test-scenario.yaml")
        require.NoError(t, err)
        
        err = phoenixCLI.Run("experiment", "run", "test-experiment")
        require.NoError(t, err)
        
        // Wait for completion
        time.Sleep(5 * time.Minute)
        
        // Promote variant B
        start := time.Now()
        err = phoenixCLI.Run("experiment", "promote", "test-experiment", "--variant", "B")
        require.NoError(t, err)
        
        // Verify promotion completed quickly
        assert.Less(t, time.Since(start), 5*time.Minute, "Promotion took too long")
        
        // Verify cleanup
        crds := k8sClient.ListPhoenixProcessPipelines()
        assert.Equal(t, 1, len(crds), "Only promoted pipeline should remain")
    })
}
```

---

This implementation guide provides concrete examples for all critical missing components. Teams can use these as starting points and adapt them to fit the exact Phoenix Platform architecture and coding standards.