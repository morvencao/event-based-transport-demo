package source

import (
	"context"
	"fmt"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

type EventType string

const (
	CreateEvent EventType = "create_event"
	UpdateEvent EventType = "update_event"
	DeleteEvent EventType = "delete_event"
)

type Event struct {
	EventType EventType
	ID        string
}

type EventHandler func(ctx context.Context, id string) error

type EventController struct {
	eventsQueue workqueue.RateLimitingInterface
	handlers    map[EventType][]EventHandler
}

func NewEventController() *EventController {
	return &EventController{
		eventsQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "events"),
		handlers:    make(map[EventType][]EventHandler),
	}
}

func (ec *EventController) AddEventHandler(eventType EventType, handler EventHandler) {
	ec.handlers[eventType] = append(ec.handlers[eventType], handler)
}

func (ec *EventController) EnqueueEvent(event Event) {
	ec.eventsQueue.Add(event)
}

func (ec *EventController) Run(ctx context.Context) {
	log.Print("Starting event controller")
	defer ec.eventsQueue.ShutDown()

	// start a goroutine to handle the event from the event queue
	go wait.Until(ec.runWorker, time.Second, ctx.Done())

	// wait until we're told to stop
	<-ctx.Done()
	log.Print("Shutting down event controller")
}

func (ec *EventController) runWorker() {
	// hot loop until we're told to stop.
	for ec.processNextEvent() {
	}
}

func (ec *EventController) processNextEvent() bool {
	key, quit := ec.eventsQueue.Get()
	if quit {
		// the current queue is shutdown and becomes empty, quit this process
		return false
	}
	defer ec.eventsQueue.Done(key)

	if err := ec.handleEvent(key.(Event)); err != nil {
		log.Printf("Failed to handle the event %v, %v ", key, err)

		// requeue the item to work on later
		ec.eventsQueue.AddRateLimited(key)
		return true
	}

	// handle the event successfully, forget it
	ec.eventsQueue.Forget(key)
	return true
}

func (ec *EventController) handleEvent(event Event) error {
	ctx := context.Background()
	handlers, found := ec.handlers[event.EventType]
	if !found {
		log.Printf("No handler functions found for '%s'\n", event.EventType)
		return nil
	}

	for _, handler := range handlers {
		err := handler(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("error handing event %s, %s: %s", event.EventType, event.ID, err)
		}
	}

	return nil
}
