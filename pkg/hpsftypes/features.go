package hpsftypes

import "strings"

// Feature represents a client capability that affects config generation behavior.
// Features are used to maintain backward compatibility with older clients while
// enabling new functionality for updated clients.
type Feature string

// Future features will be added here as new capabilities are developed
// For example:
// const (
//     FeatureEnvIDSubstitution Feature = "env_id_substitution"
// )

// Features represents a collection of client capabilities.
type Features []Feature

// Contains checks if a specific feature is present in the collection.
func (f Features) Contains(feature Feature) bool {
	for _, existing := range f {
		if existing == feature {
			return true
		}
	}
	return false
}

// Strings converts Features to a slice of strings.
// This is useful for serialization or logging.
func (f Features) Strings() []string {
	strs := make([]string, len(f))
	for i, feature := range f {
		strs[i] = string(feature)
	}
	return strs
}

// ParseFeaturesFromString parses a comma-separated string of features.
// Whitespace is trimmed from each feature name. Empty strings and empty
// features are ignored. This is useful for parsing query parameters.
//
// Example:
//   ParseFeaturesFromString("feature_one, feature_two")
//   // Returns: Features{Feature("feature_one"), Feature("feature_two")}
func ParseFeaturesFromString(s string) Features {
	if s == "" {
		return Features{}
	}
	strs := strings.Split(s, ",")
	features := make(Features, 0, len(strs))
	for _, str := range strs {
		trimmed := strings.TrimSpace(str)
		if trimmed != "" {
			features = append(features, Feature(trimmed))
		}
	}
	return features
}

// ParseFeatures converts a slice of strings into Features.
// This is useful when you already have a slice of feature strings.
// Unknown feature strings are kept as-is for forward compatibility.
func ParseFeatures(strs []string) Features {
	features := make(Features, 0, len(strs))
	for _, s := range strs {
		features = append(features, Feature(s))
	}
	return features
}
