package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubetypes "k8s.io/apimachinery/pkg/types"

	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
)

type ResourceStatus struct {
	ReconcileStatus *ReconcileStatus       `json:"reconcileStatus"`
	ContentStatus   map[string]interface{} `json:"contentStatus"`
}

type ReconcileStatus struct {
	SequenceID string             `json:"sequenceID"`
	Conditions []metav1.Condition `json:"conditions"`
}

type Resource struct {
	Source            string                     `json:"source"`
	ClusterName       string                     `json:"clusterName"`
	ResourceID        string                     `json:"resourceID"`
	ResourceVersion   int64                      `json:"resourceVersion"`
	DeletionTimestamp time.Time                  `json:"deletionTimestamp"`
	Spec              *unstructured.Unstructured `json:"spec"`
	Status            *ResourceStatus            `json:"status"`
}

var _ generic.ResourceObject = &Resource{}

func NewResource(name, clusterName string, resourceVersion int64, objectJSON string) (*Resource, error) {
	var object map[string]interface{}
	if err := json.Unmarshal([]byte(objectJSON), &object); err != nil {
		return nil, err
	}

	return &Resource{
		ResourceID:      ResourceID(clusterName, name),
		ResourceVersion: resourceVersion,
		Spec: &unstructured.Unstructured{
			Object: object,
		},
	}, nil
}

func (r *Resource) GetUID() kubetypes.UID {
	return kubetypes.UID(r.ResourceID)
}

func (r *Resource) GetResourceVersion() string {
	return fmt.Sprintf("%d", r.ResourceVersion)
}

func (r *Resource) GetDeletionTimestamp() *metav1.Time {
	return &metav1.Time{Time: r.DeletionTimestamp}
}

func ResourceID(clusterName, name string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("resource-%s-%s", clusterName, name))).String()
}
