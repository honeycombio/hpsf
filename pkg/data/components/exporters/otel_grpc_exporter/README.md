# Send OTel via gRPC

**Kind:** `OTelGRPCExporter` | **Version:** `v0.1.0` | **Status:** alpha

## Overview

description: Exports OpenTelemetry signals using OTLP via gRPC.tags:- category:output- service:collector- signal:OTelTraces

## Configuration

### Properties

See [component.yaml](./component.yaml) for the full list of configurable properties.

### Ports

See [component.yaml](./component.yaml) for port definitions.

## Examples

### Basic Usage

```yaml
components:
  - name: my-o_tel_g_r_p_c_exporter
    kind: OTelGRPCExporter
```

## Changelog

### v0.1.0 (2026-01-08)
- Component migrated to directory structure
