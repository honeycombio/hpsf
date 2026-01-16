# Migration Guide: CompareBoolFieldCondition

## CompareBoolFieldCondition → BooleanValueCondition

**Status:** CompareBoolFieldCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [BooleanValueCondition](../boolean_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `CompareBoolFieldCondition` → `BooleanValueCondition`
   - **Impact:** All configs using CompareBoolFieldCondition must update component kind
   - **Reason:** Rename for consistency and clarity

2. **Operator Removed:** No longer supports `Operator` property
   - **Impact:** Operator property removed - equality check is implicit
   - **Reason:** Boolean comparisons only need value check (true/false)

3. **Property Type Change:** `Value` from `bool` to `string` (oneof: "true", "false")
   - **Impact:** Boolean values must be strings ("true"/"false")
   - **Reason:** Standardize configuration format

### Migration Steps

**Manual:**

Before (CompareBoolFieldCondition):
```yaml
components:
  - name: check-sampled
    kind: CompareBoolFieldCondition
    properties:
      - name: Fields
        value: [meta.sampled]
      - name: Operator
        value: '='
      - name: Value
        value: true
```

After (BooleanValueCondition):
```yaml
components:
  - name: check-sampled
    kind: BooleanValueCondition
    properties:
      - name: Fields
        value: [meta.sampled]
      - name: Value
        value: "true"
```

### Operator Migration

| CompareBoolFieldCondition | BooleanValueCondition | Notes |
|--------------------------|----------------------|-------|
| Operator: `=`, Value: `true` | Value: `"true"` | Check field is true |
| Operator: `=`, Value: `false` | Value: `"false"` | Check field is false |
| Operator: `!=`, Value: `true` | Value: `"false"` | Inverted logic |
| Operator: `!=`, Value: `false` | Value: `"true"` | Inverted logic |

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

BooleanValueCondition simplifies boolean checks by removing unnecessary operator property.
