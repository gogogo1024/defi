package eventstore

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLEventStore struct {
	*BaseEventStore
}

func NewMySQLEventStore(dataSourceName string) (*MySQLEventStore, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &MySQLEventStore{&BaseEventStore{Db: db}}, nil
}
