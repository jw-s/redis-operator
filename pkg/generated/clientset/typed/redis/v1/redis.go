package v1

import (
	v1 "github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	scheme "github.com/jw-s/redis-operator/pkg/generated/clientset/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RedisesGetter has a method to return a RedisInterface.
// A group's client should implement this interface.
type RedisesGetter interface {
	Redises(namespace string) RedisInterface
}

// RedisInterface has methods to work with Redis resources.
type RedisInterface interface {
	Create(*v1.Redis) (*v1.Redis, error)
	Update(*v1.Redis) (*v1.Redis, error)
	UpdateStatus(*v1.Redis) (*v1.Redis, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Redis, error)
	List(opts meta_v1.ListOptions) (*v1.RedisList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Redis, err error)
	RedisExpansion
}

// redises implements RedisInterface
type redises struct {
	client rest.Interface
	ns     string
}

// newRedises returns a Redises
func newRedises(c *RedisV1Client, namespace string) *redises {
	return &redises{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the redis, and returns the corresponding redis object, and an error if there is any.
func (c *redises) Get(name string, options meta_v1.GetOptions) (result *v1.Redis, err error) {
	result = &v1.Redis{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("redises").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Redises that match those selectors.
func (c *redises) List(opts meta_v1.ListOptions) (result *v1.RedisList, err error) {
	result = &v1.RedisList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("redises").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested redises.
func (c *redises) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("redises").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a redis and creates it.  Returns the server's representation of the redis, and an error, if there is any.
func (c *redises) Create(redis *v1.Redis) (result *v1.Redis, err error) {
	result = &v1.Redis{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("redises").
		Body(redis).
		Do().
		Into(result)
	return
}

// Update takes the representation of a redis and updates it. Returns the server's representation of the redis, and an error, if there is any.
func (c *redises) Update(redis *v1.Redis) (result *v1.Redis, err error) {
	result = &v1.Redis{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("redises").
		Name(redis.Name).
		Body(redis).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *redises) UpdateStatus(redis *v1.Redis) (result *v1.Redis, err error) {
	result = &v1.Redis{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("redises").
		Name(redis.Name).
		SubResource("status").
		Body(redis).
		Do().
		Into(result)
	return
}

// Delete takes name of the redis and deletes it. Returns an error if one occurs.
func (c *redises) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("redises").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *redises) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("redises").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched redis.
func (c *redises) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Redis, err error) {
	result = &v1.Redis{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("redises").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
