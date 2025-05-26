package controller

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	phoenixv1alpha1 "github.com/phoenix-vnext/platform/projects/loadsim-operator/api/v1alpha1"
)

// LoadSimulationJobReconciler reconciles a LoadSimulationJob object
type LoadSimulationJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// +kubebuilder:rbac:groups=phoenix.io,resources=loadsimulationjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=phoenix.io,resources=loadsimulationjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=phoenix.io,resources=loadsimulationjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *LoadSimulationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger.Info("reconciling LoadSimulationJob", zap.String("name", req.Name), zap.String("namespace", req.Namespace))

	// Fetch the LoadSimulationJob instance
	var loadSimJob phoenixv1alpha1.LoadSimulationJob
	if err := r.Get(ctx, req.NamespacedName, &loadSimJob); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return and don't requeue
			r.Logger.Info("LoadSimulationJob not found, assuming deleted", zap.String("name", req.Name))
			return ctrl.Result{}, nil
		}
		r.Logger.Error("failed to get LoadSimulationJob", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Check if the LoadSimulationJob is marked for deletion
	if !loadSimJob.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &loadSimJob)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&loadSimJob, "loadsimulationjob.phoenix.io/finalizer") {
		controllerutil.AddFinalizer(&loadSimJob, "loadsimulationjob.phoenix.io/finalizer")
		if err := r.Update(ctx, &loadSimJob); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle different phases
	switch loadSimJob.Status.Phase {
	case "", phoenixv1alpha1.LoadSimPhasesPending:
		return r.handlePending(ctx, &loadSimJob)
	case phoenixv1alpha1.LoadSimPhasesRunning:
		return r.handleRunning(ctx, &loadSimJob)
	case phoenixv1alpha1.LoadSimPhasesCompleted, phoenixv1alpha1.LoadSimPhasesFailed:
		// Nothing to do for completed/failed jobs
		return ctrl.Result{}, nil
	default:
		r.Logger.Warn("unknown phase", zap.String("phase", string(loadSimJob.Status.Phase)))
		return ctrl.Result{}, nil
	}
}

