package eventbus

type EventBus interface {
	PublishEvent(topic string, event []byte) error
	ConsumerEvent(topic string, event []byte) error
}
