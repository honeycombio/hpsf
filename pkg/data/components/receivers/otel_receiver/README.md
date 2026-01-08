# Receive OTel

**Kind:** `OTelReceiver` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: |-Imports OTLP signals from OpenTelemetry via gRPC or HTTP. This receiver can be configured to listenon both gRPC and HTTP, or just one.tags:- category:input

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-o_tel_receiver
    kind: OTelReceiver
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
