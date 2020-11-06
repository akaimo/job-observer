package main

import (
	"flag"
	"github.com/akaimo/job-observer/cmd/controller/app"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	klog.InitFlags(nil)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := setupSignalHandler()

	cmd := app.NewCommandStartJobObserverController(stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	flag.CommandLine.Parse([]string{})
	if err := cmd.Execute(); err != nil {
		klog.Error(err, "error executing command")
		os.Exit(1)
	}
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
