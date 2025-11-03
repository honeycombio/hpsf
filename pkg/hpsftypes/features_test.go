package hpsftypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test-only features for testing the feature infrastructure
const (
	testFeature1 Feature = "test_feature_1"
	testFeature2 Feature = "test_feature_2"
	testFeature3 Feature = "test_feature_3"
)

func TestParseFeaturesFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Features
	}{
		{
			name:     "single feature",
			input:    "test_feature_1",
			expected: Features{testFeature1},
		},
		{
			name:     "multiple features",
			input:    "test_feature_1,test_feature_2",
			expected: Features{testFeature1, testFeature2},
		},
		{
			name:     "features with spaces",
			input:    "test_feature_1, test_feature_2",
			expected: Features{testFeature1, testFeature2},
		},
		{
			name:     "empty string",
			input:    "",
			expected: Features{},
		},
		{
			name:     "whitespace only",
			input:    "  ,  , ",
			expected: Features{},
		},
		{
			name:     "mixed with empty",
			input:    "test_feature_1,,test_feature_2",
			expected: Features{testFeature1, testFeature2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFeaturesFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseFeatures(t *testing.T) {
	input := []string{"test_feature_1", "test_feature_2"}
	expected := Features{testFeature1, testFeature2}

	result := ParseFeatures(input)
	assert.Equal(t, expected, result)
}

func TestFeatures_Strings(t *testing.T) {
	features := Features{testFeature1, testFeature2}
	expected := []string{"test_feature_1", "test_feature_2"}

	result := features.Strings()
	assert.Equal(t, expected, result)
}

func TestFeatures_Contains(t *testing.T) {
	features := Features{testFeature1, testFeature2}

	assert.True(t, features.Contains(testFeature1))
	assert.True(t, features.Contains(testFeature2))
	assert.False(t, features.Contains(testFeature3))

	// Empty list
	emptyFeatures := Features{}
	assert.False(t, emptyFeatures.Contains(testFeature1))
}
