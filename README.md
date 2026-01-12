[![OSS Lifecycle](https://img.shields.io/osslifecycle/honeycombio/hpsf?color/yellow)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)
[![GoDoc](https://godoc.org/github.com/honeycombio/hpsf?status.svg)](https://godoc.org/github.com/honeycombio/hpsf)

# HPSF -- EXPERIMENTAL!

## What it is

HPSF is an experimental format for a configuration language.

It will undergo radical changes for a while; please don't depend on it yet.

# hpsf

Here are some sample commands:

* go run ./cmd/hpsf -i examples/hpsf.yaml validate
* go run ./cmd/hpsf -i examples/hpsf.yaml rRules
* go run ./cmd/hpsf -i examples/hpsf.yaml rConfig

Here's an example that exercises a separate data table:

`go run ./cmd/hpsf -d API_Key=hello -i examples/hpsf2.yaml rConfig`

## Component Library

HPSF includes 53 pre-built components for telemetry processing. Components are organized by style (receivers/processors/exporters/samplers/conditions/startsampling) for easy exploration and extension.

**Browse Components:** [pkg/data/components/](./pkg/data/components/)

### Component Categories

- **Receivers** - Ingest telemetry (OTel, HTTP, etc.)
- **Processors** - Transform, filter, enrich data
- **Exporters** - Send to destinations (Honeycomb, S3, etc.)
- **Samplers** - Refinery sampling strategies
- **Conditions** - Boolean expressions for sampling rules

### Creating Custom Components

```bash
make new-component
# Follow prompts, edit component.yaml and README.md
make validate-components
```

See [Component Creation Guide](./pkg/data/components/README.md#creating-a-new-component) for:
- Component anatomy and properties
- Property validation and types
- Port configuration
- Template rendering for multiple targets

### Versioning & Migration

Components follow semantic versioning. Major version changes handled via kind suffixes (e.g., `OTelReceiverV2`) to allow coexistence.

See [Migration Guide](./pkg/data/components/MIGRATION_GUIDE.md) for deprecation lifecycle and migration standards.