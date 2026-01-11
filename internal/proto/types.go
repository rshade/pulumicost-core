package proto

// PluginMetadata contains information about a plugin's version and capabilities.
type PluginMetadata struct {
	Name               string
	Version            string
	SpecVersion        string
	SupportedProviders []string
	Metadata           map[string]string
}

// FieldMapping describes how a Pulumi resource property maps to pricing inputs.
type FieldMapping struct {
	FieldName    string
	Status       string // SUPPORTED, UNSUPPORTED, CONDITIONAL, DYNAMIC
	Condition    string
	ExpectedType string
}
