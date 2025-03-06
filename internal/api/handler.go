package api

import (
	"defi/internal/eventstore"
	"defi/internal/model"
	"encoding/json"
	"net/http"
)

var es *eventstore.BaseEventStore

func InitEventStore(store *eventstore.BaseEventStore) {
	es = store
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	var event model.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := es.SaveEvent(event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
