package app

import (
	"github.com/akaimo/job-observer/cmd/controller/app/options"
	"github.com/spf13/cobra"
)

type JobObserverControllerOptions struct {
	ControllerOptions *options.ControllerOptions
}

func NewJobObserverControllerOptions() *JobObserverControllerOptions {
	o := &JobObserverControllerOptions{
		ControllerOptions: options.NewControllerOptions(),
	}
	return o
}

func NewCommandStartJobObserverController(stopCh <-chan struct{}) *cobra.Command {
	o := NewJobObserverControllerOptions()

	cmd := &cobra.Command{
		Use: "job-observer-controller",
		Short: "",
		Long: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			o.RunJobObserverController(stopCh)
			return nil
		},
	}

	flags := cmd.Flags()
	o.ControllerOptions.AddFlags(flags)

	return cmd
}

func (o JobObserverControllerOptions) RunJobObserverController(stopCh <-chan struct{}) {
	Run(stopCh)
}
