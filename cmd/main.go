package main

import (
	"defi/internal/config"
	"defi/internal/db"
	"defi/internal/eventbus"
	"defi/internal/eventstore"
	"defi/internal/model"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	database := db.InitDB()
	defer database.Close()

	store := &eventstore.BaseEventStore{Db: database}
	mqEventBus := eventbus.InitEventBus(cfg)

	publishEvent(mqEventBus)
	consumeEvent(mqEventBus, store)
}

func publishEvent(mqEventBus eventbus.EventBus) {
	err := mqEventBus.PublishEvent("example_topic", []byte("example_event"))
	if err != nil {
		log.Fatalf("Failed to publish event: %v", err)
	}
}

func consumeEvent(mqEventBus eventbus.EventBus, store *eventstore.BaseEventStore) {
	err := mqEventBus.ConsumerEvent("example_topic", func(event model.Event) {
		log.Printf("Received event: %s", event.Data)
		if err := store.SaveEvent(event); err != nil {
			log.Printf("Failed to save event to database: %v", err)
		} else {
			log.Printf("Successfully saved event: %v", event)
		}
	})
	if err != nil {
		log.Fatalf("Failed to consume event: %v", err)
	}
}
