package notificator

import (
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	cleanerscheme "github.com/akaimo/job-observer/pkg/client/clientset/versioned/scheme"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions/jobobserver/v1alpha1"
	listers "github.com/akaimo/job-observer/pkg/client/listers/jobobserver/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	jobinformers "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
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

func NewController(
	kubeclientset kubernetes.Interface,
	notificatorclientset clientset.Interface,
	jobInformer jobinformers.JobInformer,
	notificatorinformer informers.NotificatorInformer) *Controller {

	utilruntime.Must(cleanerscheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:        kubeclientset,
		notificatorclientset: notificatorclientset,
		jobLister:            jobInformer.Lister(),
		jobSynced:            jobInformer.Informer().HasSynced,
		notificatorLister:    notificatorinformer.Lister(),
		notificatorSynced:    notificatorinformer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Notificator"),
		Recorder:             recorder,
	}

	notificatorinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNotificatorResource,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueNotificatorResource(new)
		},
	})

	return controller
}

func (c *Controller) enqueueNotificatorResource(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
