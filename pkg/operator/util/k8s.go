package util

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
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
