package spec

import (
	"fmt"

	v1 "github.com/jw-s/redis-operator/pkg/apis/redis/v1"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetHAProxyImage(redis *v1.Redis) string {
	return fmt.Sprintf("%s:%s", redis.Spec.BaseImage, redis.Spec.Version)
}

func GetHAProxyName(name string) string {
	return fmt.Sprintf(HAProxyName, name)
}

func HaproxyLabelSelector(redis *v1.Redis) map[string]string {
	return map[string]string{
		OperatorLabel: redis.Name,
		"app":         "redis",
		"role":        "haproxy",
	}
}

func GetHAProxyDeployment(redis *v1.Redis) *appsv1beta1.Deployment {
	name := GetHAProxyName(redis.Name)
	namespace := redis.Namespace

	// make CM name as haproxy deployment name, no override allowed via Spec
	configMapName := name

	haproxyImage := GetHAProxyImage(redis)

	replicas := int32(1)
	if redis.Spec.HAProxyReplicas != nil {
		replicas = *redis.Spec.HAProxyReplicas
	}

	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: HaproxyLabelSelector(redis),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: HaproxyLabelSelector(redis),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "haproxy",
							Image:           haproxyImage,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									Name:          "haproxy",
									ContainerPort: 26379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "haproxy-config",
									MountPath: "/etc/haproxy/haproxy.cfg",
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 30,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: 30,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "haproxy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateHAProxyService(redis *v1.Redis, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Service {
	name := GetHAProxyName(redis.Name)
	namespace := redis.Namespace

	labels = HaproxyLabelSelector(redis)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "http",
				"prometheus.io/path":   "/haproxy_stats",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Port:     9000,
					Protocol: corev1.ProtocolTCP,
					Name:     "exporter",
				},
			},
			Selector: labels,
		},
	}
}

func generateHAProxyConfigMap(redis *v1.Redis, labels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.ConfigMap {
	name := GetHAProxyName(redis.Name)
	namespace := redis.Namespace

	labels = HaproxyLabelSelector(redis)

	sentinelConfigFileContent := `
defaults
  mode tcp
  timeout connect 3s
  timeout server 6s
  timeout client 6s
listen stats
  mode http
  bind :9000
  stats enable
  stats hide-version
  stats realm Haproxy\ Statistics
  stats uri /haproxy_stats
frontend ft_redis
  mode tcp
  bind *:80
  default_backend bk_redis
backend bk_redis
  mode tcp
  option tcp-check
  tcp-check send PING\r\n
  tcp-check expect string +PONG
  tcp-check send info\ replication\r\n
  tcp-check expect string role:master
  tcp-check send QUIT\r\n
  tcp-check expect string +OK
# autogenerate at enpoints watch
#  server redis_backend_01 redis01:6379 maxconn 1024 check inter 1s
#  server redis_backend_02 redis02:6379 maxconn 1024 check inter 1s
#  server redis_backend_03 redis03:6379 maxconn 1024 check inter 1s
`

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: map[string]string{
			"haproxy.cfg": sentinelConfigFileContent,
		},
	}
}
