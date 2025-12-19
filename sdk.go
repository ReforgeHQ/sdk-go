// Package reforge provides a client for fetching configuration and feature flags from the Reforge API.
package reforge

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ReforgeHQ/sdk-go/internal"
	"github.com/ReforgeHQ/sdk-go/internal/contexts"
	optionsPkg "github.com/ReforgeHQ/sdk-go/internal/options"
	"github.com/ReforgeHQ/sdk-go/internal/stores"
	"github.com/ReforgeHQ/sdk-go/internal/telemetry"
	"github.com/ReforgeHQ/sdk-go/internal/utils"
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

// ConfigMatch represents a match between a config/flag key and a context. It
// has internal fields that are used by the client to determine the final value
// and report telemetry.
type ConfigMatch = internal.ConfigMatch

// ContextSet is a set of NamedContext
type ContextSet = contexts.ContextSet

// NamedContext is a named context. It is used to provide context to the client about the current user/machine/etc.
type NamedContext = contexts.NamedContext

// NewContextSet creates a new ContextSet
func NewContextSet() *ContextSet {
	return contexts.NewContextSet()
}

const (
	// ReturnError will return an error when checking config/flag values if initialization times out
	ReturnError optionsPkg.OnInitializationFailure = optionsPkg.ReturnError
	// ReturnNilMatch will continue (generally returning a zero value, ok=false result) if initialization times out
	ReturnNilMatch optionsPkg.OnInitializationFailure = optionsPkg.ReturnNilMatch
)

var ContextTelemetryMode = optionsPkg.ContextTelemetryModes

// ClientInterface is the interface for the Prefab client
type ClientInterface interface {
	GetIntValue(key string, contextSet ContextSet) (int64, bool, error)
	GetBoolValue(key string, contextSet ContextSet) (bool, bool, error)
	GetStringValue(key string, contextSet ContextSet) (string, bool, error)
	GetFloatValue(key string, contextSet ContextSet) (float64, bool, error)
	GetStringSliceValue(key string, contextSet ContextSet) ([]string, bool, error)
	GetDurationValue(key string, contextSet ContextSet) (time.Duration, bool, error)
	GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (int64, bool)
	GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (bool, bool)
	GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (string, bool)
	GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (float64, bool)
	GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) ([]string, bool)
	GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (time.Duration, bool)
	GetLogLevelStringValue(key string, contextSet ContextSet) (string, bool, error)
	GetLogLevel(loggerName string) LogLevel
	GetJSONValue(key string, contextSet ContextSet) (interface{}, bool, error)
	GetJSONValueWithDefault(key string, contextSet ContextSet, defaultValue interface{}) (interface{}, bool)
	GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error)
	GetConfig(key string) (*prefabProto.Config, bool)
	FeatureIsOn(key string, contextSet ContextSet) (bool, bool)
	WithContext(contextSet *ContextSet) *ContextBoundClient
	GetInstanceHash() string
}

// ContextBoundClient is a Client bound to a specific context. Any calls to the client will use the context provided.
type ContextBoundClient struct {
	context *ContextSet
	client  *Client
}

// Client is the Prefab client
type Client struct {
	boundClient                     *ContextBoundClient
	options                         *optionsPkg.Options
	configStore                     internal.ConfigStoreGetter
	configResolver                  *internal.ConfigResolver
	initializationComplete          chan struct{}
	closeInitializationCompleteOnce sync.Once
	telemetry                       telemetry.Submitter
	instanceHash                    string
}

