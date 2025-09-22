package tmpl

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	y "gopkg.in/yaml.v3"
)

// DottedConfig is a map that allows for keys with dots in them; it can convert
// a regular map into a DottedConfig, and when rendered, it will generate nested
// maps. This exists because dotted paths are easier to merge.
// There's a special case we have to deal with where if there are duplicate keys
// in the list of dotted configs, we want to create a list of them at the level
// above the final key.
// For example, if we have:
//
//	a.b.c: 1
//	a.b.d: 2
//	a.b.e: 3
//	a.b.c: 4
//	a.b.d: 5
//	a.b.e: 6
//
// We want to end up with:
//
//	a:
//	  b:
//	   - c: 1
//	     d: 2
//	     e: 3
//	   - c: 4
//	     d: 5
//	     e: 6
type DottedConfig map[string]any

// renderInto is a helper function that recursively renders a DottedConfig into a map.
func (dc DottedConfig) renderInto(m map[string]any, key string, value any) {
	// if the key contains a dot, split it into parts
	if strings.Contains(key, ".") {
		// split the key into parts
		parts := strings.SplitN(key, ".", 2)
		// if the first part of the key does not exist in the map, create it
		if m[parts[0]] == nil {
			m[parts[0]] = make(map[string]any)
		}
		switch m[parts[0]].(type) {
		case []map[string]any:
			// if the first part of the key is a list of maps, append to it
			// we need to create a new map for the new value
			newMap := make(map[string]any)
			dc.renderInto(newMap, parts[1], value)
			m[parts[0]] = append(m[parts[0]].([]map[string]any), newMap)
		case map[string]any:
			// if the first part of the key is a map, we need to check if the
			// second part of the key already exists in the map
			if _, ok := m[parts[0]].(map[string]any)[parts[1]]; ok {
				// if it does, we need to create a new map for the new value
				newMap := make(map[string]any)
				dc.renderInto(newMap, parts[1], value)
				// and turn the existing map into a list of maps
				m[parts[0]] = append([]map[string]any{m[parts[0]].(map[string]any)}, newMap)
			} else {
				// if it doesn't, we can just call renderInto on the existing map
				dc.renderInto(m[parts[0]].(map[string]any), parts[1], value)
			}
		default:
			log.Printf("Template error in DottedConfig.renderInto: %s is not a map", parts[0])
		}
	} else {
		// if the key does not contain a dot, assign the value
		m[key] = value
	}
}

// Iterate through the map recursively. If at any level, the key ends with a
// number in square brackets (which indicates that it's an indexed value in a
// slice), then we need to take the value of that key, determine its type T, and put it into a
// []T at the same level, but with the new key being the portion of
// the name before the `[` and `]`. The number in the brackets is the index of
// the slice.
func processIndices(in map[string]any) map[string]any {
	pat := regexp.MustCompile(`^(.*)\[(\d+)\]$`)
	out := make(map[string]any)
	for k, v := range in {
		hasIndex := pat.MatchString(k)
		switch v := v.(type) {
		case map[string]any:
			// if the value is a map, we need to recursively call processIndices on it first
			cv := processIndices(v)
			if !hasIndex {
				// if the key doesn't match our regex, just add it to the map
				out[k] = cv
			} else {
				// we need to process it -- split the key and index
				matches := pat.FindStringSubmatch(k)
				key := matches[1]
				index, _ := strconv.Atoi(matches[2])

				// maybe we have a slice already
				sl, ok := out[key].([]map[string]any)
				if !ok {
					sl = make([]map[string]any, 0)
				}
				// maybe expand the slice to fit the index
				for i := len(sl); i <= index; i++ {
					sl = append(sl, make(map[string]any))
				}
				// replace the value at the list at the index (it will be a map)
				sl[index] = cv
				out[key] = sl
			}
		default:
			if hasIndex {
				// we have a key with an index, but the value is not a map, so we need to either append
				// it to a slice or create a new slice with the value
				matches := pat.FindStringSubmatch(k)
				key := matches[1]
				index, _ := strconv.Atoi(matches[2])
				// maybe we have a slice already
				sl, ok := out[key].([]any)
				if !ok {
					sl = make([]any, 0)
				}
				// maybe expand the slice to fit the index
				for i := len(sl); i <= index; i++ {
					sl = append(sl, nil) // fill with nils to expand
				}
				// replace the value at the list at the index
				sl[index] = v
				out[key] = sl
			} else {
				// if the key doesn't match our regex, we can just add it to the map
				out[k] = v
			}
		}
	}
	return out
}

