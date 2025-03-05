package eventstore

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type PostgresEventStore struct {
	*BaseEventStore
}

func NewPostgresEventStore(dataSourceName string) (*PostgresEventStore, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &PostgresEventStore{&BaseEventStore{Db: db}}, nil
}
