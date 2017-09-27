package util

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jw-s/redis-operator/pkg/operator/spec"
	v1lister "k8s.io/client-go/listers/core/v1"
	"time"
)

func GetMasterIPByName(client *redis.Client, name string) (string, error) {

	cmd := redis.NewStringSliceCmd("SENTINEL", "get-master-addr-by-name", name)

	client.Process(cmd)

	masterAddr, err := cmd.Result()

	if err != nil {
		return "", err
	}

	return masterAddr[0], err
}

func GetSeedMasterIP(podLister v1lister.PodLister, namespace, name string) (string, error) {

	var masterIP string

	if err := WaitForResourceToBeEstablished(10, func() (bool, error) {
		masterSeed, err := podLister.Pods(namespace).Get(spec.GetMasterPodName(name))

		if err != nil {
			if ResourceNotFoundError(err) {
				return false, nil
			}

			return false, err
		}

		if IsPodReady(masterSeed) && masterSeed.Status.PodIP != "" {
			masterIP = masterSeed.Status.PodIP
			return true, nil
		}

		return false, nil

	}); err != nil {
		return "", err
	}

	return masterIP, nil
}

func NewSentinelRedisClient(name string) *redis.Client {
	sentinelService := fmt.Sprintf("%s:%v", name, spec.RedisSentinelPort)

	return redis.NewClient(&redis.Options{
		Addr:        sentinelService,
		Password:    "",
		DB:          0,
		MaxRetries:  10,
		DialTimeout: time.Second * 30,
	})
}
