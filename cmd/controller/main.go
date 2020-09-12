package main

import (
	"flag"
	"github.com/akaimo/job-observer/cmd/controller/app"
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := setupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	cleanerInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	controller := app.NewController(kubeClient, exampleClient,
		kubeInformerFactory.Batch().V1().Jobs(),
		cleanerInformerFactory.JobObserver().V1alpha1().Cleaners())

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	cleanerInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
var onlyOneSignalHandler = make(chan struct{})

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}
