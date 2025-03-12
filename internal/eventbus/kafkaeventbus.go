package eventbus

import (
	"defi/internal/model"
	"fmt"
	"github.com/Shopify/sarama"
	stdlog "log"
	"time"
)

type KafkaEventBus struct {
	producer sarama.AsyncProducer
	consumer sarama.Consumer
}

func NewKafkaEventBus(brokers []string) (EventBus, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Idempotent = true
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Flush.Messages = 100
	config.Producer.Flush.MaxMessages = 1000

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		stdlog.Fatalf("Failed to start Sarama producer: %v", err)
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, config)
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
	eb.producer.Input() <- msg

	select {
	case err := <-eb.producer.Errors():
		logError("Kafka", topic, err)
		return fmt.Errorf("kafka publish error: %w", err)
	case success := <-eb.producer.Successes():
		logSuccess("Kafka", topic, success)
		return nil
	}
}

func (eb *KafkaEventBus) ConsumerEvent(topic string, handler func(event model.Event)) error {
	partitionList, err := eb.consumer.Partitions(topic)
	if err != nil {
		logError("Kafka", topic, err)
		return fmt.Errorf("kafka partition error: %w", err)
	}

	for _, partition := range partitionList {
		pc, err := eb.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			logError("Kafka", topic, err)
			return fmt.Errorf("kafka consume partition error: %w", err)
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
