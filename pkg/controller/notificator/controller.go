package notificator

import (
	"bytes"
	"encoding/json"
	"fmt"
	jobobserverv1alpha1 "github.com/akaimo/job-observer/pkg/apis/jobobserver/v1alpha1"
	clientset "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	notificatorscheme "github.com/akaimo/job-observer/pkg/client/clientset/versioned/scheme"
	informers "github.com/akaimo/job-observer/pkg/client/informers/externalversions/jobobserver/v1alpha1"
	listers "github.com/akaimo/job-observer/pkg/client/listers/jobobserver/v1alpha1"
	"github.com/akaimo/job-observer/pkg/controller"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	jobinformers "k8s.io/client-go/informers/batch/v1"
	secretinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	secretlisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"net/http"
	"time"
)

const controllerAgentName = "notificator"

type Controller struct {
	kubeclientset        kubernetes.Interface
	notificatorclientset clientset.Interface

	jobLister         batchlisters.JobLister
	jobSynced         cache.InformerSynced
	secretLister      secretlisters.SecretLister
	secretSynced      cache.InformerSynced
	notificatorLister listers.NotificatorLister
	notificatorSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

func NewController(
	kubeclientset kubernetes.Interface,
	notificatorclientset clientset.Interface,
	jobInformer jobinformers.JobInformer,
	secretInformer secretinformers.SecretInformer,
	notificatorinformer informers.NotificatorInformer) *Controller {

	utilruntime.Must(notificatorscheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:        kubeclientset,
		notificatorclientset: notificatorclientset,
		jobLister:            jobInformer.Lister(),
		jobSynced:            jobInformer.Informer().HasSynced,
		secretLister:         secretInformer.Lister(),
		secretSynced:         secretInformer.Informer().HasSynced,
		notificatorLister:    notificatorinformer.Lister(),
		notificatorSynced:    notificatorinformer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Notificator"),
		recorder:             recorder,
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

func (c *Controller) Recorder() record.EventRecorder {
	return c.recorder
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	klog.Info("Starting Notificator controller")

	if ok := cache.WaitForCacheSync(stopCh, c.notificatorSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Shutting down Notificator controller")
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

	n, err := c.notificatorLister.Notificators(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	noticeList, err := c.getNoticeJob(n)
	if err != nil {
		return err
	}

	return c.notice(n, noticeList)
}

func (c *Controller) getNoticeJob(n *jobobserverv1alpha1.Notificator) ([]*batchv1.Job, error) {
	selector, err := metav1.LabelSelectorAsSelector(n.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert Notificator selector: %v", err)
	}

	jobs, err := c.jobLister.Jobs(n.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var noticeJob []*batchv1.Job
	for _, j := range jobs {
		if isNotice, err := isNotice(n, j); err != nil {
			klog.Error(err)
			continue
		} else if isNotice {
			noticeJob = append(noticeJob, j)
		}
	}

	return noticeJob, nil
}

func (c *Controller) notice(n *jobobserverv1alpha1.Notificator, js []*batchv1.Job) error {
	return c.noticeSlack(n, js)
}

type SlackMessage struct {
	Channel     string       `json:"channel"`
	Name        string       `json:"username"`
	IconEmoji   string       `json:"icon_emoji"`
	IconURL     string       `json:"icon_url"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Color string `json:"color"`
}

func (c *Controller) noticeSlack(n *jobobserverv1alpha1.Notificator, js []*batchv1.Job) error {
	s, err := c.secretLister.Secrets(n.Namespace).Get(n.Spec.Receiver.SlackConfig.ApiURL.Name)
	if err != nil {
		return err
	}
	if _, found := s.Data[n.Spec.Receiver.SlackConfig.ApiURL.Key]; !found {
		return fmt.Errorf("key %q in secret %q not found", n.Spec.Receiver.SlackConfig.ApiURL.Key, n.Spec.Receiver.SlackConfig.ApiURL.Name)
	}
	webhookURL := string(s.Data[n.Spec.Receiver.SlackConfig.ApiURL.Key])
	message := SlackMessage{
		Channel:   n.Spec.Receiver.SlackConfig.Channel,
		Name:      n.Spec.Receiver.SlackConfig.Username,
		IconEmoji: n.Spec.Receiver.SlackConfig.IconEmoji,
		IconURL:   n.Spec.Receiver.SlackConfig.IconURL,
		Attachments: []Attachment{
			{
				Title: fmt.Sprintf("[NOTICE] job-observer:%s/%s", n.Namespace, n.Name),
				Text:  jobNameList(js),
				Color: "danger",
			},
		},
	}
	j, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		webhookURL,
		bytes.NewBuffer(j),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func jobNameList(js []*batchv1.Job) string {
	var name string
	for i, v := range js {
		if i == 0 {
			name = fmt.Sprintf("%s/%s", v.Namespace, v.Name)
		} else {
			name = fmt.Sprintf("%s\n%s/%s", name, v.Namespace, v.Name)
		}
	}
	return name
}

func isNotice(n *jobobserverv1alpha1.Notificator, job *batchv1.Job) (bool, error) {
	if controller.IsJobFinished(job) {
		return false, nil
	}

	deadline, err := time.ParseDuration(n.Spec.Rule.FinishingDeadline)
	if err != nil {
		return false, err
	}
	runTime := jobRunningTime(job, time.Now())

	return runTime > deadline, nil
}

func jobRunningTime(job *batchv1.Job, now time.Time) time.Duration {
	return now.Sub(job.GetCreationTimestamp().Time)
}