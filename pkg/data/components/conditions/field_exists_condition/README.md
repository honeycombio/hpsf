# Check if field exists

**Kind:** `FieldExistsCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: This checks if any span in a trace has a specific field that exists or does not exist.tags:- category:condition- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-field_exists_condition
    kind: FieldExistsCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
