package reforge

import prefabProto "github.com/ReforgeHQ/sdk-go/proto"

// ContextValueGetter provides access to context values by property name
type ContextValueGetter interface {
	GetContextValue(propertyName string) (value interface{}, valueExists bool)
}

// ProjectEnvIDSupplier provides access to the project environment ID
type ProjectEnvIDSupplier interface {
	GetProjectEnvID() int64
}

// ConfigStoreGetter defines the interface for custom config stores that can be plugged into the SDK.
// This allows external systems to provide config data without going through the standard API sources.
type ConfigStoreGetter interface {
	GetConfig(key string) (config *prefabProto.Config, exists bool)
	Keys() []string
	ContextValueGetter
	ProjectEnvIDSupplier
}