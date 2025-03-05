package eventbus

import (
	"github.com/nats-io/nats.go"
)

type NatsEventBus struct {
	conn *nats.Conn
}

func NewNatsEventBus(url string) (*NatsEventBus, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsEventBus{conn: conn}, nil
}

func (eb *NatsEventBus) PublishEvent(subject string, event []byte) error {
	return eb.conn.Publish(subject, event)
}
func (eb *NatsEventBus) Subscribe(subject string, handler func(event []byte)) (*nats.Subscription, error) {
	return eb.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
}
