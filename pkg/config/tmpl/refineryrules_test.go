package tmpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetMemberValue(t *testing.T) {
	vsc := &V2SamplerChoice{}
	err := setMemberValue("DeterministicSampler.SampleRate", vsc, 100)
	assert.NoError(t, err)
	assert.Equal(t, 100, vsc.DeterministicSampler.SampleRate)

	err = setMemberValue("RulesBasedSampler.Rules.0.SampleRate", vsc, 200)
	assert.NoError(t, err)
	assert.Equal(t, 200, vsc.RulesBasedSampler.Rules[0].SampleRate)

	err = setMemberValue("RulesBasedSampler.Rules.1.SampleRate", vsc, 300)
	assert.NoError(t, err)
	assert.Equal(t, 300, vsc.RulesBasedSampler.Rules[1].SampleRate)

	err = setMemberValue("RulesBasedSampler.Rules.0.Sampler.DynamicSampler.SampleRate", vsc, int64(400))
	assert.NoError(t, err)
	assert.Equal(t, int64(400), vsc.RulesBasedSampler.Rules[0].Sampler.DynamicSampler.SampleRate)

	err = setMemberValue("WindowedThroughputSampler.UseClusterSize", vsc, true)
	assert.NoError(t, err)
	assert.True(t, vsc.WindowedThroughputSampler.UseClusterSize)

	err = setMemberValue("WindowedThroughputSampler.UpdateFrequency", vsc, "10s")
	assert.NoError(t, err)
	assert.Equal(t, Duration(10*time.Second), vsc.WindowedThroughputSampler.UpdateFrequency)
}

func TestSetMemberValueFields(t *testing.T) {
	vsc := &V2SamplerChoice{}
	err := setMemberValue("RulesBasedSampler.Rules.0.Conditions.0.Fields", vsc, []string{"field1", "field2"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"field1", "field2"}, vsc.RulesBasedSampler.Rules[0].Conditions[0].Fields)

	err = setMemberValue("RulesBasedSampler.Rules.0.Conditions.0.Field", vsc, "field3")
	assert.NoError(t, err)
	assert.Equal(t, "field3", vsc.RulesBasedSampler.Rules[0].Conditions[0].Field)
}

func TestSetMemberValueZeroValue(t *testing.T) {
	vsc := &V2SamplerChoice{}
	err := setMemberValue("RulesBasedSampler.Rules.0.SampleRate", vsc, 1000)
	assert.NoError(t, err)
	assert.Equal(t, 1000, vsc.RulesBasedSampler.Rules[0].SampleRate)
	// Verify that unrelated fields remain nil/uninitialized
	assert.Nil(t, vsc.RulesBasedSampler.Rules[0].Conditions)
}
