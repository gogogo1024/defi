package eventstore

import (
	"database/sql"
	"defi/internal/model"
	"fmt"
)

type BaseEventStore struct {
	Db *sql.DB
}

func (es *BaseEventStore) SaveEvent(event model.Event) error {
	query := `INSERT INTO events (id, aggregate_id, type, data, timestamp) VALUES (?, ?, ?, ?, ?)`
	_, err := es.Db.Exec(query, event.ID, event.AggregateID, event.Type, event.Data, event.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}
	return nil
}

func (es *BaseEventStore) GetEvents(aggregateID string) ([]model.Event, error) {
	query := `SELECT id, aggregate_id, type, data, timestamp FROM events WHERE aggregate_id = ?`
	rows, err := es.Db.Query(query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		if err := rows.Scan(&event.ID, &event.AggregateID, &event.Type, &event.Data, &event.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return events, nil
}
func (es *BaseEventStore) QueryEvents(aggregateID string) ([]model.Event, error) {
	rows, err := es.Db.Query("SELECT * FROM events WHERE aggregate_id = ? ORDER BY timestamp", aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		if err := rows.Scan(&event.ID, &event.AggregateID, &event.Type, &event.Data, &event.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
