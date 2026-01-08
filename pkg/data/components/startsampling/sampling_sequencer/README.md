# Start Sampling

**Kind:** `SamplingSequencer` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-Converts traces and logs to a format for advanced tail-based sampling with Refinery. Also containssome advanced configuration options for sending data to Refinery; most installations will not needto change these.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-sampling_sequencer
    kind: SamplingSequencer
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
