package projection

import (
	"defi/internal/model"
	"sync"
)

type Projection struct {
	State map[string]interface{}
	mu    sync.Mutex
}

func NewProjection() *Projection {
	return &Projection{
		State: make(map[string]interface{}),
	}
}

func (p *Projection) HandleEvents(events []model.Event) {
	var wg sync.WaitGroup
	eventCh := make(chan model.Event, len(events))

	for _, event := range events {
		eventCh <- event
	}
	close(eventCh)

	for i := 0; i < 10; i++ { // Number of concurrent workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := range eventCh {
				p.handleEvent(event)
			}
		}()
	}

	wg.Wait()
}

func (p *Projection) handleEvent(event model.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Implement event handling logic for projection
	switch event.Type {
	case "EventType1":
		// Update state based on EventType1
	case "EventType2":
		// Update state based on EventType2
		// Add more cases as needed
	}
}

func (p *Projection) GetState() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.State
}
