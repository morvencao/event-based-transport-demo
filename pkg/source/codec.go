package source

import (
	"encoding/json"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventstypes "github.com/cloudevents/sdk-go/v2/types"
	"github.com/morvencao/event-based-transport-demo/pkg/api"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	workv1 "open-cluster-management.io/api/work/v1"

	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/common"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/payload"
)

type ResourceCodec struct{}

var _ generic.Codec[*api.Resource] = &ResourceCodec{}

func (c *ResourceCodec) EventDataType() types.CloudEventsDataType {
	return payload.ManifestEventDataType
}

func (c *ResourceCodec) Encode(source string, eventType types.CloudEventsType, resource *api.Resource) (*cloudevents.Event, error) {
	if resource.Source != "" {
		source = resource.Source
	}

	if eventType.CloudEventsDataType != payload.ManifestEventDataType {
		return nil, fmt.Errorf("unsupported cloudevents data type %s", eventType.CloudEventsDataType)
	}

	eventBuilder := types.NewEventBuilder(source, eventType).
		WithResourceID(resource.ResourceID).
		WithResourceVersion(resource.ResourceVersion).
		WithClusterName(resource.ClusterName)

	if !resource.GetDeletionTimestamp().IsZero() {
		evt := eventBuilder.WithDeletionTimestamp(resource.GetDeletionTimestamp().Time).NewEvent()
		return &evt, nil
	}

	evt := eventBuilder.NewEvent()

	eventPayload := &payload.Manifest{
		Manifest: *resource.Spec,
		DeleteOption: &workv1.DeleteOption{
			PropagationPolicy: workv1.DeletePropagationPolicyTypeForeground,
		},
		ConfigOption: &payload.ManifestConfigOption{
			FeedbackRules: []workv1.FeedbackRule{
				{
					Type: workv1.JSONPathsType,
					JsonPaths: []workv1.JsonPath{
						{
							Name: "status",
							Path: ".status",
						},
					},
				},
			},
			UpdateStrategy: &workv1.UpdateStrategy{
				Type: workv1.UpdateStrategyTypeServerSideApply,
			},
		},
	}

	if err := evt.SetData(cloudevents.ApplicationJSON, eventPayload); err != nil {
		return nil, fmt.Errorf("failed to encode manifests to cloud event: %v", err)
	}

	return &evt, nil
}

func (c *ResourceCodec) Decode(evt *cloudevents.Event) (*api.Resource, error) {
	eventType, err := types.ParseCloudEventsType(evt.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud event type %s, %v", evt.Type(), err)
	}

	if eventType.CloudEventsDataType != payload.ManifestEventDataType {
		return nil, fmt.Errorf("unsupported cloudevents data type %s", eventType.CloudEventsDataType)
	}

	evtExtensions := evt.Context.GetExtensions()

	resourceID, err := cloudeventstypes.ToString(evtExtensions[types.ExtensionResourceID])
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceid extension: %v", err)
	}

	resourceVersion, err := cloudeventstypes.ToInteger(evtExtensions[types.ExtensionResourceVersion])
	if err != nil {
		return nil, fmt.Errorf("failed to get resourceversion extension: %v", err)
	}

	clusterName, err := cloudeventstypes.ToString(evtExtensions[types.ExtensionClusterName])
	if err != nil {
		return nil, fmt.Errorf("failed to get clustername extension: %v", err)
	}

	sequenceID, err := cloudeventstypes.ToString(evtExtensions[types.ExtensionStatusUpdateSequenceID])
	if err != nil {
		return nil, fmt.Errorf("failed to get sequenceid extension: %v", err)
	}

	manifestStatus := &payload.ManifestStatus{}
	if err := evt.DataAs(manifestStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data %s, %v", string(evt.Data()), err)
	}

	resource := &api.Resource{
		ResourceID:      resourceID,
		ResourceVersion: int64(resourceVersion),
		ClusterName:     clusterName,
		Status: &api.ResourceStatus{
			ReconcileStatus: &api.ReconcileStatus{
				SequenceID: sequenceID,
			},
		},
	}

	// set deleted condition if manifest is deleted from agent
	if meta.IsStatusConditionTrue(manifestStatus.Conditions, common.ManifestsDeleted) {
		resource.Status.ReconcileStatus.Conditions = append(resource.Status.ReconcileStatus.Conditions, metav1.Condition{
			Type:   common.ManifestsDeleted,
			Status: metav1.ConditionTrue})
	}

	if manifestStatus.Status != nil {
		resource.Status.ReconcileStatus.Conditions = append(resource.Status.ReconcileStatus.Conditions, manifestStatus.Status.Conditions...)
		for _, value := range manifestStatus.Status.StatusFeedbacks.Values {
			if value.Name == "status" {
				contentStatus := make(map[string]interface{})
				if err := json.Unmarshal([]byte(*value.Value.JsonRaw), &contentStatus); err != nil {
					return nil, fmt.Errorf("failed to convert status feedback value to content status: %v", err)
				}
				resource.Status.ContentStatus = contentStatus
			}
		}
	}

	return resource, nil
}
