package main

import (
	"database/sql"
	"defi/internal/eventstore"
	"defi/internal/model"
	"github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func main() {
	// Initialize the database connection
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := &eventstore.BaseEventStore{Db: db}

	// Initialize the Kafka consumer
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("example_topic", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to start partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	// Consume messages and save events to the database
	for msg := range partitionConsumer.Messages() {
		event := model.Event{
			Data: string(msg.Value),
		}
		if err := store.SaveEvent(event); err != nil {
			log.Printf("Failed to save event to database: %v", err)
		}
	}
}
