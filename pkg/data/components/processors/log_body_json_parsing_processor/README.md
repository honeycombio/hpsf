# Parse Log Body As JSON

**Kind:** `LogBodyJSONParsingProcessor` | **Version:** `v0.0.1` | **Status:** alpha

## Overview

description: |-Specifically designed to parse the log.body field as JSON and flatten the parsed JSON intoindividual log attributes. This processor has no configuration parameters and only works with logs.tags:- category:processor

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-log_body_j_s_o_n_parsing_processor
    kind: LogBodyJSONParsingProcessor
```

## Changelog

### v0.0.1 (2026-01-08)
- Component migrated to directory structure
