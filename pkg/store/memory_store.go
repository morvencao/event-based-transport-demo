package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/morvencao/event-based-transport-demo/pkg/api"
)

type MemoryStore struct {
	sync.RWMutex

	resources map[string]*api.Resource
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		resources: make(map[string]*api.Resource),
	}
}

func (s *MemoryStore) Add(resource *api.Resource) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.resources[resource.ResourceID]
	if !ok {
		s.resources[resource.ResourceID] = resource
	}
}

func (s *MemoryStore) Update(resource *api.Resource) error {
	s.Lock()
	defer s.Unlock()

	found, ok := s.resources[resource.ResourceID]
	if !ok {
		return fmt.Errorf("the resource %s does not exist", resource.ResourceID)
	}

	if !found.DeletionTimestamp.IsZero() {
		return fmt.Errorf("the resource %s is being deleted", resource.ResourceID)
	}

	s.resources[resource.ResourceID] = resource
	return nil
}

func (s *MemoryStore) UpSert(resource *api.Resource) {
	s.Lock()
	defer s.Unlock()

	s.resources[resource.ResourceID] = resource
}

func (s *MemoryStore) UpdateStatus(resource *api.Resource) error {
	s.Lock()
	defer s.Unlock()

	last, ok := s.resources[resource.ResourceID]
	if !ok {
		return fmt.Errorf("the resource %s does not exist", resource.ResourceID)
	}

	last.Status = resource.Status
	s.resources[resource.ResourceID] = last
	return nil
}

func (s *MemoryStore) MarkAsDeleting(resourceID string) {
	s.Lock()
	defer s.Unlock()

	resource, ok := s.resources[resourceID]
	if !ok {
		return
	}

	resource.DeletionTimestamp = time.Now()
	s.resources[resourceID] = resource
}

func (s *MemoryStore) Delete(resourceID string) {
	s.Lock()
	defer s.Unlock()

	delete(s.resources, resourceID)
}

func (s *MemoryStore) Get(resourceID string) (*api.Resource, error) {
	s.RLock()
	defer s.RUnlock()

	resource, ok := s.resources[resourceID]
	if !ok {
		return nil, fmt.Errorf("failed to find resource %s", resourceID)
	}

	return resource, nil
}

func (s *MemoryStore) List(namespace string) []*api.Resource {
	s.RLock()
	defer s.RUnlock()

	resources := []*api.Resource{}
	for _, res := range s.resources {
		if res.ClusterName != namespace {
			continue
		}

		resources = append(resources, res)
	}
	return resources
}

func (s *MemoryStore) ListAll() []*api.Resource {
	s.RLock()
	defer s.RUnlock()

	resources := []*api.Resource{}
	for _, res := range s.resources {
		resources = append(resources, res)
	}
	return resources
}
