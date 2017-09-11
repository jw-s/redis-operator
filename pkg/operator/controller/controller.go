package controller

import (
	"fmt"
	"github.com/JoelW-S/redis-operator/pkg/operator/client"
	"github.com/JoelW-S/redis-operator/pkg/operator/cr"
	"github.com/JoelW-S/redis-operator/pkg/operator/event"
	"github.com/JoelW-S/redis-operator/pkg/operator/processor"
	"github.com/JoelW-S/redis-operator/pkg/operator/spec"
	"github.com/sirupsen/logrus"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Namespace string

type Controller interface {
	Run(stopCh <-chan struct{})
}

type RedisController struct {
	logger *logrus.Entry
	Config
	redisClient *client.RedisServerCR
	processor   cache.ResourceEventHandler
	handler     event.Handler
	servers     map[string]*spec.RedisServer
	stopCh      chan struct{}
}

type Config struct {
	*rest.Config
	Namespace           Namespace
	kubernetesClient    kubernetes.Interface
	KubernetesExtClient *apiextensionsclient.Clientset
}

func NewConfig(cfg *rest.Config, namespace Namespace, kubernetesClient kubernetes.Interface, KubernetesExtClient *apiextensionsclient.Clientset) Config {
	return Config{
		Config:              cfg,
		Namespace:           namespace,
		kubernetesClient:    kubernetesClient,
		KubernetesExtClient: KubernetesExtClient,
	}
}

func New(cfg Config, redisClient *client.RedisServerCR) Controller {
	return &RedisController{
		logger:      logrus.WithField("pkg", "controller"),
		Config:      cfg,
		redisClient: redisClient,
		processor:   processor.RedisEventProcessor{},
		handler:     event.NewHandler(),
		servers:     make(map[string]*spec.RedisServer),
	}
}

func (c *RedisController) Run(stopChan <-chan struct{}) {

	c.createRedisServer()

	eventCh, errCh := c.watch()

	go c.HandleEvents(eventCh, errCh)

}

func (c *RedisController) HandleEvents(eventCh <-chan event.Event, errCh <-chan error) {

	for {
		select {
		case e := <-eventCh:

			c.servers[e.Object.Name] = e.Object
			c.determineEventType(e)

		case err := <-errCh:

			c.logger.Fatal(err)

		case <-c.stopCh:
			c.logger.Info("Shutting down handling of events")
			return

		}
	}

}

func (c *RedisController) determineEventType(event event.Event) {

	switch event.Type {

	case kwatch.Added:

		c.handler.OnAdd(event)

	case kwatch.Modified:

		c.handler.OnUpdate(event)

	case kwatch.Deleted:

		c.handler.OnDelete(event)

	default:

		c.logger.Fatal("Invalid event type!")

	}
}
func (c *RedisController) createRedisServer() {

	_, err := client.CreateCustomResourceDefinition(c.KubernetesExtClient)

	if err != nil {

		if cr.ResourceAlreadyExistError(err) {
			c.logger.Debug("Resource already exists")
			return
		}

		c.logger.Fatal("API error", err)

	}

}

func (c *RedisController) watch() (<-chan event.Event, <-chan error) {
	eventCh := make(chan event.Event)
	// On unexpected error case, controller should exit
	errCh := make(chan error, 1)

	watcher := cache.NewListWatchFromClient(c.redisClient.Client, spec.RedisNamePlural, v1.NamespaceAll, fields.Everything())

	_, controller := cache.NewInformer(watcher, &spec.RedisServer{}, 0, processor.New(eventCh, c.stopCh))

	go controller.Run(c.stopCh)

	return eventCh, errCh
}
