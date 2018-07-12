package controller

import (
	"time"

	redisclient "github.com/jw-s/redis-operator/pkg/generated/clientset/typed/redis/v1"
	redisinformer "github.com/jw-s/redis-operator/pkg/generated/informers/externalversions/redis/v1"
	redislister "github.com/jw-s/redis-operator/pkg/generated/listers/redis/v1"
	"github.com/jw-s/redis-operator/pkg/operator/redis"
	"github.com/jw-s/redis-operator/pkg/operator/spec"
	"github.com/jw-s/redis-operator/pkg/operator/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	v1beta1informer "k8s.io/client-go/informers/apps/v1beta1"
	v1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	v1beta1Lister "k8s.io/client-go/listers/apps/v1beta1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller interface {
	Run(<-chan struct{})
}

type RedisController struct {
	logger *logrus.Entry
	Config
	kubernetesClient  kubernetes.Interface
	redisClient       redisclient.RedisesGetter
	cacheSyncs        []cache.InformerSynced
	redisLister       redislister.RedisLister
	podLister         v1lister.PodLister
	deploymentLister  v1beta1Lister.DeploymentLister
	serviceLister     v1lister.ServiceLister
	endpointsLister   v1lister.EndpointsLister
	configMapLister   v1lister.ConfigMapLister
	statefulSetLister v1beta1Lister.StatefulSetLister
	queue             workqueue.RateLimitingInterface
	redises           map[string]*redis.Redis
}

type Config struct {
	*rest.Config
	DefaultResync time.Duration
}

func NewConfig(cfg *rest.Config,
	defaultResync time.Duration) Config {
	return Config{
		Config:        cfg,
		DefaultResync: defaultResync,
	}
}

func New(cfg Config,
	kubernetesClient kubernetes.Interface,
	redisClient redisclient.RedisesGetter,
	redisInformer redisinformer.RedisInformer,
	podInformer v1informer.PodInformer,
	deploymentInformer v1beta1informer.DeploymentInformer,
	serviceInformer v1informer.ServiceInformer,
	endpointsInformer v1informer.EndpointsInformer,
	configMapInformer v1informer.ConfigMapInformer,
	statefulSetInformer v1beta1informer.StatefulSetInformer) Controller {

	cacheSyncs := []cache.InformerSynced{
		redisInformer.Informer().HasSynced,
		podInformer.Informer().HasSynced,
		deploymentInformer.Informer().HasSynced,
		serviceInformer.Informer().HasSynced,
		endpointsInformer.Informer().HasSynced,
		configMapInformer.Informer().HasSynced,
		statefulSetInformer.Informer().HasSynced,
	}

	c := &RedisController{
		logger:            logrus.WithField("pkg", "controller"),
		Config:            cfg,
		kubernetesClient:  kubernetesClient,
		redisClient:       redisClient,
		cacheSyncs:        cacheSyncs,
		redisLister:       redisInformer.Lister(),
		podLister:         podInformer.Lister(),
		deploymentLister:  deploymentInformer.Lister(),
		serviceLister:     serviceInformer.Lister(),
		endpointsLister:   endpointsInformer.Lister(),
		configMapLister:   configMapInformer.Lister(),
		statefulSetLister: statefulSetInformer.Lister(),
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		redises:           make(map[string]*redis.Redis),
	}

	redisInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					c.logger.WithError(err).Fatal("Add")
					return
				}
				c.queue.Add(key)
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					c.logger.WithError(err).Fatal("Delete")
					return
				}
				c.queue.Add(key)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(newObj)
				if err != nil {
					c.logger.WithError(err).Fatal("Update")
					return
				}
				c.queue.Add(key)
			},
		})

	return c
}

func (c *RedisController) processQueue() bool {
	k, shutdown := c.queue.Get()

	if shutdown {
		c.logger.Info("Shutting down queue")
		return false
	}

	defer c.queue.Done(k)

	logKeyEntry := c.logger.WithField("key", k)

	logKeyEntry.Debug("Working on queue item")

	err := c.process(k.(string))

	if err == nil {
		logKeyEntry.Debug("Finished Working on queue item")
		c.queue.Forget(k)
		return true
	}

	logKeyEntry.
		WithError(err).
		Error("Re-queueing as encountered an error")

	c.queue.AddRateLimited(k)

	return true
}

func (c *RedisController) process(key string) error {

	ns, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		return err
	}

	obj, err := c.redisLister.Redises(ns).Get(name)

	if err != nil {

		if util.ResourceNotFoundError(err) {
			delete(c.redises, name)
			return c.deleteResources(ns, name)
		}

		return err
	}

	redisCopy := obj.DeepCopy()

	myRedis, exists := c.redises[name]

	if !exists {

		cfg := redis.Config{
			RedisCRClient: c.redisClient,
			RedisClient: util.
				NewSentinelRedisClient(spec.GetSentinelServiceName(name)),
		}

		newRedis := redis.New(cfg, redisCopy)

		c.redises[name] = newRedis

		if err := newRedis.ReportCreating(); err != nil {
			return err
		}

		if err := c.reconcile(newRedis); err != nil {

			return errors.NewAggregate([]error{
				err,
				newRedis.ReportFailed(),
			})
		}

		return newRedis.ReportRunning()

	}

	myRedis.Redis = redisCopy

	if err := c.reconcile(myRedis); err != nil {
		return errors.NewAggregate([]error{
			err,
			myRedis.ReportFailed(),
		})
	}
	if err := myRedis.MarkReadyCondition(); err != nil {
		return errors.NewAggregate([]error{
			err,
			myRedis.ReportFailed(),
		})
	}

	return myRedis.ReportRunning()

}

func (c *RedisController) workOnQueue() {
	for c.processQueue() {
	}
}
func (c *RedisController) Run(stopCh <-chan struct{}) {

	c.logger.Info("Starting controller")

	defer c.logger.Info("Exiting controller")

	if !cache.WaitForCacheSync(stopCh, c.cacheSyncs...) {
		c.logger.Fatal("Timeout waiting for cache to sync")
	}

	c.logger.Info("Sync completed")

	wait.Until(c.workOnQueue, time.Second, stopCh)
}
