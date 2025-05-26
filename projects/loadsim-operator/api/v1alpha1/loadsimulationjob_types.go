package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadSimulationJobSpec defines the desired state of LoadSimulationJob
type LoadSimulationJobSpec struct {
	// ExperimentID is the ID of the experiment this load simulation is for
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^exp-[a-z0-9]{8}$`
	ExperimentID string `json:"experimentID"`

	// Profile defines the load simulation profile to use
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=realistic;high-cardinality;process-churn;custom
	Profile string `json:"profile"`

	// Duration specifies how long the load simulation should run
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[0-9]+[hm]$`
	Duration string `json:"duration"`

	// ProcessCount is the number of processes to simulate
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=10000
	ProcessCount int32 `json:"processCount,omitempty"`

	// NodeSelector for pod assignment
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// CustomProfile defines custom load patterns when Profile is "custom"
	// +optional
	CustomProfile *CustomProfile `json:"customProfile,omitempty"`
}

// CustomProfile defines custom load simulation patterns
type CustomProfile struct {
	// Patterns defines the process patterns to simulate
	Patterns []ProcessPattern `json:"patterns,omitempty"`

	// ChurnRate defines the rate at which processes are created/destroyed (0-1)
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	ChurnRate float64 `json:"churnRate,omitempty"`
}

// ProcessPattern defines a pattern for simulating processes
type ProcessPattern struct {
	// NameTemplate is the template for process names
	NameTemplate string `json:"nameTemplate,omitempty"`

	// CPUPattern defines the CPU usage pattern
	// +kubebuilder:validation:Enum=steady;spiky;growing;random
	CPUPattern string `json:"cpuPattern,omitempty"`

	// MemPattern defines the memory usage pattern
	// +kubebuilder:validation:Enum=steady;spiky;growing;random
	MemPattern string `json:"memPattern,omitempty"`

	// Lifetime defines how long processes of this pattern should live
	Lifetime string `json:"lifetime,omitempty"`

	// Count is the number of processes following this pattern
	Count int32 `json:"count,omitempty"`
}

// LoadSimPhases represents the phases of a load simulation job
type LoadSimPhases string

const (
	// LoadSimPhasesPending means the job hasn't started yet
	LoadSimPhasesPending LoadSimPhases = "Pending"
	// LoadSimPhasesRunning means the job is currently running
	LoadSimPhasesRunning LoadSimPhases = "Running"
	// LoadSimPhasesCompleted means the job completed successfully
	LoadSimPhasesCompleted LoadSimPhases = "Completed"
	// LoadSimPhasesFailed means the job failed
	LoadSimPhasesFailed LoadSimPhases = "Failed"
)

// LoadSimulationJobStatus defines the observed state of LoadSimulationJob
type LoadSimulationJobStatus struct {
	// Phase represents the current phase of the load simulation
	// +optional
	Phase LoadSimPhases `json:"phase,omitempty"`

	// StartTime is when the load simulation started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the load simulation completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// ActiveProcesses is the current number of active simulated processes
	// +optional
	ActiveProcesses int `json:"activeProcesses,omitempty"`

	// Message provides additional information about the current status
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lsj
// +kubebuilder:printcolumn:name="Experiment",type=string,JSONPath=`.spec.experimentID`
// +kubebuilder:printcolumn:name="Profile",type=string,JSONPath=`.spec.profile`
// +kubebuilder:printcolumn:name="Duration",type=string,JSONPath=`.spec.duration`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Active",type=integer,JSONPath=`.status.activeProcesses`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// LoadSimulationJob is the Schema for the loadsimulationjobs API
type LoadSimulationJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadSimulationJobSpec   `json:"spec,omitempty"`
	Status LoadSimulationJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadSimulationJobList contains a list of LoadSimulationJob
type LoadSimulationJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadSimulationJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoadSimulationJob{}, &LoadSimulationJobList{})
}