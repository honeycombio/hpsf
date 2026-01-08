# Sample at a Fixed Rate

**Kind:** `DeterministicSampler` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: A sampler that deterministically samples a fixed fraction of traces based on trace ID.tags:- category:sampler- service:refinery- signal:HoneycombEvents

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-deterministic_sampler
    kind: DeterministicSampler
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
