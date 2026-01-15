# Compare with a List of Values

**Kind:** `ListComparisonCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-Check if any span in a trace has a field value that is contained in (or excluded from) a list of specified values.tags:- category:condition- service:refinery

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-list_comparison_condition
    kind: ListComparisonCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
