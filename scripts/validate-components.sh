#!/bin/bash
# Validates component directory structure

set -e

COMPONENTS_DIR="pkg/data/components"
ERRORS=0
WARNINGS=0

echo "=== HPSF Component Structure Validation ==="
echo ""

# Check if components directory exists
if [ ! -d "$COMPONENTS_DIR" ]; then
    echo "ERROR: Components directory not found: $COMPONENTS_DIR"
    exit 1
fi

# Check for flat YAML files (old structure)
flat_yamls=$(find "$COMPONENTS_DIR" -maxdepth 1 -name "*.yaml" 2>/dev/null | wc -l | tr -d ' ')
if [ "$flat_yamls" -gt 0 ]; then
    echo "ERROR: Found $flat_yamls flat YAML file(s) in $COMPONENTS_DIR"
    echo "       All components must be in subdirectories with component.yaml"
    ls "$COMPONENTS_DIR"/*.yaml 2>/dev/null | sed 's/^/       - /'
    ERRORS=$((ERRORS + 1))
    echo ""
fi

# Find all component directories (2-level structure: style/component)
component_dirs=()
for level1_dir in "$COMPONENTS_DIR"/*/; do
    level1_name=$(basename "$level1_dir")

    # Skip special directories
    if [[ "$level1_name" == _* ]] || [[ "$level1_name" == .* ]]; then
        continue
    fi

    if [ -d "$level1_dir" ]; then
        # Level 2: component directories
        for level2_dir in "$level1_dir"/*/; do
            if [ -d "$level2_dir" ]; then
                comp_name=$(basename "$level2_dir")
                if [ -f "$level2_dir/${comp_name}.yaml" ]; then
                    component_dirs+=("$level2_dir")
                fi
            fi
        done
    fi
done

if [ ${#component_dirs[@]} -eq 0 ]; then
    echo "ERROR: No component directories found in $COMPONENTS_DIR"
    exit 1
fi

echo "Found ${#component_dirs[@]} component directories"
echo ""

# Validate each component directory
for dir in "${component_dirs[@]}"; do
    dir_name=$(basename "$dir")

    echo "Validating: $dir_name"

    # Check component YAML exists (should match directory name)
    component_yaml="$dir/${dir_name}.yaml"
    if [ ! -f "$component_yaml" ]; then
        echo "  ✗ ERROR: Missing ${dir_name}.yaml"
        ERRORS=$((ERRORS + 1))
        continue
    else
        echo "  ✓ ${dir_name}.yaml exists"
    fi

    # Validate YAML syntax using yq
    if command -v yq &> /dev/null; then
        if ! yq eval '.' "$component_yaml" > /dev/null 2>&1; then
            echo "  ✗ ERROR: Invalid YAML syntax in ${dir_name}.yaml"
            ERRORS=$((ERRORS + 1))
        else
            echo "  ✓ Valid YAML syntax"
        fi
    else
        echo "  ✗ ERROR: yq not found (required for YAML validation)"
        ERRORS=$((ERRORS + 1))
    fi

    # Extract and validate kind field
    kind=$(grep "^kind:" "$component_yaml" 2>/dev/null | head -1 | awk '{print $2}' | tr -d '\r\n')
    if [ -z "$kind" ]; then
        echo "  ✗ ERROR: No 'kind' field found in component.yaml"
        ERRORS=$((ERRORS + 1))
    else
        echo "  ✓ Kind: $kind"

        # Check if directory name matches expected convention
        expected_dir=$(echo "$kind" | sed 's/\([A-Z]\)/_\1/g' | sed 's/^_//' | tr '[:upper:]' '[:lower:]')
        if [ "$dir_name" != "$expected_dir" ]; then
            echo "  ! WARN: Directory name '$dir_name' doesn't match kind '$kind' (expected '$expected_dir')"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi

    # Check README.md exists
    if [ ! -f "$dir/README.md" ]; then
        echo "  ! WARN: Missing README.md"
        WARNINGS=$((WARNINGS + 1))
    else
        echo "  ✓ README.md exists"
    fi

    # Check for MIGRATIONS.md if deprecated/archived
    status=$(grep "^status:" "$component_yaml" 2>/dev/null | head -1 | awk '{print $2}' | tr -d '\r\n')
    if [[ "$status" == "deprecated" || "$status" == "archived" ]]; then
        if [ ! -f "$dir/MIGRATIONS.md" ] && [ ! -d "$dir/migrations" ]; then
            echo "  ✗ ERROR: Status is '$status' but no MIGRATIONS.md or migrations/ directory found"
            ERRORS=$((ERRORS + 1))
        else
            echo "  ✓ Migration documentation exists"
        fi
    fi

    # Check for required fields
    for field in "name" "version" "status" "summary"; do
        if ! grep -q "^$field:" "$component_yaml" 2>/dev/null; then
            echo "  ! WARN: Missing required field '$field'"
            WARNINGS=$((WARNINGS + 1))
        fi
    done

    echo ""
done

echo "=========================================="
echo "Validation complete!"
echo ""
echo "Components validated: ${#component_dirs[@]}"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"
echo ""

if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo "✓ All validations passed!"
    else
        echo "⚠ Validation passed with warnings"
    fi
    exit 0
else
    echo "✗ Validation failed with $ERRORS error(s)"
    exit 1
fi
