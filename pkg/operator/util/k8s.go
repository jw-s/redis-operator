package util

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/JoelW-S/redis-operator/pkg/operator/spec"
	"k8s.io/client-go/rest"
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

func GetRedisServerList(restcli rest.Interface, ns string) (*spec.RedisServerList, error) {
	b, err := restcli.Get().
		RequestURI(
			ListRedisListURI(ns)).
		DoRaw()

	if err != nil {
		return nil, err
	}

	servers := &spec.RedisServerList{}
	if err := json.Unmarshal(b, servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func ListRedisListURI(namespace string) string {
	return fmt.Sprintf("/apis/%s/namespaces/%s/%s", spec.RedisNamePlural, namespace, spec.RedisNamePlural)
}
