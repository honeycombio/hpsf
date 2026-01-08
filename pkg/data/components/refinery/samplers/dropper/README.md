# Drop

**Kind:** `Dropper` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: This sampler drops all traces.tags:- category:sampler- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-dropper
    kind: Dropper
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
