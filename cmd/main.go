package main

import (
	"defi/internal/config"
	"defi/internal/db"
	"defi/internal/eventbus"
	"defi/internal/eventstore"
	"defi/internal/model"
	"log"
	"os"
)

func main() {
	// Set the APP_ENV environment variable
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local" // Default to local if APP_ENV is not set
		os.Setenv("APP_ENV", appEnv)
	}

	mqConfigs, dbConfigs, cacheConfigs, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database := db.InitDB(dbConfigs.MySQL, cacheConfigs.Cache)
	defer func() {
		if database.SQL != nil {
			database.SQL.Close()
		}
		if database.Redis != nil {
			database.Redis.Close()
		}
	}()

	store := &eventstore.BaseEventStore{Db: database.SQL}
	mqEventBus := eventbus.InitEventBus(mqConfigs.Kafka)

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
