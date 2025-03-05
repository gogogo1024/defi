package model

type Event struct {
	ID          string
	AggregateID string
	Type        string
	Data        string
	Timestamp   int64
}