// RenderToMap renders the config into a map.
func (dc DottedConfig) RenderToMap(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	for k, v := range dc {
		dc.renderInto(m, k, v)
	}
	cm := processIndices(m)
	return cm
}

// RenderYAML renders the config into YAML and returns a hash of it.
func (dc DottedConfig) RenderYAML() ([]byte, error) {
	m := dc.RenderToMap(nil)
	data, err := y.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var indexPattern = regexp.MustCompile(`\.([^.]+)\[(\d+)\]`)

func findIndexedValue(s string) (string, int, bool) {
	// search for stuff like `a.b.foo[0]` and `a.b.bar[1]`,
	// and return a tuple like `("foo", 0, true)` or `("bar", 1, true)`
	// we only want the first match
	matches := indexPattern.FindStringSubmatch(s)
	if len(matches) < 3 {
		return "", 0, false
	}
	key := matches[1]
	index, _ := strconv.Atoi(matches[2])
	return key, index, true
}

func (dc DottedConfig) FindIndexedValues() map[string]int {
	indices := make(map[string]int)
	// search for stuff like `a.b.foo[0]` and `a.b.bar[1]`
	// and return a map like `{"foo": 0, "bar": 1}`
	for k := range dc {
		key, index, ok := findIndexedValue(k)
		if !ok {
			// if we don't have a match, continue
			continue
		}
		// if the key is not in the map, it will return 0
		// if it exists, we only want the largest index
		if index >= indices[key] {
			indices[key] = index
		}
	}
	return indices
}

// Merge combines two `DottedConfig` structs together; the values from the
// `DottedConfig` passed in will override any values that are not slices.
func (dc DottedConfig) Merge(other TemplateConfig) error {
	otherDotted, ok := other.(DottedConfig)
	if !ok {
		// if the other TemplateConfig is not a DottedConfig, we can't merge it
		return fmt.Errorf("cannot merge %T with DottedConfig", other)
	}
	baseIndices := dc.FindIndexedValues()
	for k, v := range otherDotted {
		// let's check if we need to adjust the value based on indices
		otherKey, otherIndex, ok := findIndexedValue(k)
		if ok {
			// if we have an indexed value, look if we have a base index for it
			if baseIndex, ok := baseIndices[otherKey]; ok {
				// if we have a base index, we need to adjust the index
				otherIndex += baseIndex + 1
				splits := indexPattern.Split(k, 2)
				// put the new key back together
				k = fmt.Sprintf("%s.%s[%d]%s", splits[0], otherKey, otherIndex, splits[1])
			}
		}

		if _, ok := dc[k]; !ok {
			// if the key doesn't exist, we can just add it
			dc[k] = v
			continue
		}

		// if the key exists, we need to check the type of the value
		// and append it to the existing value if it's a slice, or overwrite it if
		// it's not a slice
		switch v := v.(type) {
		case []any:
			dc[k] = append(dc[k].([]any), v...)
		case []string:
			dc[k] = append(dc[k].([]string), v...)
		case []int:
			dc[k] = append(dc[k].([]int), v...)
		case []float64:
			dc[k] = append(dc[k].([]float64), v...)
		default:
			dc[k] = v
		}
	}
	return nil
}

// NewDottedConfig recursively converts a map into a DottedConfig.
func NewDottedConfig(m map[string]any) DottedConfig {
	dc := DottedConfig{}
	for k, v := range m {
		switch v := v.(type) {
		case map[string]any:
			for kk, vv := range NewDottedConfig(v) {
				dc[k+"."+kk] = vv
			}
		default:
			dc[k] = v
		}
	}
	return dc
}
