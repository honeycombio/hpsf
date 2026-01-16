# Migration Guide: ErrorExistsCondition

## ErrorExistsCondition → FieldExistenceCondition

**Status:** ErrorExistsCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [FieldExistenceCondition](../field_existence_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `ErrorExistsCondition` → `FieldExistenceCondition`
   - **Impact:** All configs using ErrorExistsCondition must update component kind
   - **Reason:** Generalize to support any field existence checks, not just errors

2. **Property Rename:** `ErrorFields` → `Fields`
   - **Impact:** All configs must rename property
   - **Reason:** Align with generalized field checking purpose

3. **New Property:** `Operator` (default: `exists`)
   - **Impact:** Optional - defaults match ErrorExistsCondition behavior
   - **Reason:** Support inverse checks with `does-not-exist` operator

### Migration Steps

**Manual:**

Before (ErrorExistsCondition):
```yaml
components:
  - name: check-errors
    kind: ErrorExistsCondition
    properties:
      - name: ErrorFields
        value: ["error"]
```

After (FieldExistenceCondition):
```yaml
components:
  - name: check-errors
    kind: FieldExistenceCondition
    properties:
      - name: Fields
        value: ["error"]
      - name: Operator
        value: exists  # optional - this is the default
```

### Multiple Fields Example

Before:
```yaml
properties:
  - name: ErrorFields
    value: ["error", "exception", "error.message"]
```

After:
```yaml
properties:
  - name: Fields
    value: ["error", "exception", "error.message"]
  - name: Operator
    value: exists
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

FieldExistenceCondition adds:
- Support for `does-not-exist` operator for inverse checks
- Smart scope selection based on operator
