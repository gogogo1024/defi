package eventbus

import (
	"defi/internal/model"
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestNatsEventBus(t *testing.T) {
	url := nats.DefaultURL
	eb, err := NewNatsEventBus(url)
	if err != nil {
		t.Fatalf("Failed to create NatsEventBus: %v", err)
	}

	topic := "test_topic"
	event := []byte("test_event")

	// Test PublishEvent
	err = eb.PublishEvent(topic, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Test ConsumerEvent
	received := make(chan model.Event)
	err = eb.ConsumerEvent(topic, func(event model.Event) {
		received <- event
	})
	if err != nil {
		t.Fatalf("Failed to consume event: %v", err)
	}

	select {
	case e := <-received:
		if string(e.Data) != string(event) {
			t.Fatalf("Expected event data %s, got %s", event, e.Data)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for event")
	}
}

func TestKafkaEventBus(t *testing.T) {
	brokers := []string{"localhost:9092"}
	eb, err := NewKafkaEventBus(brokers)
	if err != nil {
		t.Fatalf("Failed to create KafkaEventBus: %v", err)
	}

	topic := "test_topic"
	event := []byte("test_event")

	// Test PublishEvent
	err = eb.PublishEvent(topic, event)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Test ConsumerEvent
	received := make(chan model.Event)
	err = eb.ConsumerEvent(topic, func(event model.Event) {
		received <- event
	})
	if err != nil {
		t.Fatalf("Failed to consume event: %v", err)
	}

	select {
	case e := <-received:
		if string(e.Data) != string(event) {
			t.Fatalf("Expected event data %s, got %s", event, e.Data)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for event")
	}
}
