# Component Migration Guide

Standards for documenting breaking changes & version migrations in HPSF components.

## When to Create Migration Documentation

**Required:**
- Component `status` changes to `deprecated` or `archived`
- Major version bump (v1.x → v2.0)
- Breaking changes to:
  - Component `kind`
  - Port names or types
  - Required property names
  - Property behavior (e.g., changing default from opt-out to opt-in)

**Optional:**
- Minor/patch updates with deprecation warnings
- Significant feature additions
- Performance optimizations requiring config changes

## Migration Documentation Format

### Option 1: MIGRATIONS.md (Single File)

Use for components with <5 major versions or simple migration paths.

**Location:** `components/[component_name]/MIGRATIONS.md`

**Template:**

```markdown
# Migration Guide: [ComponentKind]

## Version History

### v2.0.0 → v3.0.0 (Breaking Changes)

**Status:** Current
**Released:** 2026-01-15
**Deprecation Timeline:** v2.0.0 deprecated 2026-03-01, removed 2026-09-01

#### Breaking Changes

1. **Property Rename:** `OldProperty` → `NewProperty`
   - **Impact:** All configs using `OldProperty` must update
   - **Reason:** Align with OTel semantic conventions

2. **Port Type Change:** `Input` port from `OTelEvent` → `OTelTraces`
   - **Impact:** Connections from components outputting `OTelEvent` will break
   - **Reason:** Narrowed scope to trace-specific processing

#### Migration Steps

**Automated (Recommended):**
```bash
hpsf migrate component --from OldComponentV2 --to OldComponentV3 workflow.yaml
```

**Manual:**

Before (v2.0.0):
```yaml
components:
  - name: my-comp
    kind: OldComponentV2
    properties:
      - name: OldProperty
        value: 100
```

After (v3.0.0):
```yaml
components:
  - name: my-comp
    kind: OldComponentV3
    properties:
      - name: NewProperty
        value: 100
```

#### Deprecation Warnings

Starting v2.1.0, using `OldProperty` emits:
```
WARN: Property 'OldProperty' deprecated, use 'NewProperty' (removed in v3.0.0)
```

#### Backward Compatibility

v3.0.0 includes temporary compatibility shim:
- Set `__enable_v2_compat: true` to auto-map old properties
- Compatibility removed in v4.0.0

---

### v1.0.0 → v2.0.0 (Non-Breaking)

**Released:** 2025-11-01

#### New Features
- Added `NewFeature` property for enhanced functionality
- Performance: 30% reduction in memory usage

#### Deprecated Features
- `LegacyMode` property deprecated (no functional changes)

#### Migration
No action required. Existing v1.0.0 configs fully compatible.
```

### Option 2: migrations/ Directory

Use for components with many versions or complex migration paths.

**Structure:**
```
components/[component_name]/
├── component.yaml
├── README.md
└── migrations/
    ├── v1-to-v2.md
    ├── v2-to-v3.md
    └── v3-to-v4.md
```

Each file follows same template as above, focused on one version transition.

## Deprecation Lifecycle

### Stage 1: Announce (Minor Release)

```yaml
# component.yaml
status: alpha  # or stable, depending on current status
version: v2.3.0
```

- Add deprecation notice to README
- If entire component deprecated, update `status` field
- Document replacement in MIGRATIONS.md
- **No functional changes**

### Stage 2: Warn (Next Minor Release)

```yaml
version: v2.4.0
```

- Emit warnings when deprecated features used
- Include removal timeline in warnings
- Continue full functionality

**Example warning implementation:**
```go
if oldPropertyUsed {
    log.Warn("Property 'OldProperty' deprecated since v2.3.0, use 'NewProperty'. Removed in v3.0.0")
}
```

### Stage 3: Remove (Major Release)

```yaml
status: deprecated  # if entire component being replaced
version: v3.0.0
```

