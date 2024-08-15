package source

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/morvencao/event-based-transport-demo/pkg/api"
)

func StatusHashGetter(resource *api.Resource) (string, error) {
	statusBytes, err := json.Marshal(resource.Status)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource status, %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(statusBytes)), nil
}
