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

# Find all component directories (3-level structure: target/style/component)
component_dirs=()
for level1_dir in "$COMPONENTS_DIR"/*/; do
    level1_name=$(basename "$level1_dir")

    # Skip special directories
    if [[ "$level1_name" == _* ]] || [[ "$level1_name" == .* ]]; then
        continue
    fi

    if [ -d "$level1_dir" ]; then
        # Level 2: style directories
        for level2_dir in "$level1_dir"/*/; do
            if [ ! -d "$level2_dir" ]; then
                continue
            fi

            # Level 3: component directories
            for level3_dir in "$level2_dir"/*/; do
                if [ -d "$level3_dir" ] && [ -f "$level3_dir/component.yaml" ]; then
                    component_dirs+=("$level3_dir")
                fi
            done
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

    # Check component.yaml exists
    if [ ! -f "$dir/component.yaml" ]; then
        echo "  ✗ ERROR: Missing component.yaml"
        ERRORS=$((ERRORS + 1))
        continue
    else
        echo "  ✓ component.yaml exists"
    fi

    # Validate YAML syntax (using Python if available)
    if command -v python3 &> /dev/null; then
        if python3 -c "import yaml" 2>/dev/null; then
            if ! python3 -c "import yaml; yaml.safe_load(open('$dir/component.yaml'))" 2>/dev/null; then
                echo "  ✗ ERROR: Invalid YAML syntax in component.yaml"
                ERRORS=$((ERRORS + 1))
            else
                echo "  ✓ Valid YAML syntax"
            fi
        else
            echo "  - Skipping YAML validation (pyyaml module not installed)"
        fi
    else
        echo "  - Skipping YAML validation (python3 not available)"
    fi

    # Extract and validate kind field
    kind=$(grep "^kind:" "$dir/component.yaml" 2>/dev/null | head -1 | awk '{print $2}' | tr -d '\r\n')
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
        ((WARNINGS++))
    else
        echo "  ✓ README.md exists"
    fi

    # Check for MIGRATIONS.md if deprecated/archived
    status=$(grep "^status:" "$dir/component.yaml" 2>/dev/null | head -1 | awk '{print $2}' | tr -d '\r\n')
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
        if ! grep -q "^$field:" "$dir/component.yaml" 2>/dev/null; then
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
