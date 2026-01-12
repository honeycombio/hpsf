# Send to stdout

**Kind:** `DebugExporter` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-Exports signal data from a pipeline to stdout. This is useful for debugging, but only if you haveaccess to the stdout stream in your environment. This component is not intended for production use.tags:- category:output

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-debug_exporter
    kind: DebugExporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
