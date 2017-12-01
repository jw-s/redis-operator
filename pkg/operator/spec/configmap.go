package spec

import (
	"fmt"
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetSentinelConfigMapName(name string) string {
	return fmt.Sprintf(SentinelConfigMapName, name)
}

func DefaultSentinelConfig(redis *v1.Redis) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSentinelConfigMapName(redis.Name),
			Namespace: redis.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				redis.AsOwner(),
			},
		},
		Data: map[string]string{
			ConfigMapConfKeyName: defaultSentinelConfig(redis.Name, redis.Spec.Sentinels.Quorum),
		},
	}
}

func defaultSentinelConfig(redisName string, quorum int32) string {
	return fmt.Sprintf(`
dir %[1]s
sentinel monitor %[2]s %[3]s %[4]d %[5]d
sentinel down-after-milliseconds %[2]s 30000
sentinel parallel-syncs %[2]s 1
sentinel failover-timeout %[2]s 180000`,
		DataMountPath,
		redisName,
		GetMasterServiceName(redisName),
		RedisPort,
		quorum)

}
