package eventbus

import (
	"defi/internal/model"
	"github.com/Shopify/sarama"
)

type KafkaEventBus struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewKafkaEventBus(brokers []string) (EventBus, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaEventBus{
		producer: producer,
		consumer: consumer,
	}, nil
}

func (eb *KafkaEventBus) PublishEvent(topic string, event []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(event),
	}
	_, _, err := eb.producer.SendMessage(msg)
	return err
}

func (eb *KafkaEventBus) ConsumerEvent(topic string, handler func(event model.Event)) error {
	partitionList, err := eb.consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitionList {
		pc, err := eb.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				event := model.Event{
					Data: string(msg.Value),
				}
				handler(event)
			}
		}(pc)
	}

	return nil
}