// NewSdk creates a new Reforge SDK. It takes options as arguments (e.g. WithSdkKey)
func NewSdk(opts ...Option) (*Client, error) {
	options := optionsPkg.GetDefaultOptions()

	var client Client

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	// Normalize SDK key from env var if not explicitly set
	// This ensures all components can safely use options.SdkKey
	_, err := options.SdkKeySettingOrEnvVar()
	if err != nil {
		return nil, err
	}

	slog.Debug("Initializing client", "options", options)

	var configStores []internal.ConfigStoreGetter

	// Add custom stores first (they take precedence)
	for _, customStore := range options.CustomStores {
		if store, ok := customStore.(ConfigStoreGetter); ok {
			configStores = append(configStores, store)
		}
	}

	apiSourceFinishedLoading := func() {
		client.closeInitializationCompleteOnce.Do(func() {
			close(client.initializationComplete)
		})
	}

	anyAsync := false

	for _, source := range options.Sources {
		configStore, asyncInit, err := stores.BuildConfigStore(options, source, apiSourceFinishedLoading)
		if err != nil {
			return nil, err
		}

		if asyncInit {
			anyAsync = true
		}

		configStores = append(configStores, configStore)
	}

	if (len(options.Sources) > 1 || (len(options.Sources) == 1 && options.Sources[0].Raw != optionsPkg.MemoryStoreKey)) && len(options.Configs) > 0 {
		return nil, errors.New("cannot use WithConfigs with other sources")
	}

	if len(options.CustomStores) > 0 && len(options.Configs) > 0 {
		return nil, errors.New("cannot use WithConfigs with custom stores")
	}

	configStore := stores.BuildCompositeConfigStore(configStores...)

	configResolver := internal.NewConfigResolver(configStore)

	// Apply custom EnvLookup if provided
	if options.CustomEnvLookup != nil {
		if envLookup, ok := options.CustomEnvLookup.(internal.EnvLookup); ok {
			configResolver.EnvLookup = envLookup
		}
	}

	client = Client{
		options:                &options,
		configStore:            configStore,
		configResolver:         configResolver,
		initializationComplete: make(chan struct{}),
		telemetry:              *telemetry.NewTelemetrySubmitter(options),
		instanceHash:           options.InstanceHash,
	}

	if !anyAsync {
		client.closeInitializationCompleteOnce.Do(func() {
			close(client.initializationComplete)
		})
	}

	client.boundClient = &ContextBoundClient{client: &client, context: options.GlobalContext}

	if options.TelemetryEnabled() {
		client.telemetry.StartPeriodicSubmission(options.TelemetrySyncInterval)
	}

	return &client, nil
}

// GetIntValue returns an int value for a given key and context
func (c *Client) GetIntValue(key string, contextSet ContextSet) (value int64, ok bool, err error) {
	return c.boundClient.GetIntValue(key, contextSet)
}

// GetBoolValue returns a bool value for a given key and context
func (c *Client) GetBoolValue(key string, contextSet ContextSet) (value bool, ok bool, err error) {
	return c.boundClient.GetBoolValue(key, contextSet)
}

// GetStringValue returns a string value for a given key and context
func (c *Client) GetStringValue(key string, contextSet ContextSet) (value string, ok bool, err error) {
	return c.boundClient.GetStringValue(key, contextSet)
}

// GetFloatValue returns a float value for a given key and context
func (c *Client) GetFloatValue(key string, contextSet ContextSet) (value float64, ok bool, err error) {
	return c.boundClient.GetFloatValue(key, contextSet)
}

// GetStringSliceValue returns a string slice value for a given key and context
func (c *Client) GetStringSliceValue(key string, contextSet ContextSet) (value []string, ok bool, err error) {
	return c.boundClient.GetStringSliceValue(key, contextSet)
}

// GetDurationValue returns a duration value for a given key and context
func (c *Client) GetDurationValue(key string, contextSet ContextSet) (value time.Duration, ok bool, err error) {
	return c.boundClient.GetDurationValue(key, contextSet)
}

// GetJSONValue returns a JSON value for a given key and context
func (c *Client) GetJSONValue(key string, contextSet ContextSet) (value interface{}, ok bool, err error) {
	return c.boundClient.GetJSONValue(key, contextSet)
}

