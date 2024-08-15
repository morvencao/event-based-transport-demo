# Event-Based Transport Demo

## Setup

1. **Deploy Agent and Transport**
```bash
make setup
```

```bash
export KUBECONFIG=${PWD}/test/.kubeconfig
```

2. **Run the Source**
```bash
go run ./cmd/main.go source --transport-addr localhost:31883
```

## Resource Management

### 1. Create a Resource
```bash
curl -X POST localhost:8080/resources -d @example/resource.json | jq
```

### 2. Get All Resources
```bash
curl localhost:8080/resources | jq
```

### 3. Get Resource by ID
```bash
resourceID=$(curl localhost:8080/resources | jq -r .[].resourceID)
curl localhost:8080/resources/${resourceID} | jq
```

### 4. Verify Resource Creation in Cluster
```bash
kubectl get deploy -n default
```

### 5. Update the Resource
```bash
curl -X PATCH localhost:8080/resources/${resourceID} -d @example/resource-patch.json | jq
kubectl get deploy -n default
```

### 6. Delete the Resource
```bash
curl -X DELETE localhost:8080/resources/${resourceID} | jq
kubectl get deploy -n default
```

### 7. Verify Resource Deletion at Source
```bash
curl localhost:8080/resources/${resourceID} | jq
```

## Resync Resources

### 1. Scale Down Agent
```bash
kubectl -n agent scale deploy/agent --replicas 0
```

### 2. Create a New Resource
```bash
curl -X POST localhost:8080/resources -d @example/resource.json | jq
```

### 3. Verify the resource is not created in cluster
```bash
kubectl get deploy -n default
```

### 4. Scale up agent
```bash
kubectl -n agent scale deploy/agent --replicas 1
```

### 5. Verify the resource is synced
```bash
kubectl get deploy -n default
curl localhost:8080/resources | jq
```
