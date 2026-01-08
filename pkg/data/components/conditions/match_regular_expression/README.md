# Match a regular expression

**Kind:** `MatchRegularExpression` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: This checks if any span in a trace has a specific field that matches a given regular expression.tags:- category:condition- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-match_regular_expression
    kind: MatchRegularExpression
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