// GetIntValueWithDefault returns an int value for a given key and context, with a default value if the key does not exist
func (c *Client) GetIntValueWithDefault(key string, contextSet ContextSet, defaultValue int64) (value int64, wasFound bool) {
	return c.boundClient.GetIntValueWithDefault(key, contextSet, defaultValue)
}

// GetBoolValueWithDefault returns a bool value for a given key and context, with a default value if the key does not exist
func (c *Client) GetBoolValueWithDefault(key string, contextSet ContextSet, defaultValue bool) (value bool, wasFound bool) {
	return c.boundClient.GetBoolValueWithDefault(key, contextSet, defaultValue)
}

// GetStringValueWithDefault returns a string value for a given key and context, with a default value if the key does not exist
func (c *Client) GetStringValueWithDefault(key string, contextSet ContextSet, defaultValue string) (value string, wasFound bool) {
	return c.boundClient.GetStringValueWithDefault(key, contextSet, defaultValue)
}

// GetFloatValueWithDefault returns a float value for a given key and context, with a default value if the key does not exist
func (c *Client) GetFloatValueWithDefault(key string, contextSet ContextSet, defaultValue float64) (value float64, wasFound bool) {
	return c.boundClient.GetFloatValueWithDefault(key, contextSet, defaultValue)
}

// GetStringSliceValueWithDefault returns a string slice value for a given key and context, with a default value if the key does not exist
func (c *Client) GetStringSliceValueWithDefault(key string, contextSet ContextSet, defaultValue []string) (value []string, wasFound bool) {
	return c.boundClient.GetStringSliceValueWithDefault(key, contextSet, defaultValue)
}

// GetDurationWithDefault returns a duration value for a given key and context, with a default value if the key does not exist
func (c *Client) GetDurationWithDefault(key string, contextSet ContextSet, defaultValue time.Duration) (value time.Duration, wasFound bool) {
	return c.boundClient.GetDurationWithDefault(key, contextSet, defaultValue)
}

// GetJSONValueWithDefault returns a JSON value for a given key and context, with a default value if the key does not exist
func (c *Client) GetJSONValueWithDefault(key string, contextSet ContextSet, defaultValue interface{}) (value interface{}, wasFound bool) {
	return c.boundClient.GetJSONValueWithDefault(key, contextSet, defaultValue)
}

// FeatureIsOn returns a bool indicating if a feature is on for a given key and context. It will default to false if the key does not exist.
func (c *Client) FeatureIsOn(key string, contextSet ContextSet) (result bool, wasFound bool) {
	return c.boundClient.FeatureIsOn(key, contextSet)
}

// GetLogLevelStringValue returns a string value for a given key and context, representing a log level.
func (c *Client) GetLogLevelStringValue(key string, contextSet ContextSet) (result string, ok bool, err error) {
	return c.boundClient.GetLogLevelStringValue(key, contextSet)
}

// GetLogLevel returns the log level for a given logger name. It evaluates the logger key (from options)
// with a context containing the logger name and returns the appropriate LogLevel. If no config is found or
// an error occurs, it returns Debug as the default level.
func (c *Client) GetLogLevel(loggerName string) LogLevel {
	return c.boundClient.GetLogLevel(loggerName)
}

// WithContext returns a new ContextBoundClient bound to the provided context (merged with the parent context)
func (c *Client) WithContext(contextSet *ContextSet) *ContextBoundClient {
	mergedContext := contexts.Merge(c.options.GlobalContext, contextSet)

	return &ContextBoundClient{context: mergedContext, client: c}
}

// GetConfig returns a Config object for a given key. You're unlikely to need this method.
func (c *Client) GetConfig(key string) (*prefabProto.Config, bool) {
	return c.boundClient.GetConfig(key)
}

// GetConfigMatch returns a ConfigMatch object for a given key and context. You're unlikely to need this method.
func (c *Client) GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error) {
	return c.boundClient.GetConfigMatch(key, contextSet)
}

