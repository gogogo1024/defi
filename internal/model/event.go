package model

type State string

const (
	StateInitial State = "initial"
	StateActive  State = "active"
	StateClosed  State = "closed"
)

type Event struct {
	ID          string
	AggregateID string
	Type        string
	Data        string
	Timestamp   int64
}
