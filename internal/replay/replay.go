package replay

import (
	"defi/internal/eventstore"
	"defi/internal/projection"
)

func ReplayEvents(es *eventstore.BaseEventStore, p *projection.Projection, aggregateID string) error {
	events, err := es.GetEvents(aggregateID)
	if err != nil {
		return err
	}
	p.HandleEvents(events)

	return nil
}
