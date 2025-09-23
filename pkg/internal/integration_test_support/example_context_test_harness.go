package integrationtestsupport

import (
	reforge "github.com/ReforgeHQ/sdk-go/pkg"
	"github.com/ReforgeHQ/sdk-go/pkg/internal/contexts"
	"github.com/ReforgeHQ/sdk-go/pkg/internal/telemetry"
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

type ExampleContextTestHarness struct {
	testCase TelemetryTestCase
}

func (c ExampleContextTestHarness) GetOptions() []reforge.Option {
	return []reforge.Option{reforge.WithContextTelemetryMode(reforge.ContextTelemetryMode.PeriodicExample)}
}

func (c ExampleContextTestHarness) GetExpectedEvents() ([]*prefabProto.TelemetryEvent, error) {
	if c.testCase.Yaml.ExpectedData == nil {
		return nil, nil
	}

	contextSet := ctxDataToContextSet(c.testCase.Yaml.ExpectedData.(map[string]interface{}))

	return []*prefabProto.TelemetryEvent{
		// This is the primary payload, but example contexts also sends context shapes
		{
			Payload: &prefabProto.TelemetryEvent_ExampleContexts{
				ExampleContexts: &prefabProto.ExampleContexts{
					Examples: []*prefabProto.ExampleContext{
						{
							Timestamp:  0,
							ContextSet: contextSet.ToProto(),
						},
					},
				},
			},
		},
		// This is the context shape payload
		{
			Payload: &prefabProto.TelemetryEvent_ContextShapes{
				ContextShapes: &prefabProto.ContextShapes{
					Shapes: contextShapesForContextSet(contextSet),
				},
			},
		},
	}, nil
}

func (c ExampleContextTestHarness) Exercise(client *reforge.ContextBoundClient) error {
	context := ctxDataToContextSet(c.testCase.Yaml.Data.(map[string]interface{}))

	_, _, err := client.GetIntValue("does.not.exist", *context)

	return err
}

func (c ExampleContextTestHarness) MassagePayload(payload *prefabProto.TelemetryEvents) *prefabProto.TelemetryEvents {
	return payload
}

func contextShapesForContextSet(contextSet *contexts.ContextSet) []*prefabProto.ContextShape {
	shapes := []*prefabProto.ContextShape{}

	for _, context := range contextSet.Data {
		shape := &prefabProto.ContextShape{
			Name:       context.Name,
			FieldTypes: make(map[string]int32),
		}

		for key, value := range context.Data {
			shape.FieldTypes[key] = telemetry.FieldTypeForValue(value)
		}

		shapes = append(shapes, shape)
	}

	return shapes
}
