// These types are largely copied with edits from Refinery's config package. We
// can't import them directly because Refinery wasn't designed to be imported as
// a library, and it both has many dependencies and also doesn't do the Go
// module thing correctly, so we actually can't import it at all in its current
// form (it tries to import the last v1 version, which is very old).
package tmpl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// In order to be able to unmarshal "15s" etc. into time.Duration, we need to
// define a new type and implement MarshalText and UnmarshalText.
type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

type DeterministicSamplerConfig struct {
	SampleRate int `yaml:"SampleRate,omitempty"`
}

type DynamicSamplerConfig struct {
	SampleRate     int64    `yaml:"SampleRate,omitempty"`
	ClearFrequency Duration `yaml:"ClearFrequency,omitempty"`
	FieldList      []string `yaml:"FieldList,omitempty"`
	MaxKeys        int      `yaml:"MaxKeys,omitempty"`
	UseTraceLength bool     `yaml:"UseTraceLength,omitempty"`
}

type EMADynamicSamplerConfig struct {
	GoalSampleRate      int      `yaml:"GoalSampleRate,omitempty"`
	AdjustmentInterval  Duration `yaml:"AdjustmentInterval,omitempty"`
	Weight              float64  `yaml:"Weight,omitempty"`
	AgeOutValue         float64  `yaml:"AgeOutValue,omitempty"`
	BurstMultiple       float64  `yaml:"BurstMultiple,omitempty"`
	BurstDetectionDelay uint     `yaml:"BurstDetectionDelay,omitempty"`
	FieldList           []string `yaml:"FieldList,omitempty"`
	MaxKeys             int      `yaml:"MaxKeys,omitempty"`
	UseTraceLength      bool     `yaml:"UseTraceLength,omitempty"`
}

type EMAThroughputSamplerConfig struct {
	GoalThroughputPerSec int      `yaml:"GoalThroughputPerSec,omitempty"`
	UseClusterSize       bool     `yaml:"UseClusterSize,omitempty"`
	InitialSampleRate    int      `yaml:"InitialSampleRate,omitempty"`
	AdjustmentInterval   Duration `yaml:"AdjustmentInterval,omitempty"`
	Weight               float64  `yaml:"Weight,omitempty"`
	AgeOutValue          float64  `yaml:"AgeOutValue,omitempty"`
	BurstMultiple        float64  `yaml:"BurstMultiple,omitempty"`
	BurstDetectionDelay  uint     `yaml:"BurstDetectionDelay,omitempty"`
	FieldList            []string `yaml:"FieldList,omitempty"`
	MaxKeys              int      `yaml:"MaxKeys,omitempty"`
	UseTraceLength       bool     `yaml:"UseTraceLength,omitempty"`
}

type WindowedThroughputSamplerConfig struct {
	UpdateFrequency      Duration `yaml:"UpdateFrequency,omitempty"`
	LookbackFrequency    Duration `yaml:"LookbackFrequency,omitempty"`
	GoalThroughputPerSec int      `yaml:"GoalThroughputPerSec,omitempty"`
	UseClusterSize       bool     `yaml:"UseClusterSize,omitempty"`
	FieldList            []string `yaml:"FieldList,omitempty"`
	MaxKeys              int      `yaml:"MaxKeys,omitempty"`
	UseTraceLength       bool     `yaml:"UseTraceLength,omitempty"`
}

type TotalThroughputSamplerConfig struct {
	GoalThroughputPerSec int      `yaml:"GoalThroughputPerSec,omitempty"`
	UseClusterSize       bool     `yaml:"UseClusterSize,omitempty"`
	ClearFrequency       Duration `yaml:"ClearFrequency,omitempty"`
	FieldList            []string `yaml:"FieldList,omitempty"`
	MaxKeys              int      `yaml:"MaxKeys,omitempty"`
	UseTraceLength       bool     `yaml:"UseTraceLength,omitempty"`
}

type RulesBasedDownstreamSampler struct {
	DynamicSampler            *DynamicSamplerConfig            `yaml:"DynamicSampler,omitempty"`
	EMADynamicSampler         *EMADynamicSamplerConfig         `yaml:"EMADynamicSampler,omitempty"`
	EMAThroughputSampler      *EMAThroughputSamplerConfig      `yaml:"EMAThroughputSampler,omitempty"`
	WindowedThroughputSampler *WindowedThroughputSamplerConfig `yaml:"WindowedThroughputSampler,omitempty"`
	TotalThroughputSampler    *TotalThroughputSamplerConfig    `yaml:"TotalThroughputSampler,omitempty"`
	DeterministicSampler      *DeterministicSamplerConfig      `yaml:"DeterministicSampler,omitempty"`
}

type RulesBasedSamplerRule struct {
	Name       string                        `yaml:"Name,omitempty"`
	SampleRate int                           `yaml:"SampleRate,omitempty"`
	Drop       bool                          `yaml:"Drop,omitempty"`
	Scope      string                        `yaml:"Scope,omitempty"`
	Conditions []*RulesBasedSamplerCondition `yaml:"Conditions,omitempty"`
	Sampler    *RulesBasedDownstreamSampler  `yaml:"Sampler,omitempty"`
}

type RulesBasedSamplerCondition struct {
	Field    string                            `yaml:"Field,omitempty"`
	Fields   []string                          `yaml:"Fields,omitempty"`
	Operator string                            `yaml:"Operator,omitempty"`
	Value    any                               `yaml:"Value,omitempty"`
	Datatype string                            `yaml:"Datatype,omitempty"`
	Matches  func(value any, exists bool) bool `yaml:"-"`
}

