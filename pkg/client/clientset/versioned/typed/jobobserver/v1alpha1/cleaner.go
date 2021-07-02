// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/akaimo/job-observer/pkg/apis/jobobserver/v1alpha1"
	scheme "github.com/akaimo/job-observer/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CleanersGetter has a method to return a CleanerInterface.
// A group's client should implement this interface.
type CleanersGetter interface {
	Cleaners(namespace string) CleanerInterface
}

// CleanerInterface has methods to work with Cleaner resources.
type CleanerInterface interface {
	Create(*v1alpha1.Cleaner) (*v1alpha1.Cleaner, error)
	Update(*v1alpha1.Cleaner) (*v1alpha1.Cleaner, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Cleaner, error)
	List(opts v1.ListOptions) (*v1alpha1.CleanerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cleaner, err error)
	CleanerExpansion
}

// cleaners implements CleanerInterface
type cleaners struct {
	client rest.Interface
	ns     string
}

// newCleaners returns a Cleaners
func newCleaners(c *JobObserverV1alpha1Client, namespace string) *cleaners {
	return &cleaners{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cleaner, and returns the corresponding cleaner object, and an error if there is any.
func (c *cleaners) Get(name string, options v1.GetOptions) (result *v1alpha1.Cleaner, err error) {
	result = &v1alpha1.Cleaner{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cleaners").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Cleaners that match those selectors.
func (c *cleaners) List(opts v1.ListOptions) (result *v1alpha1.CleanerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CleanerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cleaners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cleaners.
func (c *cleaners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cleaners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a cleaner and creates it.  Returns the server's representation of the cleaner, and an error, if there is any.
func (c *cleaners) Create(cleaner *v1alpha1.Cleaner) (result *v1alpha1.Cleaner, err error) {
	result = &v1alpha1.Cleaner{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cleaners").
		Body(cleaner).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cleaner and updates it. Returns the server's representation of the cleaner, and an error, if there is any.
func (c *cleaners) Update(cleaner *v1alpha1.Cleaner) (result *v1alpha1.Cleaner, err error) {
	result = &v1alpha1.Cleaner{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cleaners").
		Name(cleaner.Name).
		Body(cleaner).
		Do().
		Into(result)
	return
}

// Delete takes name of the cleaner and deletes it. Returns an error if one occurs.
func (c *cleaners) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cleaners").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cleaners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cleaners").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cleaner.
func (c *cleaners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cleaner, err error) {
	result = &v1alpha1.Cleaner{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cleaners").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}