package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type JobObserverControllerOptions struct {
	ControllerOptions *ControllerOptions
}

type ControllerOptions struct {
	Kubeconfig string
	MasterURL string
}

const (
	defaultKubeconfig = ""
	defaultMasterURL = ""
)

func (o *ControllerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", defaultKubeconfig, "Path to a kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&o.MasterURL, "master", defaultMasterURL, "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func NewControllerOptions() *ControllerOptions {
	return &ControllerOptions{
		Kubeconfig: defaultKubeconfig,
		MasterURL: defaultMasterURL,
	}
}

func NewJobObserverControllerOptions() *JobObserverControllerOptions {
	o := &JobObserverControllerOptions{
		ControllerOptions: NewControllerOptions(),
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
	// Run controller
}
