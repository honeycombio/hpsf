# Migration Guide: CompareDecimalFieldCondition

## CompareDecimalFieldCondition → NumericValueCondition

**Status:** CompareDecimalFieldCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [NumericValueCondition](../numeric_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `CompareDecimalFieldCondition` → `NumericValueCondition`
   - **Impact:** All configs using CompareDecimalFieldCondition must update component kind
   - **Reason:** Consolidate numeric comparisons (int and float) into single component

2. **New Property:** `ForceIntegerCompare` (default: `false`)
   - **Impact:** Optional - leave as default `false` for float comparison (original behavior)
   - **Reason:** Support both integer and float comparisons in one component

### Migration Steps

**Manual:**

Before (CompareDecimalFieldCondition):
```yaml
components:
  - name: check-duration
    kind: CompareDecimalFieldCondition
    properties:
      - name: Fields
        value: [duration_ms]
      - name: Operator
        value: '>'
      - name: Value
        value: 100.5
```

After (NumericValueCondition):
```yaml
components:
  - name: check-duration
    kind: NumericValueCondition
    properties:
      - name: Fields
        value: [duration_ms]
      - name: Operator
        value: '>'
      - name: Value
        value: 100.5
      # ForceIntegerCompare: false is the default (float comparison)
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

NumericValueCondition adds:
- Support for both integer and float comparisons
- `ForceIntegerCompare` property for strict integer mode
- Same operators: `=`, `!=`, `>`, `>=`, `<`, `<=`
