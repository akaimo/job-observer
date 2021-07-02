package app

import (
	"context"
	"fmt"
	"github.com/akaimo/job-observer/cmd/controller/app/options"
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions"
	"github.com/akaimo/job-observer/pkg/controller"
	cleanercontroller "github.com/akaimo/job-observer/pkg/controller/cleaner"
	notificatorcontroller "github.com/akaimo/job-observer/pkg/controller/notificator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"os"
	"sync"
	"time"
)

func Run(opts *options.ControllerOptions, stopCh <-chan struct{}) {
	rootCtx := contextWithStopCh(context.Background(), stopCh)

	controllers, kubeCfg, err := buildControllerContext(rootCtx, stopCh, opts)
	if err != nil {
		klog.Error(err, "error building controller context", "options", opts)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	run := func(_ context.Context) {
		for _, c := range controllers {
			wg.Add(1)
			go func(controller controller.Controller) {
				defer wg.Done()
				klog.Info("start controller")

				workers := 2
				err := controller.Run(workers, stopCh)
				if err != nil {
					klog.Fatalf("Error running controller: %s", err.Error())
				}
			}(c)
		}

		// TODO: Start SharedInformerFactories
		wg.Wait()

		klog.Info("control loops exited")
		os.Exit(0)
	}

	klog.V(2).Info("starting leader election")
	leaderElectionClient, err := kubernetes.NewForConfig(rest.AddUserAgent(kubeCfg, "leader-election"))
	if err != nil {
		klog.Error(err, "error creating leader election client")
		os.Exit(1)
	}

	// FIXME: recorder
	startLeaderElection(rootCtx, opts, leaderElectionClient, controllers[0].Recorder(), run)
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

func buildControllerContext(ctx context.Context, stopCh <-chan struct{}, opts *options.ControllerOptions) ([]controller.Controller, *rest.Config, error) {
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
	notificatorInformerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	controllers := []controller.Controller{
		cleanercontroller.NewController(kubeClient, client,
			kubeInformerFactory.Batch().V1().Jobs(),
			cleanerInformerFactory.JobObserver().V1alpha1().Cleaners(),
		),
		notificatorcontroller.NewController(kubeClient, client,
			kubeInformerFactory.Batch().V1().Jobs(),
			notificatorInformerFactory.JobObserver().V1alpha1().Notificators(),
		),
	}

	klog.V(4).Info("start shared informer factories")
	kubeInformerFactory.Start(stopCh)
	cleanerInformerFactory.Start(stopCh)
	notificatorInformerFactory.Start(stopCh)

	return controllers, kubeCfg, nil
}

func startLeaderElection(ctx context.Context, opts *options.ControllerOptions, leaderElectionClient kubernetes.Interface, recorder record.EventRecorder, run func(context.Context)) {
	// Identity used to distinguish between multiple controller manager instances
	id, err := os.Hostname()
	if err != nil {
		klog.Error(err, "error getting hostname")
		os.Exit(1)
	}

	// Lock required for leader election
	rl := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{
			Namespace: opts.LeaderElectionNamespace,
			Name:      "job-observer-controller",
		},
		Client: leaderElectionClient.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      id + "-external-job-observer-controller",
			EventRecorder: recorder,
		},
	}

	// Try and become the leader and start controller manager loops
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          &rl,
		LeaseDuration: opts.LeaderElectionLeaseDuration,
		RenewDeadline: opts.LeaderElectionRenewDeadline,
		RetryPeriod:   opts.LeaderElectionRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				klog.V(0).Info("leader election lost")
				os.Exit(1)
			},
		},
	})
}
