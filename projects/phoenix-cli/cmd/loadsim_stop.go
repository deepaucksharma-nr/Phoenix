package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/output"
)

// loadsimStopCmd represents the loadsim stop command
var loadsimStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running load simulation",
	Long: `Stop a running load simulation by deleting its LoadSimulationJob resource.

This will gracefully terminate all simulated processes and clean up resources.

Examples:
  # Stop a load simulation
  phoenix loadsim stop loadsim-12345678

  # Stop using experiment ID
  phoenix loadsim stop exp-12345678`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// If the user provided an experiment ID, convert it to loadsim name
		if len(name) == 12 && name[:4] == "exp-" {
			name = fmt.Sprintf("loadsim-%s", name[4:])
		}

		// Create Kubernetes client
		k8sClient, err := client.GetKubernetesClient()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		// Get the current job to show its status
		ctx := context.Background()
		job, err := k8sClient.PhoenixV1alpha1().LoadSimulationJobs("phoenix-system").Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get load simulation: %w", err)
		}

		// Delete the LoadSimulationJob
		err = k8sClient.PhoenixV1alpha1().LoadSimulationJobs("phoenix-system").Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to stop load simulation: %w", err)
		}

		output.Success("Load simulation stop initiated")
		
		data := [][]string{
			{"Name", job.Name},
			{"Experiment ID", job.Spec.ExperimentID},
			{"Profile", job.Spec.Profile},
			{"Previous Status", string(job.Status.Phase)},
			{"Active Processes", fmt.Sprintf("%d", job.Status.ActiveProcesses)},
		}
		
		output.Table([]string{"Field", "Value"}, data)

		return nil
	},
}

func init() {
	loadsimCmd.AddCommand(loadsimStopCmd)
}