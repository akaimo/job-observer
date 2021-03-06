// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	cleanerv1alpha1 "github.com/akaimo/job-observer/pkg/apis/cleaner/v1alpha1"
	versioned "github.com/akaimo/job-observer/pkg/client/clientset/versioned"
	internalinterfaces "github.com/akaimo/job-observer/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/akaimo/job-observer/pkg/client/listers/cleaner/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CleanerInformer provides access to a shared informer and lister for
// Cleaners.
type CleanerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CleanerLister
}

type cleanerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewCleanerInformer constructs a new informer for Cleaner type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCleanerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCleanerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredCleanerInformer constructs a new informer for Cleaner type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCleanerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.JobObserverV1alpha1().Cleaners(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.JobObserverV1alpha1().Cleaners(namespace).Watch(options)
			},
		},
		&cleanerv1alpha1.Cleaner{},
		resyncPeriod,
		indexers,
	)
}

func (f *cleanerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCleanerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cleanerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cleanerv1alpha1.Cleaner{}, f.defaultInformer)
}

func (f *cleanerInformer) Lister() v1alpha1.CleanerLister {
	return v1alpha1.NewCleanerLister(f.Informer().GetIndexer())
}
