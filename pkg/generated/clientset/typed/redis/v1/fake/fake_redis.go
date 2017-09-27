package fake

import (
	redis_v1 "github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRedises implements RedisInterface
type FakeRedises struct {
	Fake *FakeRedisV1
	ns   string
}

var redisesResource = schema.GroupVersionResource{Group: "redis.operator.joelws.com", Version: "v1", Resource: "redises"}

var redisesKind = schema.GroupVersionKind{Group: "redis.operator.joelws.com", Version: "v1", Kind: "Redis"}

// Get takes name of the redis, and returns the corresponding redis object, and an error if there is any.
func (c *FakeRedises) Get(name string, options v1.GetOptions) (result *redis_v1.Redis, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(redisesResource, c.ns, name), &redis_v1.Redis{})

	if obj == nil {
		return nil, err
	}
	return obj.(*redis_v1.Redis), err
}

// List takes label and field selectors, and returns the list of Redises that match those selectors.
func (c *FakeRedises) List(opts v1.ListOptions) (result *redis_v1.RedisList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(redisesResource, redisesKind, c.ns, opts), &redis_v1.RedisList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &redis_v1.RedisList{}
	for _, item := range obj.(*redis_v1.RedisList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested redises.
func (c *FakeRedises) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(redisesResource, c.ns, opts))

}

// Create takes the representation of a redis and creates it.  Returns the server's representation of the redis, and an error, if there is any.
func (c *FakeRedises) Create(redis *redis_v1.Redis) (result *redis_v1.Redis, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(redisesResource, c.ns, redis), &redis_v1.Redis{})

	if obj == nil {
		return nil, err
	}
	return obj.(*redis_v1.Redis), err
}

// Update takes the representation of a redis and updates it. Returns the server's representation of the redis, and an error, if there is any.
func (c *FakeRedises) Update(redis *redis_v1.Redis) (result *redis_v1.Redis, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(redisesResource, c.ns, redis), &redis_v1.Redis{})

	if obj == nil {
		return nil, err
	}
	return obj.(*redis_v1.Redis), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRedises) UpdateStatus(redis *redis_v1.Redis) (*redis_v1.Redis, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(redisesResource, "status", c.ns, redis), &redis_v1.Redis{})

	if obj == nil {
		return nil, err
	}
	return obj.(*redis_v1.Redis), err
}

// Delete takes name of the redis and deletes it. Returns an error if one occurs.
func (c *FakeRedises) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(redisesResource, c.ns, name), &redis_v1.Redis{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRedises) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(redisesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &redis_v1.RedisList{})
	return err
}

// Patch applies the patch and returns the patched redis.
func (c *FakeRedises) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *redis_v1.Redis, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(redisesResource, c.ns, name, data, subresources...), &redis_v1.Redis{})

	if obj == nil {
		return nil, err
	}
	return obj.(*redis_v1.Redis), err
}
