package main

import (
	"github.com/JoelW-S/redis-operator/pkg/operator/client"
	"github.com/JoelW-S/redis-operator/pkg/operator/controller"
	"github.com/JoelW-S/redis-operator/pkg/operator/util"
	"github.com/sirupsen/logrus"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	doneChan := make(chan struct{})

	config, err := util.InClusterConfig()
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic("placeholder")
	}

	extClient, err := apiextensionsclient.NewForConfig(config)

	if err != nil {
		panic("placeholder")
	}

	redisClient, err := client.NewClient(config)

	if err != nil {
		panic("placeholder")
	}

	controllerConfig := controller.NewConfig(config, "Default", kubeClient, extClient)

	c := controller.New(controllerConfig, redisClient)

	c.Run(doneChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			logrus.Error("Shutdown signal received, exiting...")
			close(doneChan)
			os.Exit(0)
		}
	}

}