// Keys returns a list of all keys in the config store
func (c *Client) Keys() ([]string, error) {
	if c.awaitInitialization() == timeout {
		switch c.options.OnInitializationFailure {
		case ReturnNilMatch:
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		case ReturnError:
			return []string{}, errors.New("initialization timeout")
		}
	}

	return c.configResolver.Keys(), nil
}

// GetInstanceHash returns the instance hash for the client
func (c *Client) GetInstanceHash() string {
	return c.instanceHash
}

func clientInternalGetValueFunc[T any](contextBoundClient *ContextBoundClient, key string, contextSet contexts.ContextSet, parseFunc func(*prefabProto.ConfigValue) (T, bool)) (T, bool, error) {
	var zeroValue T

	mergedContextSet := *contexts.Merge(contextBoundClient.context, &contextSet)

	contextBoundClient.client.telemetry.RecordContext(&mergedContextSet)

	fetchResult, fetchOk, fetchErr := contextBoundClient.fetchAndProcessValue(key, mergedContextSet, func(cv *prefabProto.ConfigValue) (any, bool) {
		pVal, pOk := clientParseValueWrapper(cv, parseFunc)
		if !pOk {
			return nil, false
		}

		return pVal, pOk
	})

	if fetchErr != nil || !fetchOk {
		return zeroValue, false, fetchErr
	}

	if fetchResult == nil {
		return zeroValue, false, nil
	}

	typedValue, ok := (fetchResult).(T)
	if !ok {
		slog.Warn(fmt.Sprintf("unexpected type for %T value: %T", zeroValue, fetchResult))

		return zeroValue, false, nil
	}

	return typedValue, true, nil
}

// SendTelemetry sends telemetry data to the Prefab Cloud API. If
// waitOnQueueToDrain is true, the method will block until the telemetry queue
// is empty. You likely don't want to waitOnQueueToDrain in a production
// environment unless you're shutting down and confident you're not further
// enqueueing telemetry.
func (c *Client) SendTelemetry(waitOnQueueToDrain bool) error {
	return c.telemetry.Submit(waitOnQueueToDrain)
}

// SendTelemetry sends telemetry data to the Prefab Cloud API. If
// waitOnQueueToDrain is true, the method will block until the telemetry queue
// is empty. You likely don't want to waitOnQueueToDrain in a production
// environment unless you're shutting down and confident you're not further
// enqueueing telemetry.
func (c *ContextBoundClient) SendTelemetry(waitOnQueueToDrain bool) error {
	return c.client.telemetry.Submit(waitOnQueueToDrain)
}

// GetIntValueWithDefault returns an int value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetIntValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue int64) (value int64, wasFound bool) {
	value, ok, err := c.GetIntValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetIntValue returns an int value for a given key and context
func (c *ContextBoundClient) GetIntValue(key string, contextSet contexts.ContextSet) (value int64, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractIntValue)
}

// GetStringValueWithDefault returns a string value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetStringValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue string) (value string, wasFound bool) {
	value, ok, err := c.GetStringValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetStringValue returns a string value for a given key and context
func (c *ContextBoundClient) GetStringValue(key string, contextSet contexts.ContextSet) (value string, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractStringValue)
}

// GetJSONValue returns a JSON value for a given key and context
func (c *ContextBoundClient) GetJSONValue(key string, contextSet contexts.ContextSet) (value interface{}, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractJSONValueWithoutError)
}

// GetJSONValueWithDefault returns a JSON value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetJSONValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue interface{}) (value interface{}, wasFound bool) {
	value, ok, err := c.GetJSONValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// FeatureIsOn returns a bool indicating if a feature is on for a given key and context. It will default to false if the key does not exist.
func (c *ContextBoundClient) FeatureIsOn(key string, contextSet contexts.ContextSet) (result bool, wasFound bool) {
	value, ok := c.GetBoolValueWithDefault(key, contextSet, false)

	return value, ok
}

