# gRPC Contract: `Supports` Method

This document describes the gRPC contract for the `Supports` method of the `CostSourceService`.

## Service: `CostSourceService`

The core gRPC service that all plugins must implement.

### RPC: `Supports`

Checks if the cost source plugin supports pricing for a given resource.

```protobuf
rpc Supports(SupportsRequest) returns (SupportsResponse);
```

---

### Message: `SupportsRequest`

The request payload for the `Supports` RPC.

| Field | Type | Description |
| :--- | :--- | :--- |
| `resource` | `ResourceDescriptor` | The resource to check for support. |

**Example (JSON representation):**
```json
{
  "resource": {
    "provider": "aws",
    "resource_type": "ec2",
    "sku": "t3.micro",
    "region": "us-east-1"
  }
}
```
---

### Message: `SupportsResponse`

The response payload for the `Supports` RPC.

| Field | Type | Description |
| :--- | :--- | :--- |
| `supported` | `bool` | `true` if the resource is supported, `false` otherwise. |
| `reason` | `string` | An optional message explaining why the resource is not supported. Required if `supported` is `false`. |

**Example Success:**
```json
{
  "supported": true
}
```

**Example Failure:**
```json
{
  "supported": false,
  "reason": "Plugin does not support provider 'gcp'."
}
```
---

### Message: `ResourceDescriptor`

See `data-model.md` for a full description of the `ResourceDescriptor` entity.