type RulesBasedSamplerConfig struct {
	Rules             []*RulesBasedSamplerRule `yaml:"Rules,omitempty"`
	CheckNestedFields bool                     `yaml:"CheckNestedFields,omitempty"`
}

type V2SamplerChoice struct {
	DeterministicSampler      *DeterministicSamplerConfig      `yaml:"DeterministicSampler,omitempty"`
	RulesBasedSampler         *RulesBasedSamplerConfig         `yaml:"RulesBasedSampler,omitempty"`
	DynamicSampler            *DynamicSamplerConfig            `yaml:"DynamicSampler,omitempty"`
	EMADynamicSampler         *EMADynamicSamplerConfig         `yaml:"EMADynamicSampler,omitempty"`
	EMAThroughputSampler      *EMAThroughputSamplerConfig      `yaml:"EMAThroughputSampler,omitempty"`
	WindowedThroughputSampler *WindowedThroughputSamplerConfig `yaml:"WindowedThroughputSampler,omitempty"`
	TotalThroughputSampler    *TotalThroughputSamplerConfig    `yaml:"TotalThroughputSampler,omitempty"`
}

// setSliceElementValue handles setting a value on a slice element, including extending the slice,
// initializing pointer elements, and recursing into the element if needed.
// This is a helper function for SetMemberValue.
func setSliceElementValue(field reflect.Value, index int, subKey string, value any) error {
	// append enough elements to the slice to get to the index
	if index >= field.Len() {
		for i := field.Len(); i <= index; i++ {
			field.Set(reflect.Append(field, reflect.Zero(field.Type().Elem())))
		}
	}
	elem := field.Index(index)
	// Always initialize pointer elements if nil before recursing
	if elem.Kind() == reflect.Ptr {
		if elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct in slice, got %s", elem.Kind())
	}

	nextMember := field.Index(index).Interface()
	if field.Index(index).CanAddr() {
		nextMember = field.Index(index).Addr().Interface()
	}

	if subKey != "" {
		return setMemberValue(subKey, nextMember, value)
	} else {
		if field.Index(index).Kind() == reflect.Ptr && field.Index(index).IsNil() {
			field.Index(index).Set(reflect.New(field.Index(index).Type().Elem()))
		}
		// At the end of the path, set the value with type conversion if needed
		v := reflect.ValueOf(value)
		if v.Type() != field.Index(index).Elem().Type() {
			if v.Type().ConvertibleTo(field.Index(index).Elem().Type()) {
				v = v.Convert(field.Index(index).Elem().Type())
			} else {
				return fmt.Errorf("cannot assign value of type %s to field of type %s", v.Type(), field.Index(index).Elem().Type())
			}
		}
		field.Index(index).Elem().Set(v)
		return nil
	}
}

// setMemberValue sets a value on a member of a struct, including nested structs, slices, and maps, dereferencing pointers as needed.
// It also handles Duration type conversion.
// This code uses reflection. The go proverb says "Clear is better than clever. Reflection is never clear."
// This is an example of that, but the alternative is a whole bunch of type-specific code to do this for every possible type.
// It was written with a lot of help from Claude, and passes its tests.
// If this function returns an error, it is almost certainly due to a component design issue.
func setMemberValue(key string, member any, value any) error {
	memberValue := reflect.ValueOf(member)
	// Always dereference pointer(s) to get to the struct for FieldByName
	for memberValue.Kind() == reflect.Ptr {
		if memberValue.IsNil() {
			return fmt.Errorf("nil pointer encountered for key %s", key)
		}
		memberValue = memberValue.Elem()
	}
	if memberValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s for key %s", memberValue.Kind(), key)
	}

	parts := strings.SplitN(key, ".", 2)

	if len(parts) == 2 {
		field := memberValue.FieldByName(parts[0])
		if !field.IsValid() {
			return fmt.Errorf("member %s is not a valid field in type %T", parts[0], memberValue.Interface())
		}

		// Always initialize pointer fields if nil before recursing
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		} else if field.Kind() == reflect.Slice {
			subParts := strings.SplitN(parts[1], ".", 2)
			index, err := strconv.Atoi(subParts[0])
			if err != nil {
				return fmt.Errorf("invalid slice index: %s", subParts[0])
			}
			var subKey string
			if len(subParts) == 2 {
				subKey = subParts[1]
			}
			return setSliceElementValue(field, index, subKey, value)
		}

		nextMember := field.Interface()
		if field.Kind() != reflect.Ptr && field.CanAddr() {
			nextMember = field.Addr().Interface()
		}
		return setMemberValue(parts[1], nextMember, value)
	} else {
		field := memberValue.FieldByName(parts[0])
		if !field.IsValid() {
			return fmt.Errorf("field %s not found in type %T", parts[0], memberValue.Interface())
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		// At the end of the path, set the value with type conversion if needed
		v := reflect.ValueOf(value)
		if v.Type() != field.Type() {
			// Special handling for Duration type
			if field.Type() == reflect.TypeOf(Duration(0)) && v.Kind() == reflect.String {
				dur, err := time.ParseDuration(v.String())
				if err != nil {
					return fmt.Errorf("invalid duration format: %s", v.String())
				}
				v = reflect.ValueOf(Duration(dur))
			} else if v.Type().ConvertibleTo(field.Type()) {
				v = v.Convert(field.Type())
			} else {
				return fmt.Errorf("cannot assign value of type %s to field of type %s", v.Type(), field.Type())
			}
		}
		field.Set(v)
	}
	return nil
}
