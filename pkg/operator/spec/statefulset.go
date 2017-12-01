package spec

import (
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SlaveStatefulSet(owner *v1.Redis) *appsv1beta1.StatefulSet {

	redis := RedisContainer(owner.Spec.BaseImage, owner.Spec.Version)

	replicas := owner.Spec.Slaves.Replicas

	return &appsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSlaveStatefulSetName(owner.Name),
			Namespace: owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Spec: appsv1beta1.StatefulSetSpec{
			ServiceName: GetSlaveStatefulSetName(owner.Name),
			Replicas:    &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      GetSlaveStatefulSetName(owner.Name),
					Namespace: owner.Namespace,
					Labels:    SlaveLabelSelector(owner),
					OwnerReferences: []metav1.OwnerReference{
						owner.AsOwner(),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						RedisSlave(redis, owner),
					},
				},
			},
			VolumeClaimTemplates: []apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      SlavePersistentVolumeClaimName,
						Namespace: owner.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							owner.AsOwner(),
						},
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: []apiv1.PersistentVolumeAccessMode{
							apiv1.ReadWriteOnce,
						},
						Resources: owner.Spec.Pod.Resources,
					},
				},
			},
		},
	}
}
