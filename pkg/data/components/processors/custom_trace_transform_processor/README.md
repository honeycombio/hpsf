# Transform Traces with Custom OTTL

**Kind:** `CustomTraceTransformProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Perform custom modifications to trace spans using configuration statements written in OpenTelemetryTransformation Language (OTTL). This is an advanced component. No OTTL validation is provided.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-custom_trace_transform_processor
    kind: CustomTraceTransformProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
