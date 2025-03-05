package eventbus

import (
	"defi/internal/model"
	"github.com/Shopify/sarama"
)

type KafkaEventBus struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewKafkaEventBus(brokers []string) (*KafkaEventBus, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaEventBus{producer: producer, consumer: consumer}, nil
}

func (eb *KafkaEventBus) PublishEvent(topic string, event []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(event),
	}
	_, _, err := eb.producer.SendMessage(msg)
	return err
}
func (eb *KafkaEventBus) Subscribe(topic string, handler func(event model.Event)) error {
	partitionConsumer, err := eb.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for msg := range partitionConsumer.Messages() {
			event := model.Event{
				Data: string(msg.Value),
			}
			handler(event)
		}
	}()

	return nil
}
