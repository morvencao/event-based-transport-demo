package source

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/morvencao/event-based-transport-demo/pkg/api"
	"github.com/morvencao/event-based-transport-demo/pkg/store"
)

type APIServer struct {
	sourceID        string
	server          *http.Server
	store           store.Store
	eventController EventController
}

func NewAPIServer(addr, sourceID string, store store.Store, eventController *EventController) *APIServer {
	s := &APIServer{
		sourceID:        sourceID,
		store:           store,
		eventController: *eventController,
	}

	router := gin.Default()
	router.GET("/resources", s.getResources)
	router.GET("/resources/:id", s.getResourceByID)
	router.POST("/resources", s.postResource)
	router.PATCH("/resources/:id", s.updateResource)
	router.DELETE("/resources/:id", s.deleteResource)

	s.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return s
}

func (s *APIServer) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.server.Shutdown(ctx)
	}()

	return s.server.ListenAndServe()
}

func (s *APIServer) getResources(c *gin.Context) {
	resources := s.store.ListAll()
	c.JSON(http.StatusOK, resources)
}

func (s *APIServer) getResourceByID(c *gin.Context) {
	id := c.Param("id")
	resource, err := s.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resource)
}

func (s *APIServer) postResource(c *gin.Context) {
	var resource *api.Resource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// server generates a resource ID with UUID
	resource.ResourceID = uuid.New().String()
	// server sets the source ID
	resource.Source = s.sourceID
	// server sets the resource version to 1
	resource.ResourceVersion = 1
	// persist the resource
	s.store.Add(resource)

	// enqueue a create event
	event := Event{
		EventType: CreateEvent,
		ID:        resource.ResourceID,
	}
	s.eventController.EnqueueEvent(event)

	c.JSON(http.StatusCreated, resource)
}

func (s *APIServer) updateResource(c *gin.Context) {
	id := c.Param("id")
	var resource *api.Resource
	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	found, err := s.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if reflect.DeepEqual(resource.Spec, found.Spec) {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "no change in resource spec"})
		return
	}

	// update the resource spec
	found.Spec = resource.Spec
	// increment the resource version
	found.ResourceVersion = found.ResourceVersion + 1
	// persist the resource
	if err := s.store.Update(found); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// enqueue an update event
	event := Event{
		EventType: UpdateEvent,
		ID:        id,
	}
	s.eventController.EnqueueEvent(event)

	c.JSON(http.StatusOK, resource)
}

func (s *APIServer) deleteResource(c *gin.Context) {
	id := c.Param("id")
	resource, err := s.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// mark the resource as deleting
	s.store.MarkAsDeleting(id)

	// enqueue a delete event
	event := Event{
		EventType: DeleteEvent,
		ID:        resource.ResourceID,
	}
	s.eventController.EnqueueEvent(event)

	c.JSON(http.StatusNoContent, nil)
}