// GetLogLevelStringValue returns a string value for a given key and context, representing a log level.
func (c *ContextBoundClient) GetLogLevelStringValue(key string, contextSet contexts.ContextSet) (value string, ok bool, err error) {
	rawValue, ok, err := clientInternalGetValueFunc(c, key, contextSet, utils.ExtractLogLevelValue)

	if err != nil || !ok {
		return "", false, err
	}

	return rawValue.String(), true, nil
}

// GetLogLevel returns the log level for a given logger name. It evaluates the logger key (from options)
// with a context containing the logger name and returns the appropriate LogLevel. If no config is found,
// the config is not LOG_LEVEL_V2 type, or an error occurs, it returns Debug as the default level.
func (c *ContextBoundClient) GetLogLevel(loggerName string) LogLevel {
	// Create context with reforge-sdk-logging
	loggerContext := NewContextSet().WithNamedContextValues("reforge-sdk-logging", map[string]any{
		"lang":        "go",
		"logger-path": loggerName,
	})

	// Get the config match for the logger key
	configMatch, err := c.GetConfigMatch(c.client.options.LoggerKey, *loggerContext)
	if err != nil || configMatch == nil {
		return Debug // Default to Debug if config not found or error
	}

	// Get the full config to check the type
	config, found := c.GetConfig(c.client.options.LoggerKey)
	if !found || config.ConfigType != prefabProto.ConfigType_LOG_LEVEL_V2 {
		return Debug // Default to Debug if not LOG_LEVEL_V2 type
	}

	// Extract the log level value
	protoLogLevel, ok := utils.ExtractLogLevelValue(configMatch.OriginalMatch)
	if !ok {
		return Debug // Default to Debug if extraction fails
	}

	// Convert proto LogLevel to our SDK LogLevel
	return protoLogLevelToLogLevel(protoLogLevel)
}

// GetBoolValueWithDefault returns a bool value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetBoolValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue bool) (value bool, wasFound bool) {
	value, ok, err := c.GetBoolValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetBoolValue returns a bool value for a given key and context
func (c *ContextBoundClient) GetBoolValue(key string, contextSet contexts.ContextSet) (value bool, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractBoolValue)
}

// GetFloatValueWithDefault returns a float value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetFloatValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue float64) (value float64, wasFound bool) {
	value, ok, err := c.GetFloatValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetFloatValue returns a float value for a given key and context
func (c *ContextBoundClient) GetFloatValue(key string, contextSet contexts.ContextSet) (value float64, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractFloatValue)
}

// GetStringSliceValueWithDefault returns a string slice value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetStringSliceValueWithDefault(key string, contextSet contexts.ContextSet, defaultValue []string) (value []string, wasFound bool) {
	value, ok, err := c.GetStringSliceValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetStringSliceValue returns a string slice value for a given key and context
func (c *ContextBoundClient) GetStringSliceValue(key string, contextSet contexts.ContextSet) (value []string, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractStringListValue)
}

// GetDurationWithDefault returns a duration value for a given key and context, with a default value if the key does not exist
func (c *ContextBoundClient) GetDurationWithDefault(key string, contextSet contexts.ContextSet, defaultValue time.Duration) (value time.Duration, wasFound bool) {
	value, ok, err := c.GetDurationValue(key, contextSet)
	if err != nil || !ok {
		return defaultValue, true
	}

	return value, ok
}

// GetDurationValue returns a duration value for a given key and context
func (c *ContextBoundClient) GetDurationValue(key string, contextSet contexts.ContextSet) (value time.Duration, ok bool, err error) {
	return clientInternalGetValueFunc(c, key, contextSet, utils.ExtractDurationValue)
}

