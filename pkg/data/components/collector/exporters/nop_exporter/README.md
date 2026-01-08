# Export Nothing

**Kind:** `NopExporter` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: A simple no-op exporter. This exporter does nothing. It is required for the minimal collector.tags:- category:output- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-nop_exporter
    kind: NopExporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
