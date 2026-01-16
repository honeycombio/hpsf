# Migration Guide: FieldExistsCondition

## FieldExistsCondition → FieldExistenceCondition

**Status:** FieldExistsCondition deprecated 2026-01-15
**Released:** 2026-01-15
**Replacement:** [FieldExistenceCondition](../field_existence_condition/README.md)

### Breaking Changes

1. **Component Kind Change:** `FieldExistsCondition` → `FieldExistenceCondition`
   - **Impact:** All configs using FieldExistsCondition must update component kind
   - **Reason:** Consistent naming with other conditions

2. **Property Change:** `Field` (string) → `Fields` (array)
   - **Impact:** Single field must be wrapped in array
   - **Reason:** Support checking multiple fields with same condition

### Migration Steps

**Manual:**

Before (FieldExistsCondition):
```yaml
components:
  - name: check-field
    kind: FieldExistsCondition
    properties:
      - name: Field
        value: error.message
      - name: Operator
        value: exists
```

After (FieldExistenceCondition):
```yaml
components:
  - name: check-field
    kind: FieldExistenceCondition
    properties:
      - name: Fields
        value: [error.message]
      - name: Operator
        value: exists
```

### Multiple Fields Example

FieldExistenceCondition supports checking multiple fields:

```yaml
components:
  - name: check-error-fields
    kind: FieldExistenceCondition
    properties:
      - name: Fields
        value: [error, exception, error.message]
      - name: Operator
        value: exists
```

### Backward Compatibility

None. This is a complete replacement requiring config updates.

### Additional Capabilities

FieldExistenceCondition adds:
- Support for multiple fields in single condition
- Same operators: `exists`, `does-not-exist`
- Same smart scope selection based on operator
