# Transform with Custom OTTL

**Kind:** `CustomTransformProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Perform custom modifications to your telemetry data using configuration statementswritten in OpenTelemetry Transformation Language (OTTL). This is an advanced component.No OTTL validation is provided.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-custom_transform_processor
    kind: CustomTransformProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
