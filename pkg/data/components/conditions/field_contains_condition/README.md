# Check That Field Contains a Substring

**Kind:** `FieldContainsCondition` | **Version:** `v0.1.0` | **Status:** deprecated

> **⚠️ DEPRECATED:** This component is deprecated. Use [StringValueCondition](../string_value_condition/README.md) instead.
>
> See [MIGRATIONS.md](./MIGRATIONS.md) for migration guide.

## Overview

description: This checks if any span in a trace has a specific field that contains a given substring.tags:- category:condition- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-field_contains_condition
    kind: FieldContainsCondition
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
