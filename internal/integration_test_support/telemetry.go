package integrationtestsupport

import (
	"fmt"

	"gopkg.in/yaml.v3"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/ReforgeHQ/sdk-go/internal/contexts"
	"github.com/ReforgeHQ/sdk-go/internal/telemetry"
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

type TelemetryTestCaseYaml struct {
	Aggregator   string      `yaml:"aggregator"`
	Endpoint     string      `yaml:"endpoint"`
	Data         interface{} `yaml:"data"`
	ExpectedData interface{} `yaml:"expected_data"`
	CaseName     string      `yaml:"name"`
	Client       string      `yaml:"client"`
	Function     string      `yaml:"function"`
}

type TelemetryTest struct {
	Cases []TelemetryTestCaseYaml `yaml:"cases"`
}

type TelemetryTestCase struct {
	Contexts   TestCaseContexts
	Aggregator string
	TestName   string
	Err        error
	Yaml       TelemetryTestCaseYaml
	RawYaml    []byte
}

func (t TelemetryTestCase) GetGlobalContexts() *contexts.ContextSet {
	return t.Contexts.Global
}

func (t TelemetryTestCase) GetBlockContexts() *contexts.ContextSet {
	return t.Contexts.Block
}

func (t TelemetryTestCase) GetClientOverrides() *ClientOverridesYaml {
	return nil
}

type TelemetryTestHarness interface {
	GetOptions() []reforge.Option
	GetExpectedEvents() ([]*prefabProto.TelemetryEvent, error)
	Exercise(*reforge.ContextBoundClient) error
	MassagePayload(events *prefabProto.TelemetryEvents) *prefabProto.TelemetryEvents
}

func NewTelemetryTestHarness(testCase TelemetryTestCase) TelemetryTestHarness {
	switch testCase.Aggregator {
	case "example_contexts":
		return ExampleContextTestHarness{testCase: testCase}
	case "context_shape":
		return ContextShapeTestHarness{testCase: testCase}
	case "evaluation_summary":
		return EvaluationSummaryTestHarness{testCase: testCase}
	default:
		fmt.Println("Error: Unknown aggregator type", testCase.Aggregator)

		return nil
	}
}

func unmarshalExampleData(yamlData interface{}, target interface{}) error {
	dataBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(dataBytes, target)
	if err != nil {
		return err
	}

	return nil
}

func MockNowProvider() {
	nowCalls := 0
	telemetry.NowProvider = func() int64 {
		nowCalls++

		return int64(nowCalls)
	}
}
