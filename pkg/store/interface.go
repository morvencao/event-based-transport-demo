package store

import (
	"github.com/morvencao/event-based-transport-demo/pkg/api"
)

type Store interface {
	// Add adds a resource to the store
	Add(resource *api.Resource)
	// Get retrieves a resource from the store
	Get(resourceID string) (*api.Resource, error)
	// Update updates a resource in the store
	Update(resource *api.Resource) error
	// UpSert updates or inserts a resource into the store
	UpSert(resource *api.Resource)
	// UpdateStatus updates the status of a resource in the store
	UpdateStatus(resource *api.Resource) error
	// MarkAsDeleting marks a resource as deleting in the store
	MarkAsDeleting(resourceID string)
	// Delete deletes a resource from the store
	Delete(resourceID string)
	// List lists all resources in the store
	List(namespace string) []*api.Resource
	// ListAll lists all resources in the store
	ListAll() []*api.Resource
}