func (c *ContextBoundClient) fetchAndProcessValue(key string, contextSet contexts.ContextSet, parser utils.ExtractValueFunction) (any, bool, error) {
	getResult, err := c.client.internalGetValue(key, contextSet)
	if err != nil {
		return nil, false, err
	}

	if getResult.match.Match == nil {
		return nil, false, errors.New("config did not produce a result and no default is specified")
	}

	parsedValue, ok := parser(getResult.match.Match)
	if !ok {
		return nil, false, nil
	}

	return parsedValue, true, nil
}

// WithContext returns a new ContextBoundClient bound to the provided context (merged with the parent context)
func (c *ContextBoundClient) WithContext(contextSet *ContextSet) *ContextBoundClient {
	mergedContext := contexts.Merge(c.context, contextSet)

	return &ContextBoundClient{context: mergedContext, client: c.client}
}

// GetConfig returns a Config object for a given key. You're unlikely to need this method.
func (c *ContextBoundClient) GetConfig(key string) (*prefabProto.Config, bool) {
	return c.client.configStore.GetConfig(key)
}

// GetConfigMatch returns a ConfigMatch object for a given key and context. You're unlikely to need this method.
func (c *ContextBoundClient) GetConfigMatch(key string, contextSet ContextSet) (*ConfigMatch, error) {
	mergedContextSet := *contexts.Merge(c.context, &contextSet)
	getResult, err := c.client.internalGetValue(key, mergedContextSet)
	if err != nil {
		return nil, err
	}

	return &getResult.match, nil
}

// GetInstanceHash returns the instance hash for the client
func (c *ContextBoundClient) GetInstanceHash() string {
	return c.client.GetInstanceHash()
}

func (c *Client) internalGetValue(key string, contextSet contexts.ContextSet) (resolutionResult, error) {
	if c.awaitInitialization() == timeout {
		switch c.options.OnInitializationFailure {
		case optionsPkg.ReturnNilMatch:
			c.closeInitializationCompleteOnce.Do(func() {
				close(c.initializationComplete)
			})
		case optionsPkg.ReturnError:
			return resolutionResultError(), errors.New("initialization timeout")
		}
	}

	match, err := c.configResolver.ResolveValue(key, &contextSet)
	if err != nil {
		return resolutionResultError(), err
	}

	c.telemetry.RecordEvaluation(match)

	return resolutionResultSuccess(match), nil
}

type resolutionResultType int

const (
	resolutionResultTypeSuccess resolutionResultType = iota
	resolutionResultTypeError
)

type resolutionResult struct {
	match      ConfigMatch
	resultType resolutionResultType
}

func resolutionResultError() resolutionResult {
	return resolutionResult{
		resultType: resolutionResultTypeError,
	}
}

func resolutionResultSuccess(match ConfigMatch) resolutionResult {
	return resolutionResult{
		resultType: resolutionResultTypeSuccess,
		match:      match,
	}
}

func clientParseValueWrapper[T any](cv *prefabProto.ConfigValue, parseFunc func(*prefabProto.ConfigValue) (T, bool)) (T, bool) {
	if cv != nil {
		pValue, pOk := parseFunc(cv)

		return pValue, pOk
	}

	var zeroValue T

	return zeroValue, false
}

func (c *Client) awaitInitialization() awaitInitializationResult {
	select {
	case <-c.initializationComplete:
		return success
	case <-time.After(time.Duration(c.options.InitializationTimeoutSeconds) * time.Second):
		slog.Warn(fmt.Sprintf("%f second timeout expired, proceeding without waiting further. Configure in options `InitializationTimeoutSeconds`", c.options.InitializationTimeoutSeconds))

		return timeout
	}
}

type awaitInitializationResult int

const (
	success awaitInitializationResult = iota
	timeout
)

// ExtractValue extracts the underlying value from a ConfigValue. You're unlikely to need this method.
func ExtractValue(cv *prefabProto.ConfigValue) (any, bool, error) {
	return utils.ExtractValue(cv)
}
