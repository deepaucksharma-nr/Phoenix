package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix-vnext/platform/projects/phoenix-cli/internal/output"
)

var loadSimWatch bool

// loadsimStatusCmd represents the loadsim status command
var loadsimStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show status of load simulations",
	Long: `Show the status of one or all load simulations.

If no name is provided, lists all load simulations in the system.
Use the --watch flag to continuously monitor status updates.

Examples:
  # Show all load simulations
  phoenix loadsim status

  # Show specific load simulation
  phoenix loadsim status loadsim-12345678

  # Watch load simulation status
  phoenix loadsim status loadsim-12345678 --watch`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Kubernetes client
		k8sClient, err := client.GetKubernetesClient()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		ctx := context.Background()

		if len(args) == 0 {
			// List all load simulations
			return listLoadSimulations(ctx, k8sClient)
		}

		// Show specific load simulation
		name := args[0]
		
		// If the user provided an experiment ID, convert it to loadsim name
		if len(name) == 12 && name[:4] == "exp-" {
			name = fmt.Sprintf("loadsim-%s", name[4:])
		}

		if loadSimWatch {
			return watchLoadSimulation(ctx, k8sClient, name)
		}

		return showLoadSimulation(ctx, k8sClient, name)
	},
}

func listLoadSimulations(ctx context.Context, k8sClient client.Interface) error {
	list, err := k8sClient.PhoenixV1alpha1().LoadSimulationJobs("phoenix-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list load simulations: %w", err)
	}

	if len(list.Items) == 0 {
		output.Info("No load simulations found")
		return nil
	}

	headers := []string{"Name", "Experiment", "Profile", "Duration", "Status", "Active", "Age"}
	var data [][]string

	for _, job := range list.Items {
		age := time.Since(job.CreationTimestamp.Time).Round(time.Second).String()
		data = append(data, []string{
			job.Name,
			job.Spec.ExperimentID,
			job.Spec.Profile,
			job.Spec.Duration,
			string(job.Status.Phase),
			fmt.Sprintf("%d", job.Status.ActiveProcesses),
			age,
		})
	}

	output.Table(headers, data)
	return nil
}

func showLoadSimulation(ctx context.Context, k8sClient client.Interface, name string) error {
	job, err := k8sClient.PhoenixV1alpha1().LoadSimulationJobs("phoenix-system").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get load simulation: %w", err)
	}

	output.Success(fmt.Sprintf("Load Simulation: %s", job.Name))
	
	data := [][]string{
		{"Experiment ID", job.Spec.ExperimentID},
		{"Profile", job.Spec.Profile},
		{"Duration", job.Spec.Duration},
		{"Process Count", fmt.Sprintf("%d", job.Spec.ProcessCount)},
		{"Status", string(job.Status.Phase)},
		{"Active Processes", fmt.Sprintf("%d", job.Status.ActiveProcesses)},
		{"Message", job.Status.Message},
	}

	if job.Status.StartTime != nil {
		data = append(data, []string{"Start Time", job.Status.StartTime.Format(time.RFC3339)})
		elapsed := time.Since(job.Status.StartTime.Time).Round(time.Second)
		data = append(data, []string{"Elapsed", elapsed.String()})
	}

	if job.Status.CompletionTime != nil {
		data = append(data, []string{"Completion Time", job.Status.CompletionTime.Format(time.RFC3339)})
	}

	if len(job.Spec.NodeSelector) > 0 {
		selectors := ""
		for k, v := range job.Spec.NodeSelector {
			if selectors != "" {
				selectors += ", "
			}
			selectors += fmt.Sprintf("%s=%s", k, v)
		}
		data = append(data, []string{"Node Selectors", selectors})
	}

	output.Table([]string{"Field", "Value"}, data)
	return nil
}

func watchLoadSimulation(ctx context.Context, k8sClient client.Interface, name string) error {
	output.Info(fmt.Sprintf("Watching load simulation %s (press Ctrl+C to stop)...\n", name))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Show initial status
	if err := showLoadSimulation(ctx, k8sClient, name); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Println()
			if err := showLoadSimulation(ctx, k8sClient, name); err != nil {
				return err
			}
		}
	}
}

func init() {
	loadsimCmd.AddCommand(loadsimStatusCmd)
	
	loadsimStatusCmd.Flags().BoolVarP(&loadSimWatch, "watch", "w", false, 
		"Watch status updates continuously")
}