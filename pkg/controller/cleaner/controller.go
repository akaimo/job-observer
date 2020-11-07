package cleaner

import (
	"fmt"
	"time"

	cleanerv1alpha1 "github.com/akaimo/job-observer/pkg/apis/cleaner/v1alpha1"
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	cleanerscheme "github.com/akaimo/job-observer/pkg/client/clientset/versioned/scheme"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions/cleaner/v1alpha1"
	listers "github.com/akaimo/job-observer/pkg/client/listers/cleaner/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
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

const controllerAgentName = "cleaner"

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// cleanerclientset is a clientset for our own API group
	cleanerclientset clientset.Interface

	jobLister     batchlisters.JobLister
	jobSynced     cache.InformerSynced
	cleanerLister listers.CleanerLister
	cleanerSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder
}

func NewController(
	kubeclientset kubernetes.Interface,
	cleanerclientset clientset.Interface,
	jobInformer jobinformers.JobInformer,
	cleanerinformer informers.CleanerInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(cleanerscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:    kubeclientset,
		cleanerclientset: cleanerclientset,
		jobLister:        jobInformer.Lister(),
		jobSynced:        jobInformer.Informer().HasSynced,
		cleanerLister:    cleanerinformer.Lister(),
		cleanerSynced:    cleanerinformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Cleaners"),
		Recorder:         recorder,
	}

	klog.Info("Setting up event handlers")
	cleanerinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCleanerResource,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueCleanerResource(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting controller")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cleanerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)

		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	cr, err := c.cleanerLister.Cleaners(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	deleteList, err := c.getDeletableJobs(cr)
	if err != nil {
		return err
	}

	err = c.deleteJobs(deleteList)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) enqueueCleanerResource(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) getDeletableJobs(cr *cleanerv1alpha1.Cleaner) ([]*batchv1.Job, error) {
	jobs, err := c.getJobsMatchCleaner(cr)
	if err != nil {
		return nil, err
	}

	jobs, err = filterJobFromCleanerTarget(cr, jobs)
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(cr.Spec.TtlAfterFinished)
	if err != nil {
		return nil, err
	}

	var deleteList []*batchv1.Job
	for _, v := range jobs {
		if expired, err := processTTL(v, duration); err != nil {
			klog.Error(err)
			continue
		} else if !expired {
			continue
		}
		deleteList = append(deleteList, v)
	}

	return deleteList, nil
}

func (c *Controller) getJobsMatchCleaner(cr *cleanerv1alpha1.Cleaner) ([]*batchv1.Job, error) {
	selector, err := metav1.LabelSelectorAsSelector(cr.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert Cleaner selector: %v", err)
	}
	jobs, err := c.jobLister.Jobs(cr.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	var finishedJobs []*batchv1.Job
	for _, j := range jobs {
		if isJobFinished(j) {
			finishedJobs = append(finishedJobs, j)
		}
	}
	return finishedJobs, nil
}

func filterJobFromCleanerTarget(cr *cleanerv1alpha1.Cleaner, jobs []*batchv1.Job) ([]*batchv1.Job, error) {
	if cr.Spec.CleaningJobStatus == "" || cr.Spec.CleaningJobStatus == "All" {
		return jobs, nil
	}

	var targetStatus batchv1.JobConditionType
	switch cr.Spec.CleaningJobStatus {
	case "Complete":
		targetStatus = batchv1.JobComplete
	case "Failed":
		targetStatus = batchv1.JobFailed
	default:
		return nil, fmt.Errorf("Unsupport Status '%s'\n", cr.Spec.CleaningJobStatus)
	}

	return filterJobFromCondition(jobs, targetStatus), nil
}

func filterJobFromCondition(jobs []*batchv1.Job, condition batchv1.JobConditionType) []*batchv1.Job {
	var filteredJob []*batchv1.Job
	for _, j := range jobs {
		if isJobCondition(j, condition) {
			filteredJob = append(filteredJob, j)
		}
	}
	return filteredJob
}

func isJobCondition(job *batchv1.Job, condition batchv1.JobConditionType) bool {
	c := jobLastCondition(job)
	return c.Type == condition
}

func jobLastCondition(job *batchv1.Job) batchv1.JobCondition {
	n := len(job.Status.Conditions)
	return job.Status.Conditions[n-1]
}

func (c Controller) deleteJobs(jobs []*batchv1.Job) error {
	for _, v := range jobs {
		policy := metav1.DeletePropagationForeground
		options := &metav1.DeleteOptions{
			PropagationPolicy: &policy,
			Preconditions:     &metav1.Preconditions{UID: &v.UID},
		}
		klog.V(4).Infof("Cleaning up Job %s/%s", v.Namespace, v.Name)
		err := c.kubeclientset.BatchV1().Jobs(v.Namespace).Delete(v.Name, options)
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}

func processTTL(job *batchv1.Job, ttl time.Duration) (expired bool, err error) {
	if job.DeletionTimestamp != nil || !isJobFinished(job) {
		return false, nil
	}

	now := time.Now()
	t, err := timeLeft(job, &now, ttl)
	if err != nil {
		return false, err
	}

	if *t <= 0 {
		return true, nil
	}

	return false, nil
}

func isJobFinished(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func timeLeft(j *batchv1.Job, since *time.Time, ttl time.Duration) (*time.Duration, error) {
	finishAt, expireAt, err := getFinishAndExpireTime(j, ttl)
	if err != nil {
		return nil, err
	}
	if finishAt.UTC().After(since.UTC()) {
		klog.Warningf("Warning: Found Job %s/%s finished in the future. This is likely due to time skew in the cluster. Job cleanup will be deferred.", j.Namespace, j.Name)
	}
	remaining := expireAt.UTC().Sub(since.UTC())
	klog.V(4).Infof("Found Job %s/%s finished at %v, remaining TTL %v since %v, TTL will expire at %v", j.Namespace, j.Name, finishAt.UTC(), remaining, since.UTC(), expireAt.UTC())
	return &remaining, nil
}

func getFinishAndExpireTime(j *batchv1.Job, ttl time.Duration) (*time.Time, *time.Time, error) {
	if !isJobFinished(j) {
		return nil, nil, fmt.Errorf("job %s/%s should not be cleaned up", j.Namespace, j.Name)
	}
	finishAt, err := jobFinishTime(j)
	if err != nil {
		return nil, nil, err
	}
	finishAtUTC := finishAt.UTC()
	expireAtUTC := finishAtUTC.Add(ttl)
	return &finishAtUTC, &expireAtUTC, nil
}

func jobFinishTime(finishedJob *batchv1.Job) (metav1.Time, error) {
	for _, c := range finishedJob.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			finishAt := c.LastTransitionTime
			if finishAt.IsZero() {
				return metav1.Time{}, fmt.Errorf("unable to find the time when the Job %s/%s finished", finishedJob.Namespace, finishedJob.Name)
			}
			return c.LastTransitionTime, nil
		}
	}

	return metav1.Time{}, fmt.Errorf("unable to find the status of the finished Job %s/%s", finishedJob.Namespace, finishedJob.Name)
}
