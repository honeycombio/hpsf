# Check for Root Span

**Kind:** `RootSpanCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This checks if the trace has or does not have a root span based on the HasRootSpan property. This istypically used to ensure that the trace has been fully received before being sent for a samplingdecision.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-root_span_condition
    kind: RootSpanCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
