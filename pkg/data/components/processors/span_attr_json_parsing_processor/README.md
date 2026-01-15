# Parse Span Attribute as JSON

**Kind:** `SpanAttrJSONParsingProcessor` | **Version:** `v0.0.1` | **Status:** development

## Overview

description: |Specify a span attribute with a JSON string value to parse into individual attributeson the span. If the attribute is not found or cannot be parsed as JSON, the span willcontinue through the pipeline unchanged.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-span_attr_j_s_o_n_parsing_processor
    kind: SpanAttrJSONParsingProcessor
```

## Changelog

### v0.0.1 (2026-01-08)
- Component migrated to directory structure
