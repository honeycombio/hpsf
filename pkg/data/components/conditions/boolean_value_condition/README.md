# Check a Boolean Value

**Kind:** `BooleanValueCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-Check if any span in a trace has a specific boolean field.If the boolean field matches the specified value the condition will evaluate to true.tags:- category:condition

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-boolean_value_condition
    kind: BooleanValueCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
