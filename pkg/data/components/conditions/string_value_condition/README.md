# Check a String Value

**Kind:** `StringValueCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: |-This checks if any span in a trace has a specific string field that compares appropriately to thespecified value.tags:- category:condition

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-string_value_condition
    kind: StringValueCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
