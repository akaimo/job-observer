package options

import "github.com/spf13/pflag"

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
