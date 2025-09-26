# Refinery to HPSF Converter

This package provides functionality to generate HPSF (Honeycomb Pipeline Specification Format) workflows from Refinery sampling rule files.

## Overview

The `generator` package allows you to automatically convert existing Refinery sampling rules into HPSF workflows, making it easier to migrate from Refinery to HPSF-based pipeline management. The package uses sensible defaults for configuration, requiring only the sampling rules file as input.

## Features

- **Automatic component generation**: Creates appropriate HPSF components based on Refinery rules
- **OpenTelemetry Collector Receiver**: Generates OTel receiver components
- **Start Sampling component**: Creates sampling sequencer components for data processing
- **Condition components**: Converts Refinery sampling conditions into HPSF condition components
- **Sampler components**: Supports various Refinery samplers (Deterministic, EMA Throughput, EMA Dynamic, Rules-based)
- **Honeycomb Exporter**: Automatically creates Honeycomb exporter components
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

### Component Generation
The generator creates clean HPSF components following official template patterns:
- **Honeycomb Exporter**: Clean exporter component (configuration managed by HPSF runtime)
- **Receivers**: Clean OTel receiver component (ports and settings managed by HPSF runtime)
- **Start Sampling**: Clean sampling sequencer component (endpoints managed by HPSF runtime)

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

// Generate from raw data
workflow, err := generator.GenerateFromBytes(rulesData)
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

  - name: Start_Sampling_2
    kind: SamplingSequencer

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

connections:
  # ... appropriate connections between components
```

## Validation

Generated workflows are automatically validated against HPSF schema requirements. Any validation warnings are displayed during generation.

## Limitations

- Only supports `__default__` environment in rules
- Complex nested conditions are simplified to single conditions
- Some advanced Refinery features may not have direct HPSF equivalents

## Error Handling

The converter includes robust error handling for:
- Invalid YAML files
- Missing rules file
- Unsupported sampler types
- File system errors
- Validation errors
