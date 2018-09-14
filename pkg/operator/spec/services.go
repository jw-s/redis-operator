package spec

import (
	"fmt"

	cr "github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetMasterPodName(name string) string {
	return fmt.Sprintf(MasterPodName, name)
}

func GetSlaveStatefulSetName(name string) string {
	return fmt.Sprintf(SlaveStatefulSetName, name)
}

func GetSentinelServiceName(name string) string {
	return fmt.Sprintf(SentinelServiceName, name)
}

func GetMasterServiceName(name string) string {
	return fmt.Sprintf(MasterServiceName, name)
}

func MasterService(owner *cr.Redis) *apiv1.Service {

	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMasterServiceName(owner.Name),
			Namespace: owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Protocol:   apiv1.ProtocolTCP,
					Port:       RedisPort,
					TargetPort: intstr.FromInt(RedisPort),
				},
			},
		},
	}

}

func MasterServiceEndpoint(owner *cr.Redis, IPAddress string) *apiv1.Endpoints {

	return &apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMasterServiceName(owner.Name),
			Namespace: owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Subsets: []apiv1.EndpointSubset{
			{
				Addresses: []apiv1.EndpointAddress{
					{
						IP: IPAddress,
					},
				},
				Ports: []apiv1.EndpointPort{
					{
						Protocol: apiv1.ProtocolTCP,
						Port:     RedisPort,
					},
				},
			},
		},
	}

}

func SentinelService(owner *cr.Redis) *apiv1.Service {

	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSentinelServiceName(owner.Name),
			Namespace: owner.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				owner.AsOwner(),
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Protocol:   apiv1.ProtocolTCP,
					Port:       RedisSentinelPort,
					TargetPort: intstr.FromInt(RedisSentinelPort),
				},
			},
			Selector: SentinelLabelSelector(owner),
		},
	}

}
