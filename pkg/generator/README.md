# Refinery to HPSF Converter

This package provides functionality to generate HPSF (Honeycomb Pipeline Specification Format) workflows from Refinery sampling rule files.

## Overview

The `generator` package allows you to automatically convert existing Refinery sampling rules into HPSF workflows, making it easier to migrate from Refinery to HPSF-based pipeline management. The package uses sensible defaults for configuration, requiring only the sampling rules file as input.

## Features

- **Simplified API**: Only requires a Refinery rules file - configuration uses sensible defaults
- **Automatic component generation**: Creates appropriate HPSF components based on Refinery rules
- **OpenTelemetry Collector Receiver**: Generates OTel receiver components with standard ports (4317 gRPC, 4318 HTTP)
- **Start Sampling component**: Creates sampling sequencer components for data processing
- **Condition components**: Converts Refinery sampling conditions into HPSF condition components
- **Sampler components**: Supports various Refinery samplers (Deterministic, EMA Throughput, EMA Dynamic, Rules-based)
- **Honeycomb Exporter**: Automatically creates Honeycomb exporter components with default settings
- **Proper connections**: Automatically connects components based on data flow requirements

## Supported Refinery Features

### Samplers
- **DeterministicSampler**: Fixed-rate sampling based on trace ID
- **EMAThroughputSampler**: Targets a specific throughput rate
- **EMADynamicSampler**: Dynamic sampling based on key frequency
- **RulesBasedSampler**: Complex conditional sampling with rules

### Conditions
- **Field existence**: `exists`, `does-not-exist` operators
- **Field comparison**: `=`, `!=`, `<`, `<=`, `>`, `>=` operators for strings, integers, floats, and booleans
- **String operations**: `contains`, `starts-with` operators
- **HTTP status conditions**: Specialized handling for HTTP status codes
- **Multiple field support**: Conditions can check multiple fields

### Default Configuration
The package uses the following default values:
- **Honeycomb Exporter**: API endpoint set to `api.honeycomb.io` (API key configuration left to user)
- **Receivers**: OTel receiver on `0.0.0.0:4317` (gRPC) and `0.0.0.0:4318` (HTTP)
- **Start Sampling**: Refinery service on `refinery:8080`

## Usage

### Command Line Interface

The functionality is available through the dedicated `refinery2hpsf` command:

```bash
# Generate HPSF workflow from Refinery rules file
go run ./cmd/refinery2hpsf \
  --refinery-rules tests/refinery2hpsf/01-simple-refinery.yaml \
  -o output-workflow.yaml \
  -v
```

### Programmatic API

```go
import "github.com/honeycombio/hpsf/pkg/generator"

// Generate from file
workflow, err := generator.GenerateFromFile("rules.yaml")

// Generate from raw data
workflow, err := generator.GenerateFromBytes(rulesData)

// Write to file
err = generator.WriteWorkflowToFile(workflow, "output.yaml")
```

### Directory-based Generation

```go
// Automatically find and use Refinery rules files in a directory
workflow, err := generator.GenerateFromDirectory("/path/to/refinery/configs")
```

## Generated Workflow Structure

The generated HPSF workflow follows this general structure:

1. **OpenTelemetry Collector Receiver** - Receives traces and logs
2. **Start Sampling Component** - Converts data for sampling pipeline
3. **Condition Components** - Filter data based on rules
4. **Sampler Components** - Apply sampling logic
5. **Honeycomb Exporter** - Send sampled data to Honeycomb

## Example

### Input Refinery Rules

**rules.yaml:**
```yaml
RulesVersion: 2
Samplers:
  __default__:
    RulesBasedSampler:
      Rules:
        - Name: "Error Traces"
          SampleRate: 1
          Conditions:
            - Fields: ["error"]
              Operator: exists
        - Name: "Default"
          SampleRate: 100
```

### Generated HPSF Workflow

```yaml
kind: HPSF
version: v1
name: Generated_Refinery_Workflow
components:
  - name: OTel_Receiver_1
    kind: OTelReceiver
    properties:
      - name: Host
        value: 0.0.0.0
      - name: GRPCPort
        value: 4317
      - name: HTTPPort
        value: 4318

  - name: Start_Sampling_2
    kind: SamplingSequencer
    properties:
      - name: Host
        value: refinery
      - name: Port
        value: 8080

  - name: Condition_Error_Traces_3
    kind: FieldExistsCondition
    properties:
      - name: Field
        value: error
      - name: ShouldExist
        value: true

  - name: Sample_Error_Traces_4
    kind: DeterministicSampler
    properties:
      - name: SampleRate
        value: 1

  - name: Send_to_Honeycomb_5
    kind: HoneycombExporter
    properties:
      - name: APIEndpoint
        value: api.honeycomb.io

connections:
  # ... appropriate connections between components
```

## Validation

Generated workflows are automatically validated against HPSF schema requirements. Any validation warnings are displayed during generation.

## Limitations

- Only supports `__default__` environment in rules
- Complex nested conditions are simplified to single conditions
- Honeycomb exporter is always generated but API key configuration is left to the user
- Some advanced Refinery features may not have direct HPSF equivalents

## Error Handling

The converter includes robust error handling for:
- Invalid YAML files
- Missing rules file
- Unsupported sampler types
- File system errors
- Validation errors