# Filter by Custom OTTL

**Kind:** `CustomFilterProcessor` | **Version:** `v0.0.0` | **Status:** development

## Overview

description: Filters traces, metrics, and logs based on rules defined in the configuration.tags:- category:processor- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-custom_filter_processor
    kind: CustomFilterProcessor
```

## Changelog

### v0.0.0 (2026-01-08)
- Component migrated to directory structure
