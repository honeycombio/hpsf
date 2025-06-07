package collectorprovider

import (
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pipeline"
)

type ComponentGetResult struct {
	Found        bool
	SearchString string
	Components   []string
}

func GetProcessorConfig[T any](config *otelcol.Config, processorId string) (*T, ComponentGetResult) {
	return getTypedComponent[T](config.Processors, processorId)
}

func GetExporterConfig[T any](config *otelcol.Config, exporterId string) (*T, ComponentGetResult) {
	return getTypedComponent[T](config.Exporters, exporterId)
}

func GetReceiverConfig[T any](cfg *otelcol.Config, receiverId string) (*T, ComponentGetResult) {
	return getTypedComponent[T](cfg.Receivers, receiverId)
}

func GetExtensionConfig[T any](cfg *otelcol.Config, extensionId string) (*T, ComponentGetResult) {
	return getTypedComponent[T](cfg.Extensions, extensionId)
}

func getTypedComponent[T any](components map[component.ID]component.Config, componentId string) (*T, ComponentGetResult) {
	typeAndName := strings.Split(componentId, "/")
	var typedComponentId component.ID
	if len(typeAndName) == 2 {
		typedComponentId = component.MustNewIDWithName(typeAndName[0], typeAndName[1])
	} else {
		typedComponentId = component.MustNewID(typeAndName[0])
	}
	genericComponentConfig, componentExists := components[typedComponentId]
	if !componentExists {
		return nil, ComponentGetResult{Found: false, SearchString: componentId, Components: listComponents(components)}
	}
	typedConfig, componentConverted := genericComponentConfig.(*T)
	if componentConverted {
		return typedConfig, ComponentGetResult{Found: true, SearchString: componentId}
	}
	return nil, ComponentGetResult{Found: false, SearchString: componentId, Components: listComponents(components)}
}

func listComponents(components map[component.ID]component.Config) []string {
	componentList := make([]string, 0)
	for name, _ := range components {
		componentList = append(componentList, name.String())
	}
	return componentList
}

func GetPipelineConfig(collectorConfig *otelcol.Config, pipelineName string) (receivers, processors, exporters []string, getResult ComponentGetResult) {
	typeAndName := strings.Split(pipelineName, "/")
	var typedPipelineId pipeline.ID
	if len(typeAndName) == 2 {
		typedPipelineId = pipeline.NewIDWithName(convertTypeNameToSignal(typeAndName[0]), typeAndName[1])
	} else {
		typedPipelineId = pipeline.NewID(convertTypeNameToSignal(typeAndName[0]))
	}

	pipeline, pipelineExists := collectorConfig.Service.Pipelines[typedPipelineId]
	if !pipelineExists {
		availablePipelines := make([]string, 0)
		for name, _ := range collectorConfig.Service.Pipelines {
			availablePipelines = append(availablePipelines, name.String())
		}
		return nil, nil, nil, ComponentGetResult{Found: false, SearchString: pipelineName, Components: availablePipelines}
	}
	return getComponentNames(pipeline.Exporters), getComponentNames(pipeline.Processors), getComponentNames(pipeline.Receivers), ComponentGetResult{Found: true, SearchString: pipelineName}

}

func convertTypeNameToSignal(typeName string) pipeline.Signal {
	switch typeName {
	case "logs":
		return pipeline.SignalLogs
	case "metrics":
		return pipeline.SignalMetrics
	case "traces":
		return pipeline.SignalTraces
	default:
		return pipeline.SignalLogs
	}
}

func getComponentNames(components []component.ID) []string {
	componentNames := make([]string, 0)
	for _, component := range components {
		componentNames = append(componentNames, component.String())
	}
	return componentNames
}
