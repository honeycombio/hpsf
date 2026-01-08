#!/bin/bash
# Generates new component from template

COMPONENT_KIND="$1"
COMPONENTS_DIR="pkg/data/components"

if [ -z "$COMPONENT_KIND" ]; then
    echo "Usage: $0 <ComponentKind>"
    echo ""
    echo "Example: $0 MyNewProcessor"
    echo ""
    echo "Creates a new component directory with scaffolded files from templates."
    exit 1
fi

# Prompt for style
echo "Select component style:"
echo "  1) receivers    - Ingest telemetry data"
echo "  2) processors   - Transform, filter, enrich data"
echo "  3) exporters    - Send to destinations"
echo "  4) samplers     - Sampling strategies (Refinery)"
echo "  5) conditions   - Boolean expressions (Refinery)"
echo "  6) startsampling - Start sampling triggers (Refinery)"
read -p "Enter choice [1-6]: " style_choice

case $style_choice in
    1) STYLE="receivers" ;;
    2) STYLE="processors" ;;
    3) STYLE="exporters" ;;
    4) STYLE="samplers" ;;
    5) STYLE="conditions" ;;
    6) STYLE="startsampling" ;;
    *)
        echo "Error: Invalid choice"
        exit 1
        ;;
esac

# Convert PascalCase to snake_case lowercase for directory name
DIR_NAME=$(echo "$COMPONENT_KIND" | sed 's/\([A-Z]\)/_\1/g' | sed 's/^_//' | tr '[:upper:]' '[:lower:]')
TARGET_DIR="$COMPONENTS_DIR/$STYLE/$DIR_NAME"

if [ -d "$TARGET_DIR" ]; then
    echo "Error: Component directory already exists: $TARGET_DIR"
    exit 1
fi

echo ""
echo "Creating new component: $COMPONENT_KIND"
echo "Style: $STYLE"
echo "Directory: $TARGET_DIR"
echo ""

# Create directory
mkdir -p "$TARGET_DIR"

# Determine component section and signal type based on style
COMPONENT_SECTION="$STYLE"
case $STYLE in
    receivers)
        INPUT_TYPE="None"
        OUTPUT_TYPE="OTelTraces"
        SIGNAL_TYPES="traces"
        COLLECTOR_NAME="otlp"
        ;;
    processors)
        INPUT_TYPE="OTelTraces"
        OUTPUT_TYPE="OTelTraces"
        SIGNAL_TYPES="traces"
        COLLECTOR_NAME="transform"
        ;;
    exporters)
        INPUT_TYPE="OTelTraces"
        OUTPUT_TYPE="None"
        SIGNAL_TYPES="traces"
        COLLECTOR_NAME="otlp"
        ;;
    samplers|conditions|startsampling)
        INPUT_TYPE="OTelTraces"
        OUTPUT_TYPE="OTelTraces"
        SIGNAL_TYPES="traces"
        COLLECTOR_NAME=""
        ;;
esac

# Copy and customize component YAML from template (named after directory)
if [ -f "$COMPONENTS_DIR/_templates/component.yaml.tmpl" ]; then
    sed -e "s/{{KIND}}/$COMPONENT_KIND/g" \
        -e "s/{{NAME}}/$COMPONENT_KIND/g" \
        -e "s/{{STYLE}}/${STYLE%s}/g" \
        -e "s/{{LOGO}}/opentelemetry/g" \
        -e "s/{{CATEGORY}}/transformation/g" \
        -e "s/{{SIGNAL}}/traces/g" \
        -e "s/{{INPUT_TYPE}}/$INPUT_TYPE/g" \
        -e "s/{{OUTPUT_TYPE}}/$OUTPUT_TYPE/g" \
        -e "s/{{COMPONENT_SECTION}}/$COMPONENT_SECTION/g" \
        -e "s/{{SIGNAL_TYPES}}/$SIGNAL_TYPES/g" \
        -e "s/{{COLLECTOR_NAME}}/$COLLECTOR_NAME/g" \
        "$COMPONENTS_DIR/_templates/component.yaml.tmpl" > "$TARGET_DIR/${DIR_NAME}.yaml"
    echo "✓ Created ${DIR_NAME}.yaml"
else
    echo "Error: Template not found at $COMPONENTS_DIR/_templates/component.yaml.tmpl"
    rm -rf "$TARGET_DIR"
    exit 1
fi

# Copy and customize README.md from template
if [ -f "$COMPONENTS_DIR/_templates/README.md.tmpl" ]; then
    sed -e "s/{{KIND}}/$COMPONENT_KIND/g" \
        -e "s/{{NAME}}/$COMPONENT_KIND/g" \
        -e "s/{{VERSION}}/v0.0.1/g" \
        -e "s/{{STATUS}}/development/g" \
        -e "s/{{KIND_LOWER}}/$DIR_NAME/g" \
        -e "s/{{INPUT_TYPE}}/$INPUT_TYPE/g" \
        -e "s/{{OUTPUT_TYPE}}/$OUTPUT_TYPE/g" \
        -e "s/{{COMPONENT_SECTION}}/$COMPONENT_SECTION/g" \
        -e "s/{{COLLECTOR_NAME}}/$COLLECTOR_NAME/g" \
        -e "s/{{DATE}}/$(date +%Y-%m-%d)/g" \
        "$COMPONENTS_DIR/_templates/README.md.tmpl" > "$TARGET_DIR/README.md"
    echo "✓ Created README.md"
else
    echo "Error: Template not found at $COMPONENTS_DIR/_templates/README.md.tmpl"
    rm -rf "$TARGET_DIR"
    exit 1
fi

echo ""
echo "=========================================="
echo "Component scaffolded successfully!"
echo ""
echo "Next steps:"
echo "1. Edit $TARGET_DIR/${DIR_NAME}.yaml"
echo "   - Define ports (input/output)"
echo "   - Add properties with validations"
echo "   - Configure templates for target systems"
echo ""
echo "2. Edit $TARGET_DIR/README.md"
echo "   - Add detailed overview and use cases"
echo "   - Document all properties"
echo "   - Provide usage examples"
echo ""
echo "3. Validate:"
echo "   make validate-components"
echo ""
echo "4. Test:"
echo "   go test ./pkg/data/... -v"
echo "=========================================="
