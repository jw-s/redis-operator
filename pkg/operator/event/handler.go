package event

import "github.com/sirupsen/logrus"

type Handler interface {
	OnAdd(add Event)
	OnUpdate(update Event)
	OnDelete(delete Event)
}

type eventHandler struct {
	logger *logrus.Entry
}

func NewHandler() Handler {

	return &eventHandler{
		logger: logrus.WithField("pkg", "event"),
	}
}

func (eh *eventHandler) OnAdd(add Event) {

	eh.logger.Info(add)
}

func (eh *eventHandler) OnUpdate(update Event) {

	eh.logger.Info(update)
}

func (eh *eventHandler) OnDelete(delete Event) {

	eh.logger.Info(delete)

}
