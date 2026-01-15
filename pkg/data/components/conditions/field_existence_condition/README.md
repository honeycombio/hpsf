# Check Field Existence

**Kind:** `FieldExistenceCondition` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: When checking if a field exists, returns true if any span in the trace contains the field. When checking if a field does not exist, returns true if no spans in the trace contain the field.tags:- category:condition- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-field_existence_condition
    kind: FieldExistenceCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
