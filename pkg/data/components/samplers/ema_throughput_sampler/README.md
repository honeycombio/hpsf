# Sample by Events per Second

**Kind:** `EMAThroughputSampler` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

An Exponential Moving Average (EMA) sampler designed to achieve a target throughput (in events persecond).description: |-This is an Exponential Moving Average (EMA) sampler designed to achieve a target throughput (inevents per second) based on trying to achieve representative quantities of the specified sampling

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-e_m_a_throughput_sampler
    kind: EMAThroughputSampler
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
