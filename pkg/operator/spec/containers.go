package spec

import (
	"fmt"
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	apiv1 "k8s.io/api/core/v1"
	"strconv"
)

func GetRedisMasterName(redis *v1.Redis) string {
	return redis.Name
}

func GetSentinelVolumeMounts() []apiv1.VolumeMount {
	return []apiv1.VolumeMount{
		{
			Name:      "config",
			MountPath: ConfMountPath,
		},
		{
			Name:      "data",
			MountPath: DataMountPath,
		},
	}
}
func RedisContainer(baseImage, version string) apiv1.Container {

	c := apiv1.Container{
		Name:  "redis",
		Image: baseImage + ":" + version,
	}

	return c
}

func RedisSentinel(container apiv1.Container) apiv1.Container {

	container.Name = "redis-sentinel"

	container.Ports = []apiv1.ContainerPort{
		{
			Name:          "sentinel",
			ContainerPort: int32(RedisSentinelPort),
			Protocol:      apiv1.ProtocolTCP,
		},
	}

	container.VolumeMounts = GetSentinelVolumeMounts()

	container.Args = []string{
		fmt.Sprintf("%s/%s", DataMountPath, ConfigMapConfKeyName),
		"--sentinel",
	}

	return container
}

func RedisMaster(redis apiv1.Container) apiv1.Container {

	redis.Name = "redis-master"

	redis.Ports = []apiv1.ContainerPort{
		{
			Name:          "main",
			ContainerPort: int32(RedisPort),
			Protocol:      apiv1.ProtocolTCP,
		},
	}

	return redis
}

func RedisSlave(redis apiv1.Container, spec *v1.Redis) apiv1.Container {

	redis.Name = "redis-slave"

	redis.Ports = []apiv1.ContainerPort{
		{
			Name:          "main",
			ContainerPort: int32(RedisPort),
			Protocol:      apiv1.ProtocolTCP,
		},
	}

	redis.Args = []string{
		"--slaveof", GetMasterServiceName(spec.Name), strconv.Itoa(RedisPort),
	}

	return redis
}
