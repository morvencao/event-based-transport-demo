package source

import (
	"github.com/morvencao/event-based-transport-demo/pkg/api"
	"github.com/morvencao/event-based-transport-demo/pkg/store"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
)

type ResourceLister struct {
	store store.Store
}

var _ generic.Lister[*api.Resource] = &ResourceLister{}

func (l *ResourceLister) List(listOpts types.ListOptions) ([]*api.Resource, error) {
	return l.store.List(listOpts.ClusterName), nil
}
