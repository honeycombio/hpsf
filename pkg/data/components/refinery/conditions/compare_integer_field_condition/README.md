# Compare an integer field

**Kind:** `CompareIntegerFieldCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This checks if any span in a trace has a specific integer field that compares appropriately to thespecified value.tags:- category:condition

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-compare_integer_field_condition
    kind: CompareIntegerFieldCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