// handlePending creates the Kubernetes Job for the load simulation
func (r *LoadSimulationJobReconciler) handlePending(ctx context.Context, loadSimJob *phoenixv1alpha1.LoadSimulationJob) (ctrl.Result, error) {
	r.Logger.Info("handling pending LoadSimulationJob", zap.String("name", loadSimJob.Name))

	// Create the Kubernetes Job
	job := r.createJob(loadSimJob)
	
	// Set LoadSimulationJob as the owner
	if err := controllerutil.SetControllerReference(loadSimJob, job, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if job already exists
	existingJob := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, existingJob)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if errors.IsNotFound(err) {
		// Create the job
		if err := r.Create(ctx, job); err != nil {
			r.Logger.Error("failed to create Job", zap.Error(err))
			loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesFailed
			loadSimJob.Status.Message = fmt.Sprintf("Failed to create job: %v", err)
			if updateErr := r.Status().Update(ctx, loadSimJob); updateErr != nil {
				r.Logger.Error("failed to update status", zap.Error(updateErr))
			}
			return ctrl.Result{}, err
		}
		r.Logger.Info("created Job", zap.String("job", job.Name))
	}

	// Update status to Running
	loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesRunning
	loadSimJob.Status.StartTime = &metav1.Time{Time: time.Now()}
	loadSimJob.Status.ActiveProcesses = 0
	loadSimJob.Status.Message = "Load simulation started"
	
	if err := r.Status().Update(ctx, loadSimJob); err != nil {
		r.Logger.Error("failed to update status", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Requeue to check status
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// handleRunning monitors the running job and updates status
func (r *LoadSimulationJobReconciler) handleRunning(ctx context.Context, loadSimJob *phoenixv1alpha1.LoadSimulationJob) (ctrl.Result, error) {
	r.Logger.Info("handling running LoadSimulationJob", zap.String("name", loadSimJob.Name))

	// Get the associated Job
	jobName := fmt.Sprintf("loadsim-%s", loadSimJob.Name)
	job := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: loadSimJob.Namespace}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			// Job not found, mark as failed
			loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesFailed
			loadSimJob.Status.Message = "Job not found"
			loadSimJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			if updateErr := r.Status().Update(ctx, loadSimJob); updateErr != nil {
				r.Logger.Error("failed to update status", zap.Error(updateErr))
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Update active processes count
	loadSimJob.Status.ActiveProcesses = int(job.Status.Active)

	// Check job completion
	if job.Status.Succeeded > 0 {
		loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesCompleted
		loadSimJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		loadSimJob.Status.Message = "Load simulation completed successfully"
		loadSimJob.Status.ActiveProcesses = 0
	} else if job.Status.Failed > 0 {
		loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesFailed
		loadSimJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		loadSimJob.Status.Message = "Load simulation failed"
		loadSimJob.Status.ActiveProcesses = 0
	} else {
		// Still running, check duration
		if loadSimJob.Status.StartTime != nil {
			duration, err := parseDuration(loadSimJob.Spec.Duration)
			if err == nil {
				elapsed := time.Since(loadSimJob.Status.StartTime.Time)
				if elapsed > duration {
					// Duration exceeded, stop the job
					r.Logger.Info("duration exceeded, stopping job", zap.String("name", loadSimJob.Name))
					// Delete the job to stop it
					if err := r.Delete(ctx, job); err != nil {
						r.Logger.Error("failed to delete job", zap.Error(err))
					}
					loadSimJob.Status.Phase = phoenixv1alpha1.LoadSimPhasesCompleted
					loadSimJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
					loadSimJob.Status.Message = "Load simulation completed (duration reached)"
					loadSimJob.Status.ActiveProcesses = 0
				}
			}
		}
	}

	if err := r.Status().Update(ctx, loadSimJob); err != nil {
		r.Logger.Error("failed to update status", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Requeue if still running
	if loadSimJob.Status.Phase == phoenixv1alpha1.LoadSimPhasesRunning {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// handleDeletion cleans up resources when LoadSimulationJob is deleted
func (r *LoadSimulationJobReconciler) handleDeletion(ctx context.Context, loadSimJob *phoenixv1alpha1.LoadSimulationJob) (ctrl.Result, error) {
	r.Logger.Info("handling deletion of LoadSimulationJob", zap.String("name", loadSimJob.Name))

	// Clean up any running jobs
	jobName := fmt.Sprintf("loadsim-%s", loadSimJob.Name)
	job := &batchv1.Job{}
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: loadSimJob.Namespace}, job)
	if err == nil {
		// Job exists, delete it
		if err := r.Delete(ctx, job); err != nil && !errors.IsNotFound(err) {
			r.Logger.Error("failed to delete job", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(loadSimJob, "loadsimulationjob.phoenix.io/finalizer")
	if err := r.Update(ctx, loadSimJob); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createJob creates a Kubernetes Job for the load simulation
func (r *LoadSimulationJobReconciler) createJob(loadSimJob *phoenixv1alpha1.LoadSimulationJob) *batchv1.Job {
	jobName := fmt.Sprintf("loadsim-%s", loadSimJob.Name)
	
	// Build environment variables
	envVars := []corev1.EnvVar{
		{
			Name:  "EXPERIMENT_ID",
			Value: loadSimJob.Spec.ExperimentID,
		},
		{
			Name:  "PROFILE",
			Value: loadSimJob.Spec.Profile,
		},
		{
			Name:  "DURATION",
			Value: loadSimJob.Spec.Duration,
		},
		{
			Name:  "PROCESS_COUNT",
			Value: fmt.Sprintf("%d", loadSimJob.Spec.ProcessCount),
		},
	}

	// Add custom profile if specified
	if loadSimJob.Spec.Profile == "custom" && loadSimJob.Spec.CustomProfile != nil {
		// TODO: Serialize custom profile to environment or config
		envVars = append(envVars, corev1.EnvVar{
			Name:  "CHURN_RATE",
			Value: fmt.Sprintf("%f", loadSimJob.Spec.CustomProfile.ChurnRate),
		})
	}

	// Create job spec
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: loadSimJob.Namespace,
			Labels: map[string]string{
				"app":                   "phoenix-loadsim",
				"loadsimulationjob":     loadSimJob.Name,
				"experiment":            loadSimJob.Spec.ExperimentID,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                   "phoenix-loadsim",
						"loadsimulationjob":     loadSimJob.Name,
						"experiment":            loadSimJob.Spec.ExperimentID,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "loadsim",
							Image: "phoenix/load-simulator:latest",
							Env:   envVars,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					NodeSelector: loadSimJob.Spec.NodeSelector,
				},
			},
		},
	}

	return job
}

// parseDuration parses duration string like "1h" or "30m"
func parseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// SetupWithManager sets up the controller with the Manager
func (r *LoadSimulationJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&phoenixv1alpha1.LoadSimulationJob{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}