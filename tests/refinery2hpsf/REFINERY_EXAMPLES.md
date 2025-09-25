# Refinery to HPSF Examples

This directory contains examples of how to convert Refinery sampling rules to HPSF workflows using the simplified API.

**Note:** The example Refinery rules files and generated workflows have been organized in `tests/refinery2hpsf/` for integration testing with clear input/output pairing.

## Quick Start

The `refinery2hpsf` command only requires a Refinery rules file:

```bash
go run ./cmd/refinery2hpsf --refinery-rules tests/refinery2hpsf/01-simple-refinery.yaml -o [output-file] -v
```

## Test File Organization

All test files are in the same directory with clear naming that shows the relationship between input and output:

### Test Files (tests/refinery2hpsf/)

**Input Files (Refinery Rules):**
- **`01-simple-refinery.yaml`** - Simple deterministic and EMA throughput samplers
- **`02-complex-refinery.yaml`** - Complex rules-based sampling with conditions
- **`03-comprehensive-refinery.yaml`** - Comprehensive example with all sampler types

**Expected Output Files (HPSF Workflows):**
- **`01-simple-workflow.yaml`** - Expected workflow generated from `01-simple-refinery.yaml`
- **`02-complex-workflow.yaml`** - Expected workflow generated from `02-complex-refinery.yaml`
- **`03-comprehensive-workflow.yaml`** - Expected workflow generated from `03-comprehensive-refinery.yaml`

Each `*-refinery.yaml` file has a corresponding `*-workflow.yaml` file with the same prefix, making it immediately clear which input produces which output.

## Usage Examples

### Command Line Usage

#### Simple Sampling
```bash
# Generate from simple rules (deterministic + EMA throughput samplers)
go run ./cmd/refinery2hpsf \
  --refinery-rules tests/refinery2hpsf/01-simple-refinery.yaml \
  -o simple-workflow.yaml -v
```

#### Complex Rules-based Sampling
```bash
# Generate from complex rules (conditions + multiple samplers)
go run ./cmd/refinery2hpsf \
  --refinery-rules tests/refinery2hpsf/02-complex-refinery.yaml \
  -o complex-workflow.yaml -v
```

#### Comprehensive Example
```bash
# Generate from comprehensive example (all sampler types)
go run ./cmd/refinery2hpsf \
  --refinery-rules tests/refinery2hpsf/03-comprehensive-refinery.yaml \
  -o comprehensive-workflow.yaml -v
```

### Makefile Targets

#### Generate Single Workflow
```bash
# Generate with default rules (tests/refinery2hpsf/02-complex-refinery.yaml -> tmp/generated-workflow.yaml)
make generate-workflow

# Generate with custom rules and output
make generate-workflow RULES=tests/refinery2hpsf/01-simple-refinery.yaml OUTPUT=my-workflow.yaml
```

#### Generate All Example Workflows
```bash
# Generate workflows from all example Refinery rules files
make generate-workflows-all
```

This will generate and validate workflows for:
- `tests/refinery2hpsf/01-simple-refinery.yaml` → `tmp/01-simple-refinery-workflow.yaml`
- `tests/refinery2hpsf/02-complex-refinery.yaml` → `tmp/02-complex-refinery-workflow.yaml`
- `tests/refinery2hpsf/03-comprehensive-refinery.yaml` → `tmp/03-comprehensive-refinery-workflow.yaml`

## Generated Workflow Structure

All generated workflows include:

1. **OTel Receiver** - Receives traces and logs on standard ports (4317 gRPC, 4318 HTTP)
2. **Start Sampling** - Converts data for sampling pipeline
3. **Condition Components** - Filter based on rules (when using RulesBasedSampler)
4. **Sampler Components** - Apply sampling logic
5. **Honeycomb Exporter** - Send sampled data to Honeycomb

## Default Configuration

The generator uses these defaults:
- OTel receiver on `0.0.0.0:4317` (gRPC) and `0.0.0.0:4318` (HTTP)
- Refinery service at `refinery:8080`
- Honeycomb API endpoint at `api.honeycomb.io` (API key configuration left to user)

## Validation

All generated workflows are automatically validated:

```bash
go run ./cmd/hpsf -i [generated-workflow.yaml] validate
```