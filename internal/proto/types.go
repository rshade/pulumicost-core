package proto

type FieldMappingStatus string

const (
	StatusSupported   FieldMappingStatus = "SUPPORTED"
	StatusUnsupported FieldMappingStatus = "UNSUPPORTED"
	StatusConditional FieldMappingStatus = "CONDITIONAL"
	StatusDynamic     FieldMappingStatus = "DYNAMIC"
)

// PluginMetadata contains information about a plugin's version and capabilities.
type PluginMetadata struct {
	Name               string            `json:"name"                         yaml:"name"`
	Version            string            `json:"version"                      yaml:"version"`
	SpecVersion        string            `json:"specVersion"                  yaml:"specVersion"`
	SupportedProviders []string          `json:"supportedProviders,omitempty" yaml:"supportedProviders,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"           yaml:"metadata,omitempty"`
}

// FieldMapping describes how a Pulumi resource property maps to pricing inputs.
type FieldMapping struct {
	FieldName    string             `json:"fieldName"    yaml:"fieldName"`
	Status       FieldMappingStatus `json:"status"       yaml:"status"`
	Condition    string             `json:"condition"    yaml:"condition"`
	ExpectedType string             `json:"expectedType" yaml:"expectedType"`
}
