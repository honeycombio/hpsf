# Filter Logs by Custom OTTL

**Kind:** `CustomLogFilterProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |Apply custom filters to log records using configuration statements written in OpenTelemetryTransformation Language (OTTL). This is an advanced component. No OTTL validation is provided.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-custom_log_filter_processor
    kind: CustomLogFilterProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
