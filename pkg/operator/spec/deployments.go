package spec

import (
	"fmt"
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetSentinelDeploymentName(name string) string {
	return fmt.Sprintf(SentinelDeploymentName, name)
}

func MasterSeedPod(owner *v1.Redis) *apiv1.Pod {

	redis := RedisContainer(owner.Spec.BaseImage, owner.Spec.Version)

	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMasterPodName(owner.Name),
			Namespace: owner.Namespace,
			Labels:    SentinelLabelSelector(owner),
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				RedisMaster(redis),
			},
		},
	}
}

func SentinelDeployment(owner *v1.Redis) *appsv1beta1.Deployment {

	redis := RedisContainer(owner.Spec.BaseImage, owner.Spec.Version)

	replicas := owner.Spec.Sentinels.Replicas

	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSentinelDeploymentName(owner.Name),
			Namespace: owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: SentinelLabelSelector(owner),
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: SentinelLabelSelector(owner),
					OwnerReferences: []metav1.OwnerReference{
						owner.AsOwner(),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						RedisSentinel(redis),
					},
					InitContainers: []apiv1.Container{
						{
							Name:            "tiny-tools",
							Image:           "giantswarm/tiny-tools",
							ImagePullPolicy: apiv1.PullAlways,
							Command: []string{
								"cp",
								fmt.Sprintf("%s/%s", ConfMountPath, ConfigMapConfKeyName),
								fmt.Sprintf("%s/%s", DataMountPath, ConfigMapConfKeyName),
							},
							VolumeMounts: GetVolumeMounts(),
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: ConfigVolumeName,
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: string(owner.Spec.Sentinels.ConfigMap),
									},
								},
							},
						},
						{
							Name: DataVolumeName,
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

}

func SentinelLabelSelector(redis *v1.Redis) map[string]string {
	return map[string]string{
		OperatorLabel: redis.Name,
		"app":         "redis",
		"role":        "sentinel",
	}
}

func SlaveLabelSelector(redis *v1.Redis) map[string]string {
	return map[string]string{
		OperatorLabel: redis.Name,
		"app":         "redis",
		"role":        "slave",
	}
}
