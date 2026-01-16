# Migration Guide: CompareIntegerFieldCondition

## CompareIntegerFieldCondition → NumericValueCondition

**Status:** CompareIntegerFieldCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [NumericValueCondition](../numeric_value_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `CompareIntegerFieldCondition` → `NumericValueCondition`
   - **Impact:** All configs using CompareIntegerFieldCondition must update component kind
   - **Reason:** Consolidate numeric comparisons (int and float) into single component

2. **Property Type Change:** `Value` from `int` to `float`
   - **Impact:** Integer values still work, now supports decimal values
   - **Reason:** Support both integer and float comparisons

3. **New Property:** `ForceIntegerCompare` (default: `false`)
   - **Impact:** Optional - set to `true` to force integer-only comparison (original behavior)
   - **Reason:** Preserve strict integer comparison when needed

### Migration Steps

**Manual:**

Before (CompareIntegerFieldCondition):
```yaml
components:
  - name: check-status
    kind: CompareIntegerFieldCondition
    properties:
      - name: Fields
        value: [http.status_code]
      - name: Operator
        value: '>='
      - name: Value
        value: 500
```

After (NumericValueCondition):
```yaml
components:
  - name: check-status
    kind: NumericValueCondition
    properties:
      - name: Fields
        value: [http.status_code]
      - name: Operator
        value: '>='
      - name: Value
        value: 500
      - name: ForceIntegerCompare
        value: true  # optional - preserves strict integer behavior
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

NumericValueCondition adds:
- Support for float/decimal comparisons
- Optional strict integer comparison mode via `ForceIntegerCompare`
- Same operators: `=`, `!=`, `>`, `>=`, `<`, `<=`
