package spec

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(RedisSchemeGroupVersion,
		&RedisServer{},
		&RedisServerList{},
	)
	metav1.AddToGroupVersion(s, RedisSchemeGroupVersion)
	return nil
}
