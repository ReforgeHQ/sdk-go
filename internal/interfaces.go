package internal

import (
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

type ConfigParser interface {
	Parse(data []byte) ([]*prefabProto.Config, int64, error)
}

// Re-export the public interfaces for internal use
type ContextValueGetter = interface {
	GetContextValue(propertyName string) (value interface{}, valueExists bool)
}

type ProjectEnvIDSupplier = interface {
	GetProjectEnvID() int64
}

type ConfigStoreGetter = interface {
	GetConfig(key string) (config *prefabProto.Config, exists bool)
	Keys() []string
	ContextValueGetter
	ProjectEnvIDSupplier
}

type ConfigEvaluator interface {
	EvaluateConfig(config *prefabProto.Config, contextSet ContextValueGetter) (match ConditionMatch)
}

type Decrypter interface {
	DecryptValue(secretKey string, value string) (decryptedValue string, err error)
}

type Randomer interface {
	Float64() float64
}

type Hasher interface {
	HashZeroToOne(value string) (zeroToOne float64, ok bool)
}

type WeightedValueResolverIF interface {
	Resolve(weightedValues *prefabProto.WeightedValues, propertyName string, contextGetter ContextValueGetter) (valueResult *prefabProto.ConfigValue, index int)
}
type EnvLookup interface {
	LookupEnv(key string) (string, bool)
}
