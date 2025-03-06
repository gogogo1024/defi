package eventbus

import (
	"defi/internal/config"
	"errors"
)

func NewEventBus(cfg config.MQConfig) (EventBus, error) {
	switch cfg.Type {
	case "kafka":
		return NewKafkaEventBus(cfg.Brokers)
	case "nats":
		return NewNatsEventBus(cfg.URL)
	default:
		return nil, errors.New("unsupported message queue type")
	}
}
