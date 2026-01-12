# Compare a numeric field

**Kind:** `CompareDecimalFieldCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This checks if any span in a trace has a specific numeric field that compares appropriately to thespecified value.tags:- category:condition

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-compare_decimal_field_condition
    kind: CompareDecimalFieldCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
