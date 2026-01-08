# Keep All

**Kind:** `KeepAllSampler` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: This sampler keeps (samples) all traces, setting SampleRate to 1.tags:- category:sampler- service:refinery- vendor:Honeycomb

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-keep_all_sampler
    kind: KeepAllSampler
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
