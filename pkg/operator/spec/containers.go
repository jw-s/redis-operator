package spec

import (
	"fmt"
	"strconv"

	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	apiv1 "k8s.io/api/core/v1"
)

func GetRedisMasterName(redis *v1.Redis) string {
	return redis.Name
}

func GetVolumeMounts() []apiv1.VolumeMount {
	return []apiv1.VolumeMount{
		{
			Name:      ConfigVolumeName,
			MountPath: ConfMountPath,
		},
		{
			Name:      DataVolumeName,
			MountPath: DataMountPath,
		},
	}
}
func RedisContainer(baseImage, version string) apiv1.Container {

	c := apiv1.Container{
		Name:  "redis",
		Image: baseImage + ":" + version,
		Args: []string{
			"--bind", "0.0.0.0",
			"--slave-read-only", "no",
		},
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

	container.VolumeMounts = GetVolumeMounts()

	container.Args = append([]string{
		fmt.Sprintf("%s/%s", DataMountPath, ConfigMapConfKeyName),
		"--sentinel",
	},
		container.Args...,
	)

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

	redis.SecurityContext = &apiv1.SecurityContext{
		RunAsUser: spec.Spec.GetRedisRunAsUser(),
	}

	redis.VolumeMounts = []apiv1.VolumeMount{
		{
			Name:      SlavePersistentVolumeClaimName,
			MountPath: DataMountPath,
		},
	}

	redis.Args = append(redis.Args,
		"--slaveof", GetMasterServiceName(spec.Name), strconv.Itoa(RedisPort),
	)

	return redis
}
