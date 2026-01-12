# Parse Attribute As JSON

**Kind:** `AttributeJSONParsingProcessor` | **Version:** `v0.0.1` | **Status:** alpha

## Overview

description: Takes any attribute from a log or span and parses it as JSON into individual attributestags:- category:processor- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-attribute_j_s_o_n_parsing_processor
    kind: AttributeJSONParsingProcessor
```

## Changelog

### v0.0.1 (2026-01-08)
- Component migrated to directory structure
