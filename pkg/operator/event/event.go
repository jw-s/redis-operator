package event

import (
	"github.com/JoelW-S/redis-operator/pkg/operator/spec"
	kwatch "k8s.io/apimachinery/pkg/watch"
)

type Event struct {
	Type   kwatch.EventType
	Object *spec.RedisServer
}