- Remove deprecated properties/features
- Update MIGRATIONS.md with "Removed in vX.0.0" marker
- If entire component replaced, archive old component directory

### Timeline Example

| Version | Date       | Action |
|---------|------------|--------|
| v2.3.0  | 2026-01-01 | Deprecate `OldProperty`, document `NewProperty` |
| v2.4.0  | 2026-03-01 | Emit warnings when `OldProperty` used |
| v3.0.0  | 2026-07-01 | Remove `OldProperty` support |

**Recommended timeline:** 6 months between announce and remove for stable components.

## Creating Coexisting Versions

When breaking changes require both versions to coexist:

### 1. Create New Directory

```bash
# Original: components/otelreceiver/
# New:      components/otelreceiver_v2/
mkdir -p components/otelreceiver_v2
```

### 2. Update component.yaml

**Old version:**
```yaml
# components/otelreceiver/component.yaml
kind: OTelReceiver
status: deprecated
version: v1.3.0
```

**New version:**
```yaml
# components/otelreceiver_v2/component.yaml
kind: OTelReceiverV2  # Note: V2 suffix
status: stable
version: v2.0.0
```

### 3. Document Migration

Create `components/otelreceiver/MIGRATIONS.md`:

```markdown
# Migration Guide: OTelReceiver

## OTelReceiver → OTelReceiverV2

**Status:** OTelReceiver deprecated 2026-01-01, removed 2026-07-01
**Replacement:** OTelReceiverV2

### What Changed
- Property `GRPCPort` renamed to `Port`
- Added `Protocol` property (grpc|http)
...
```

### 4. Update README

Both components should have README.md documenting:
- Current status
- Link to migration guide (if deprecated)
- Link to replacement component (if deprecated)

## Migration Documentation Checklist

Every migration entry must include:

- [ ] Version range (e.g., "v2.0.0 → v3.0.0")
- [ ] Release date
- [ ] Deprecation timeline (announce date, removal date)
- [ ] Complete list of breaking changes
- [ ] Impact assessment for each change
- [ ] Rationale for each change
- [ ] Automated migration command (if available)
- [ ] Manual migration steps with before/after examples
- [ ] Deprecation warning examples
- [ ] Backward compatibility notes

## Component Status Field Values

- `development` - Internal only, requires feature flag
- `alpha` - Public preview, expect breaking changes
- `stable` - Production-ready, semver guarantees
- `deprecated` - Scheduled for removal, migration path documented
- `archived` - Removed from active use, read-only reference

## Examples from Real Components

### Example: Non-Breaking Feature Addition

```markdown
### v0.2.0 (2026-01-08)

#### New Features
- Added `MemoryLimit` property to control max memory usage
- Default: no limit (existing behavior preserved)

#### Migration
No action required. Optionally set `MemoryLimit` to enable memory limiting.
```

### Example: Breaking Property Rename

```markdown
### v2.0.0 (2026-01-15)

#### Breaking Changes

**Property Rename:** `SampleRate` → `SamplingRate`

- **Impact:** ~500 configs in production
- **Reason:** Align with OTel semantic conventions
- **Migration:**
  ```bash
  # Automated
  sed -i 's/SampleRate:/SamplingRate:/g' workflow.yaml

  # Or use CLI
  hpsf migrate property --component DeterministicSampler \
    --from SampleRate --to SamplingRate workflow.yaml
  ```
```

### Example: Port Type Change

```markdown
### v2.0.0 (2026-02-01)

#### Breaking Changes

**Port Type Narrowing:** `Output` port from `OTelSignal` → `OTelTraces`

- **Impact:** Components only process traces now
- **Reason:** Split into specialized processors (traces/logs/metrics)
- **Migration:**
  - For traces: Use `ProcessorV2` (this component)
  - For logs: Use `LogProcessorV2`
  - For metrics: Use `MetricProcessorV2`
```

## Questions?

Reach out in #hpsf-support or see [component README](README.md#creating-a-new-component).
