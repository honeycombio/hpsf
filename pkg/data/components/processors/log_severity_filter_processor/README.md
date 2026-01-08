# Filter Logs by Severity

**Kind:** `LogSeverityFilterProcessor` | **Version:** `v0.0.1` | **Status:** alpha

## Overview

description: Filters logs using the `severity_number` attribute supplied on the logs.tags:- category:processor- service:collector- signal:OTelLogs

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-log_severity_filter_processor
    kind: LogSeverityFilterProcessor
```

## Changelog

### v0.0.1 (2026-01-08)
- Component migrated to directory structure
