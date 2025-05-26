package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// LoadSimulationJobSpec defines the desired state of LoadSimulationJob
type LoadSimulationJobSpec struct {
	// ExperimentID is the ID of the experiment this load simulation belongs to
	ExperimentID string `json:"experimentId"`
	
	// Profile defines the type of load simulation to run
	Profile string `json:"profile"`
	
	// Duration specifies how long the simulation should run
	Duration string `json:"duration"`
	
	// ProcessCount is the number of processes to simulate
	ProcessCount int32 `json:"processCount"`
	
	// NodeSelector allows targeting specific nodes
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	
	// Resources defines resource requirements for the simulation
	Resources K8sResourceRequirements `json:"resources,omitempty"`
}

// K8sResourceRequirements defines compute resource requirements for Kubernetes
type K8sResourceRequirements struct {
	// CPU request and limit
	CPU string `json:"cpu,omitempty"`
	
	// Memory request and limit
	Memory string `json:"memory,omitempty"`
}

// LoadSimulationJobStatus defines the observed state of LoadSimulationJob
type LoadSimulationJobStatus struct {
	// Phase represents the current phase of the job
	Phase string `json:"phase,omitempty"`
	
	// Message provides additional details about the current state
	Message string `json:"message,omitempty"`
	
	// StartTime indicates when the simulation started
	StartTime *metav1.Time `json:"startTime,omitempty"`
	
	// CompletionTime indicates when the simulation completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	
	// ProcessesActive shows the number of active simulated processes
	ProcessesActive int32 `json:"processesActive,omitempty"`
	
	// MetricsGenerated shows the total number of metrics generated
	MetricsGenerated int64 `json:"metricsGenerated,omitempty"`
}

// LoadSimulationJob is the Schema for the loadsimulationjobs API
type LoadSimulationJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadSimulationJobSpec   `json:"spec,omitempty"`
	Status LoadSimulationJobStatus `json:"status,omitempty"`
}

// LoadSimulationJobList contains a list of LoadSimulationJob
type LoadSimulationJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadSimulationJob `json:"items"`
}

// DeepCopy methods required for Kubernetes objects
func (in *LoadSimulationJob) DeepCopy() *LoadSimulationJob {
	if in == nil {
		return nil
	}
	out := new(LoadSimulationJob)
	in.DeepCopyInto(out)
	return out
}

func (in *LoadSimulationJob) DeepCopyInto(out *LoadSimulationJob) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *LoadSimulationJob) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *LoadSimulationJobList) DeepCopy() *LoadSimulationJobList {
	if in == nil {
		return nil
	}
	out := new(LoadSimulationJobList)
	in.DeepCopyInto(out)
	return out
}

func (in *LoadSimulationJobList) DeepCopyInto(out *LoadSimulationJobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LoadSimulationJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *LoadSimulationJobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *LoadSimulationJobSpec) DeepCopyInto(out *LoadSimulationJobSpec) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Resources = in.Resources
}

func (in *LoadSimulationJobStatus) DeepCopyInto(out *LoadSimulationJobStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
}