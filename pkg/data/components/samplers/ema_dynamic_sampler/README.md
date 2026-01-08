# Sample Proportionally by Key

**Kind:** `EMADynamicSampler` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-This is an Exponential Moving Average (EMA) sampler designed to achieve a target sample rate basedon trying to achieve representative quantities of the specified sampling keys. The keys should bechosen from fields with relatively low cardinality, such as HTTP method, status code, etc.tags:

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-e_m_a_dynamic_sampler
    kind: EMADynamicSampler
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
