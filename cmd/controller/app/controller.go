package app

import (
	"context"
	"fmt"
	"github.com/akaimo/job-observer/cmd/controller/app/options"
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions"
	"github.com/akaimo/job-observer/pkg/controller/cleaner"
	cleanercontroller "github.com/akaimo/job-observer/pkg/controller/cleaner"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"time"
)

func Run(opts *options.ControllerOptions, stopCh <-chan struct{}) {
	rootCtx := contextWithStopCh(context.Background(), stopCh)

	_, _, err := buildControllerContext(rootCtx, stopCh, opts)
	if err != nil {
		klog.Error(err, "error building controller context", "options", opts)
		os.Exit(1)
	}
}

func contextWithStopCh(ctx context.Context, stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
		case <-stopCh:
		}
	}()
	return ctx
}

func buildControllerContext(ctx context.Context, stopCh <-chan struct{}, opts *options.ControllerOptions) (*cleaner.Controller, *rest.Config, error) {
	kubeCfg, err := clientcmd.BuildConfigFromFlags(opts.MasterURL, opts.Kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating rest config: %s", err.Error())
	}

	kubeCfg = rest.AddUserAgent(kubeCfg, "akaimo/job-observer")

	// Create job-observer api client
	client, err := clientset.NewForConfig(kubeCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building example clientset: %s", err.Error())
	}

	// Create kubernetes api client
	kubeClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	cleanerInformerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	controller := cleanercontroller.NewController(kubeClient, client,
		kubeInformerFactory.Batch().V1().Jobs(),
		cleanerInformerFactory.JobObserver().V1alpha1().Cleaners())

	return controller, kubeCfg, nil
}
