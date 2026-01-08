# Send to Honeycomb

**Kind:** `HoneycombExporter` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This component sends traces, logs, metrics, and Honeycomb-formatted events to the Honeycomb's datastore for real-time analysis.tags:- category:output

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-honeycomb_exporter
    kind: HoneycombExporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
