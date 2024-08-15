package source

import (
	"context"

	"github.com/morvencao/event-based-transport-demo/pkg/api"
	"github.com/morvencao/event-based-transport-demo/pkg/store"
	"k8s.io/apimachinery/pkg/api/meta"

	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/common"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/payload"
)

var (
	createRequest = types.CloudEventsType{
		CloudEventsDataType: payload.ManifestEventDataType,
		SubResource:         types.SubResourceSpec,
		Action:              "create_request",
	}
	updateRequest = types.CloudEventsType{
		CloudEventsDataType: payload.ManifestEventDataType,
		SubResource:         types.SubResourceSpec,
		Action:              "update_request",
	}
	deleteRequest = types.CloudEventsType{
		CloudEventsDataType: payload.ManifestEventDataType,
		SubResource:         types.SubResourceSpec,
		Action:              "delete_request",
	}
)

type ResourceSourceClient struct {
	client generic.CloudEventsClient[*api.Resource]
	store  store.Store
}

func StartResourceSourceClient(
	ctx context.Context,
	options *options.CloudEventsSourceOptions,
	store store.Store,
) (*ResourceSourceClient, error) {
	client, err := generic.NewCloudEventSourceClient[*api.Resource](
		ctx,
		options,
		&ResourceLister{store: store},
		StatusHashGetter,
		&ResourceCodec{},
	)
	if err != nil {
		return nil, err
	}

	client.Subscribe(ctx, func(action types.ResourceAction, resource *api.Resource) error {
		if meta.IsStatusConditionTrue(resource.Status.ReconcileStatus.Conditions, common.ManifestsDeleted) {
			// Delete the resource if agent reports it's deleted
			store.Delete(resource.ResourceID)
			return nil
		}
		return store.UpdateStatus(resource)
	})

	return &ResourceSourceClient{client: client, store: store}, nil
}

func (c *ResourceSourceClient) OnCreate(ctx context.Context, id string) error {
	resource, err := c.store.Get(id)
	if err != nil {
		return err
	}

	return c.client.Publish(ctx, createRequest, resource)
}

func (c *ResourceSourceClient) OnUpdate(ctx context.Context, id string) error {
	resource, err := c.store.Get(id)
	if err != nil {
		return err
	}

	return c.client.Publish(ctx, updateRequest, resource)
}

func (c *ResourceSourceClient) OnDelete(ctx context.Context, id string) error {
	resource, err := c.store.Get(id)
	if err != nil {
		return err
	}

	return c.client.Publish(ctx, deleteRequest, resource)
}
