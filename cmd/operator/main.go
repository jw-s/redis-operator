package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	client "github.com/jw-s/redis-operator/pkg/generated/clientset"
	redisinformers "github.com/jw-s/redis-operator/pkg/generated/informers/externalversions"
	"github.com/jw-s/redis-operator/pkg/operator/controller"
	"github.com/jw-s/redis-operator/pkg/operator/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

var (
	Resync time.Duration = time.Minute
)

func main() {

	doneChan := make(chan struct{})

	config, err := util.InClusterConfig()
	if err != nil {
		panic("Could not create In-cluster config: " + err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic("Could not create kube client: " + err.Error())
	}

	redisClient, err := client.NewForConfig(config)

	if err != nil {
		panic("Could not create redis client: " + err.Error())
	}

	controllerConfig := controller.NewConfig(config, Resync)

	redisInformerFactory := redisinformers.NewSharedInformerFactory(redisClient, Resync)

	informerFactory := informers.NewSharedInformerFactory(kubeClient, Resync)

	c := controller.New(controllerConfig,
		kubeClient,
		redisClient.RedisV1(),
		redisInformerFactory.Redis().V1().Redises(),
		informerFactory.Core().V1().Pods(),
		informerFactory.Apps().V1beta1().Deployments(),
		informerFactory.Core().V1().Services(),
		informerFactory.Core().V1().Endpoints(),
		informerFactory.Core().V1().ConfigMaps())

	go c.Run(doneChan)

	go informerFactory.Start(doneChan)

	go redisInformerFactory.Start(doneChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			logrus.Info("Shutdown signal received, exiting...")
			close(doneChan)
			os.Exit(0)
		}
	}

}
