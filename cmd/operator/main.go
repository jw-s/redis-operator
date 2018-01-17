package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"
	client "github.com/jw-s/redis-operator/pkg/generated/clientset"
	redisinformers "github.com/jw-s/redis-operator/pkg/generated/informers/externalversions"
	"github.com/jw-s/redis-operator/pkg/operator/controller"
	"github.com/jw-s/redis-operator/pkg/operator/util"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

var (
	name        = "redis-operator"
	commit      string
	version     string
	description = "The Redis operator manages redis servers deployed to Kubernetes and automates tasks related to operating a Redis server setup."
	resync      time.Duration
	debug       bool
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Action = func(c *cli.Context) { run() }
	app.Version = fmt.Sprintf("%s (%s)", version, commit)
	app.Flags = []cli.Flag{
		cli.DurationFlag{
			Name:        "resync,re",
			Usage:       "Time between syncs for informers",
			Destination: &resync,
			Value:       time.Minute,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "Toggle on Debug level for logs",
			Destination: &debug,
		},
	}
	app.Run(os.Args)
}

func run() {

	toggleDebug(debug)

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

	controllerConfig := controller.NewConfig(config, resync)

	redisInformerFactory := redisinformers.NewSharedInformerFactory(redisClient, resync)

	informerFactory := informers.NewSharedInformerFactory(kubeClient, resync)

	c := controller.New(controllerConfig,
		kubeClient,
		redisClient.RedisV1(),
		redisInformerFactory.Redis().V1().Redises(),
		informerFactory.Core().V1().Pods(),
		informerFactory.Apps().V1beta1().Deployments(),
		informerFactory.Core().V1().Services(),
		informerFactory.Core().V1().Endpoints(),
		informerFactory.Core().V1().ConfigMaps(),
		informerFactory.Apps().V1beta1().StatefulSets())

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

func toggleDebug(toggleDebugMode bool) {
	if toggleDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
