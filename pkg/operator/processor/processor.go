package processor

import (
	"github.com/JoelW-S/redis-operator/pkg/operator/event"
	"github.com/JoelW-S/redis-operator/pkg/operator/spec"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type RedisEventProcessor struct {
	eventCh chan<- event.Event
	stopCh  <-chan struct{}
}

func New(eventCh chan<- event.Event, stopCh <-chan struct{}) cache.ResourceEventHandler {
	return RedisEventProcessor{
		eventCh: eventCh,
		stopCh:  stopCh,
	}
}

func (ep RedisEventProcessor) OnAdd(obj interface{}) {

	redis := obj.(*spec.RedisServer)

	e := event.Event{
		Type:   kwatch.Added,
		Object: redis,
	}

	ep.eventCh <- e
}
func (ep RedisEventProcessor) OnUpdate(oldObj, newObj interface{}) {

	redis := newObj.(*spec.RedisServer)

	e := event.Event{
		Type:   kwatch.Modified,
		Object: redis,
	}

	ep.eventCh <- e

}
func (ep RedisEventProcessor) OnDelete(obj interface{}) {

	redis := obj.(*spec.RedisServer)

	e := event.Event{
		Type:   kwatch.Deleted,
		Object: redis,
	}

	ep.eventCh <- e

}
