# Data Model: `Supports` Feature

## Entities

The primary data entity for this feature is the `ResourceDescriptor`, which is used to identify a resource being queried for cost support.

### ResourceDescriptor

Describes a cloud resource for cost analysis.

| Field | Type | Description | Validation |
| :--- | :--- | :--- | :--- |
| `provider` | `string` | Cloud provider. | Must match a known provider (e.g., "aws", "gcp"). |
| `resource_type` | `string` | The type of the resource. | Must match a known resource type (e.g., "ec2"). |
| `sku` | `string` | Provider-specific SKU or instance size. | Optional. No validation required at this level. |
| `region` | `string` | The deployment region. | Must match a known region (e.g., "us-east-1"). |
| `tags` | `map<string, string>` | Key-value tags for the resource. | Optional. No validation required. |

*Note: Validation rules are based on the clarification session (Strict Validation).*
