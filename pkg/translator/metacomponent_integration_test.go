package translator

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslator_MetaComponent_Integration(t *testing.T) {
	// Create a translator with some basic template components
	translator := NewEmptyTranslator()

	// Install a basic template component for testing
	basicTemplate := config.TemplateComponent{
		Kind:    "TestProcessor",
		Name:    "Test Processor",
		Version: "1.0.0",
		Style:   "processor",
	}
	translator.InstallComponents(map[string]config.TemplateComponent{
		"TestProcessor": basicTemplate,
	})

	// Create an HPSF with a meta component that contains child components
	hpsfDoc := &hpsf.HPSF{
		Components: []*hpsf.Component{
			{
				Name: "TestMeta",
				Kind: "MetaComponent",
				Children: []hpsf.Component{
					{
						Name: "ChildProcessor1",
						Kind: "TestProcessor",
						Properties: []hpsf.Property{
							{Name: "param1", Value: "value1"},
						},
					},
					{
						Name: "ChildProcessor2",
						Kind: "TestProcessor",
						Properties: []hpsf.Property{
							{Name: "param2", Value: "value2"},
						},
					},
				},
			},
		},
	}

	// Test MakeConfigComponent with the meta component
	metaComponent, err := translator.MakeConfigComponent(hpsfDoc.Components[0], "1.0.0")
	require.NoError(t, err)
	require.NotNil(t, metaComponent)

	// Verify it's a MetaComponent
	meta, ok := metaComponent.(*config.MetaComponent)
	require.True(t, ok, "Expected MetaComponent type")
	assert.Equal(t, "TestMeta", meta.Component.Name)
	assert.Len(t, meta.Children, 2)

	// Verify child components are correctly created
	assert.NotNil(t, meta.Children[0])
	assert.NotNil(t, meta.Children[1])
}

func TestTranslator_MetaComponent_EmptyChildren(t *testing.T) {
	translator := NewEmptyTranslator()

	// Create an HPSF with a meta component that has no children (should fail validation)
	hpsfDoc := &hpsf.HPSF{
		Components: []*hpsf.Component{
			{
				Name:     "EmptyMeta",
				Kind:     "MetaComponent",
				Children: []hpsf.Component{}, // Empty children
			},
		},
	}

	// This should fail because meta components must have at least one child
	_, err := translator.MakeConfigComponent(hpsfDoc.Components[0], "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one child component")
}

func TestTranslator_MetaComponent_InvalidChild(t *testing.T) {
	translator := NewEmptyTranslator()
	// Don't install any template components, so child lookup will fail

	hpsfDoc := &hpsf.HPSF{
		Components: []*hpsf.Component{
			{
				Name: "TestMeta",
				Kind: "MetaComponent",
				Children: []hpsf.Component{
					{
						Name: "UnknownChild",
						Kind: "UnknownComponent", // This component type doesn't exist
					},
				},
			},
		},
	}

	// This should fail because the child component type is unknown
	_, err := translator.MakeConfigComponent(hpsfDoc.Components[0], "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build child component")
	assert.Contains(t, err.Error(), "unknown component kind")
}

func TestTranslator_MetaComponent_ConfigGeneration(t *testing.T) {
	translator := NewEmptyTranslator()

	// Install a template component that generates some config
	basicTemplate := config.TemplateComponent{
		Kind:    "TestExporter",
		Name:    "Test Exporter",
		Version: "1.0.0",
		Style:   "exporter",
		Templates: []config.TemplateData{
			{
				Kind:   hpsftypes.CollectorConfig,
				Name:   "collector-config",
				Format: "collector",
				Meta: map[string]any{
					"componentSection":        "exporters",
					"collectorComponentName": "test",
				},
				Data: []any{
					map[string]any{
						"key":   "endpoint",
						"value": "http://example.com",
					},
				},
			},
		},
	}
	translator.InstallComponents(map[string]config.TemplateComponent{
		"TestExporter": basicTemplate,
	})

	hpsfDoc := &hpsf.HPSF{
		Components: []*hpsf.Component{
			{
				Name: "ExporterMeta",
				Kind: "MetaComponent",
				Children: []hpsf.Component{
					{
						Name: "ChildExporter",
						Kind: "TestExporter",
					},
				},
			},
		},
	}

	// Create the meta component
	metaComponent, err := translator.MakeConfigComponent(hpsfDoc.Components[0], "1.0.0")
	require.NoError(t, err)

	meta := metaComponent.(*config.MetaComponent)

	// Test configuration generation for collector config
	pipeline := hpsf.PathWithConnections{ConnType: hpsf.CTYPE_TRACES}
	config, err := meta.GenerateConfig(hpsftypes.CollectorConfig, pipeline, nil)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test configuration generation for refinery rules (may be nil for complex meta components)
	_, err = meta.GenerateConfig(hpsftypes.RefineryRules, pipeline, nil)
	require.NoError(t, err)
	// Note: refinery config may be nil for meta components that don't generate refinery rules
}

func TestTranslator_GetMatchingTemplateComponents_WithMeta(t *testing.T) {
	translator := NewEmptyTranslator()

	// Install a template component
	basicTemplate := config.TemplateComponent{
		Kind:    "TestComponent",
		Name:    "Test Component",
		Version: "1.0.0",
	}
	translator.InstallComponents(map[string]config.TemplateComponent{
		"TestComponent": basicTemplate,
	})

	hpsfDoc := &hpsf.HPSF{
		Components: []*hpsf.Component{
			{
				Name: "RegularComponent",
				Kind: "TestComponent",
			},
			{
				Name: "MetaComponent1",
				Kind: "MetaComponent",
				Children: []hpsf.Component{
					{
						Name: "Child1",
						Kind: "TestComponent",
					},
				},
			},
		},
	}

	// Get matching template components
	templateComps, result := translator.getMatchingTemplateComponents(hpsfDoc)

	// Should not have errors
	require.True(t, result.IsEmpty(), "Expected no errors: %v", result)

	// Should have entries for both components
	assert.Len(t, templateComps, 2)

	// Regular component should have the actual template
	regularComp, exists := templateComps["RegularComponent"]
	require.True(t, exists)
	assert.Equal(t, "TestComponent", regularComp.Kind)

	// Meta component should have a placeholder
	metaComp, exists := templateComps["MetaComponent1"]
	require.True(t, exists)
	assert.Equal(t, "MetaComponent", metaComp.Kind)
	assert.Equal(t, "meta", metaComp.Style)
}