# Transform Metrics with Custom OTTL

**Kind:** `CustomMetricTransformProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Perform custom modifications to metric datapoints using configuration statements written in OpenTelemetryTransformation Language (OTTL). This is an advanced component. No OTTL validation is provided.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-custom_metric_transform_processor
    kind: CustomMetricTransformProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
