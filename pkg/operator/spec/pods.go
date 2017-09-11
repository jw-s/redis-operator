package spec

import (
	"k8s.io/client-go/pkg/api/v1"
)

const (
	RedisName = "redis"
)

func redisContainer(commands, baseImage, version string, containerPorts ...v1.ContainerPort) v1.Container {

	c := v1.Container{
		Command: []string{"/bin/sh", "-ec", commands},
		Name:    "redis",
		Image:   baseImage + ":" + version,
		Ports:   containerPorts,
	}

	return c
}

func redisSentinel(commands, baseImage, version string) v1.Container {

	ports := []v1.ContainerPort{
		{
			Name:          "sentinel",
			ContainerPort: int32(26379),
			Protocol:      v1.ProtocolTCP,
		},
	}

	return redisContainer(commands, baseImage, version, ports...)
}

func redisMaster(commands, baseImage, version string) v1.Container {

	ports := []v1.ContainerPort{
		{
			Name:          "main",
			ContainerPort: int32(6379),
			Protocol:      v1.ProtocolTCP,
		},
	}

	return redisContainer(commands, baseImage, version, ports...)
}

func redisSlave(commands, baseImage, version string) v1.Container {

	return redisMaster(commands, baseImage, version)
}
