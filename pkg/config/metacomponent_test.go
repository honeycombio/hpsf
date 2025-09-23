package config

import (
	"testing"

	"github.com/honeycombio/hpsf/pkg/config/tmpl"
	"github.com/honeycombio/hpsf/pkg/hpsf"
	"github.com/honeycombio/hpsf/pkg/hpsftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaComponent_Creation(t *testing.T) {
	// Create a basic hpsf.Component for testing
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}

	// Create a new MetaComponent
	meta := NewMetaComponent(hpsfComponent)

	// Verify initial state
	assert.NotNil(t, meta)
	assert.Equal(t, "TestMeta", meta.Component.Name)
	assert.Equal(t, "MetaComponent", meta.Component.Kind)
	assert.Len(t, meta.Children, 0)
	assert.Len(t, meta.InternalConns, 0)
	assert.Len(t, meta.ExternalConns, 0)
}

func TestMetaComponent_AddChild(t *testing.T) {
	// Create MetaComponent
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}
	meta := NewMetaComponent(hpsfComponent)

	// Create a child component
	childComponent := hpsf.Component{
		Name: "TestChild",
		Kind: "TestComponent",
	}
	child := &GenericBaseComponent{Component: childComponent}

	// Add child
	meta.AddChild(child)

	// Verify child was added
	assert.Len(t, meta.Children, 1)
	assert.Equal(t, child, meta.Children[0])
}

func TestMetaComponent_GenerateConfig_RefineryConfig(t *testing.T) {
	// Create MetaComponent
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}
	meta := NewMetaComponent(hpsfComponent)

	// Generate config without children (should return base config)
	config, err := meta.GenerateConfig(hpsftypes.RefineryConfig, hpsf.PathWithConnections{}, nil)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify it's a DottedConfig with expected base values
	dottedConfig, ok := config.(tmpl.DottedConfig)
	require.True(t, ok)
	assert.Equal(t, 2, dottedConfig["General.ConfigurationVersion"])
	assert.Equal(t, "v2.0", dottedConfig["General.MinRefineryVersion"])
}

func TestMetaComponent_GenerateConfig_CollectorConfig(t *testing.T) {
	// Create MetaComponent
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}
	meta := NewMetaComponent(hpsfComponent)

	// Generate config without children (should return base collector config)
	config, err := meta.GenerateConfig(hpsftypes.CollectorConfig, hpsf.PathWithConnections{}, nil)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify it's a CollectorConfig
	collectorConfig, ok := config.(*tmpl.CollectorConfig)
	require.True(t, ok)
	assert.NotNil(t, collectorConfig.Sections)
}

func TestMetaComponent_GenerateConfig_RefineryRules(t *testing.T) {
	// Create MetaComponent
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}
	meta := NewMetaComponent(hpsfComponent)

	// Generate config without children (should return base rules config)
	config, err := meta.GenerateConfig(hpsftypes.RefineryRules, hpsf.PathWithConnections{}, nil)

	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify it's a RulesConfig
	rulesConfig, ok := config.(*tmpl.RulesConfig)
	require.True(t, ok)
	assert.NotNil(t, rulesConfig)
}

func TestMetaComponent_Connections(t *testing.T) {
	// Create MetaComponent
	hpsfComponent := hpsf.Component{
		Name: "TestMeta",
		Kind: "MetaComponent",
	}
	meta := NewMetaComponent(hpsfComponent)

	// Create test connections
	externalConn := &hpsf.Connection{
		Source:      hpsf.ConnectionPort{Component: "external"},
		Destination: hpsf.ConnectionPort{Component: "TestMeta"},
	}

	internalConn := &hpsf.Connection{
		Source:      hpsf.ConnectionPort{Component: "child1"},
		Destination: hpsf.ConnectionPort{Component: "child2"},
	}

	// Add connections
	meta.AddConnection(externalConn)
	meta.AddInternalConnection(internalConn)

	// Verify connections were added
	assert.Len(t, meta.ExternalConns, 1)
	assert.Len(t, meta.InternalConns, 1)
	assert.Equal(t, externalConn, meta.ExternalConns[0])
	assert.Equal(t, internalConn, meta.InternalConns[0])
}