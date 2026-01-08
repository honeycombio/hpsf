# Symbolicate JavaScript Stack Traces

**Kind:** `SymbolicatorProcessor` | **Version:** `v0.1.0` | **Status:** development

## Overview

description: This processor is used to symbolicate JavaScript stack traces using source maps.tags:- category:processor- service:collector- signal:OtelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-symbolicator_processor
    kind: SymbolicatorProcessor
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
