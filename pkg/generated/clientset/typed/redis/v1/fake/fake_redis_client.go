package fake

import (
	v1 "github.com/jw-s/redis-operator/pkg/generated/clientset/typed/redis/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeRedisV1 struct {
	*testing.Fake
}

func (c *FakeRedisV1) Redises(namespace string) v1.RedisInterface {
	return &FakeRedises{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeRedisV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
