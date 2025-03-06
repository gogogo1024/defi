package eventbus

import (
	"defi/internal/model"
	"github.com/nats-io/nats.go"
)

type NatsEventBus struct {
	conn *nats.Conn
}

func NewNatsEventBus(url string) (EventBus, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsEventBus{conn: conn}, nil
}

func (eb *NatsEventBus) PublishEvent(topic string, event []byte) error {
	return eb.conn.Publish(topic, event)
}

func (eb *NatsEventBus) ConsumerEvent(topic string, handler func(event model.Event)) error {
	_, err := eb.conn.Subscribe(topic, func(msg *nats.Msg) {
		event := model.Event{
			Data: string(msg.Data),
		}
		handler(event)
	})
	return err
}
