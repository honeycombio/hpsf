# Check a Numeric Field

**Kind:** `NumericValueCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-Check if any span of a trace has a specific int or num field that compares appropriately to the specified value.tags:- category:condition- service:refinery

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-numeric_value_condition
    kind: NumericValueCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
