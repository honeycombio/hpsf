# Check Span Duration

**Kind:** `LongDurationCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This condition checks the duration of all spans in a trace. If any span exceeds the specified duration, the condition evaluates to true.tags:- category:condition- service:refinery

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-long_duration_condition
    kind: LongDurationCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
