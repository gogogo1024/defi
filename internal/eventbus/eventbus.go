package eventbus

import (
	"defi/internal/config"
	"defi/internal/model"
)

type EventBus interface {
	PublishEvent(topic string, event []byte) error
	ConsumerEvent(topic string, handler func(event model.Event)) error
}

func InitEventBus(cfg config.MQConfig) EventBus {
	mqEventBus, err := NewEventBus(cfg)
	if err != nil {
		log.Fatalf("Failed to create EventBus: %v", err)
	}
	return mqEventBus
}
