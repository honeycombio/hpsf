# Check That Field Starts With

**Kind:** `FieldStartsWithCondition` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: This checks if any span in a trace has a specific field that starts with a given prefix.tags:- category:condition- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-field_starts_with_condition
    kind: FieldStartsWithCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
