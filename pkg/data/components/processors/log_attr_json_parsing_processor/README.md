# Parse Log Attribute as JSON

**Kind:** `LogAttrJSONParsingProcessor` | **Version:** `v0.0.1` | **Status:** development

## Overview

description: |Specify a log attribute with a JSON string value to parse into individual attributeson the log record. If the attribute is not found or cannot be parsed as JSON, the logwill continue through the pipeline unchanged.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-log_attr_j_s_o_n_parsing_processor
    kind: LogAttrJSONParsingProcessor
```

## Changelog

### v0.0.1 (2026-01-08)
- Component migrated to directory structure
