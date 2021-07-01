package notificator

import (
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	listers "github.com/akaimo/job-observer/pkg/client/listers/jobobserver/v1alpha1"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "notificator"

type Controller struct {
	kubeclientset        kubernetes.Interface
	notificatorclientset clientset.Interface

	jobLister         batchlisters.JobLister
	jobSynced         cache.InformerSynced
	notificatorLister listers.NotificatorLister
	notificatorSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	Recorder  record.EventRecorder
}
