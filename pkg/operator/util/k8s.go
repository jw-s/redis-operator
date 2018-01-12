package util

import (
	rediserrors "github.com/jw-s/redis-operator/pkg/errors"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"time"
)

func InClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			panic(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	return rest.InClusterConfig()
}

func WaitForResourceToBeEstablished(retries int, resourceFunc func() (bool, error)) error {

	return wait.ExponentialBackoff(
		wait.Backoff{
			Duration: time.Second,
			Factor:   2.0,
			Steps:    retries},
		resourceFunc)

}

func IsPodReady(pod *apiv1.Pod) bool {
	conditions := pod.Status.Conditions
	for _, condition := range conditions {
		if condition.Type == apiv1.PodReady {
			return true
		}
	}

	return false
}

func CanServeService(endpoints *apiv1.Endpoints) bool {

	for _, subset := range endpoints.Subsets {

		if len(subset.Addresses) >= 1 {
			return true
		}
	}
	return false
}

func InPodPhase(phase apiv1.PodPhase, pod *apiv1.Pod) bool {
	return phase == pod.Status.Phase
}

func ResourceAlreadyExistError(err error) bool {
	return errors.IsAlreadyExists(err)
}

func ResourceNotFoundError(err error) bool {
	return errors.IsNotFound(err)
}

func CreateKubeResource(kubeClient kubernetes.Interface, namespace string, obj runtime.Object) (err error) {
	switch t := obj.(type) {

	case *apiv1.Pod:
		_, err = kubeClient.CoreV1().Pods(namespace).Create(t)

	case *appsv1beta1.Deployment:
		_, err = kubeClient.AppsV1beta1().Deployments(namespace).Create(t)

	case *appsv1beta1.StatefulSet:
		_, err = kubeClient.AppsV1beta1().StatefulSets(namespace).Create(t)

	case *apiv1.Endpoints:
		_, err = kubeClient.CoreV1().Endpoints(namespace).Create(t)

	case *apiv1.Service:
		_, err = kubeClient.CoreV1().Services(namespace).Create(t)

	case *apiv1.ConfigMap:
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(t)

	default:
		err = rediserrors.UnsupportedKubeResource
	}

	return
}
