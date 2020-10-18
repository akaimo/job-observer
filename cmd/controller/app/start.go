package app

import "github.com/spf13/cobra"

func NewCommandStartJobObserverController(stopCh <-chan struct{}) *cobra.Command {
	cmd := &cobra.Command{
		Use: "job-observer-controller",
		Short: "",
		Long: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			RunJobObserverController(stopCh)
			return nil
		},
	}

	// Parse flags

	return cmd
}

func RunJobObserverController(stopCh <-chan struct{}) {
	// Run controller
}
