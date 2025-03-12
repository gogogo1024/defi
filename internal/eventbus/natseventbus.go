package eventbus

import (
	"defi/internal/model"
	"fmt"
	"github.com/nats-io/nats.go"
)

type NatsEventBus struct {
	conn *nats.Conn
}

func NewNatsEventBus(url string) (EventBus, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("NATS connection error: %w", err)
	}
	return &NatsEventBus{conn: conn}, nil
}

func (eb *NatsEventBus) PublishEvent(topic string, event []byte) error {
	err := eb.conn.Publish(topic, event)
	if err != nil {
		logError("NATS", topic, err)
		return fmt.Errorf("NATS publish error: %w", err)
	}
	logSuccess("NATS", topic, event)
	return nil
}

func (eb *NatsEventBus) ConsumerEvent(topic string, handler func(event model.Event)) error {
	_, err := eb.conn.Subscribe(topic, func(msg *nats.Msg) {
		event := model.Event{
			Data: string(msg.Data),
		}
		handler(event)
	})
	if err != nil {
		logError("NATS", topic, err)
		return fmt.Errorf("NATS subscription error: %w", err)
	}
	return nil
}
