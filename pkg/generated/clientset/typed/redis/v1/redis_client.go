package v1

import (
	v1 "github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	"github.com/jw-s/redis-operator/pkg/generated/clientset/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type RedisV1Interface interface {
	RESTClient() rest.Interface
	RedisesGetter
}

// RedisV1Client is used to interact with features provided by the redis.operator.joelws.com group.
type RedisV1Client struct {
	restClient rest.Interface
}

func (c *RedisV1Client) Redises(namespace string) RedisInterface {
	return newRedises(c, namespace)
}

// NewForConfig creates a new RedisV1Client for the given config.
func NewForConfig(c *rest.Config) (*RedisV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &RedisV1Client{client}, nil
}

// NewForConfigOrDie creates a new RedisV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *RedisV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new RedisV1Client for the given RESTClient.
func New(c rest.Interface) *RedisV1Client {
	return &RedisV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *RedisV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
