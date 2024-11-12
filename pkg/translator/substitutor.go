package translator

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

type SubstitutionContext map[string]string

type Substitutor struct {
	// substitutions is a map of context names to collections of key-value
	// pairs. The key-value pairs are the substitutions that will be made when
	// DoSubstitutions is called. A substitution looks like ${context.key} and
	// will be replaced with the matching value within that context. If the
	// context is not specified or doesn't exist, the entire key will be
	// searched for in all contexts.
	// Substitutions have no knowledge of syntax, so they can be used for any
	// kind of substitution.
	substitutions map[string]SubstitutionContext
	// Contexts have priorities.
	// priorities is a map of context names to their priority. The higher the
	// number, the higher the priority. If a key with unspecified context is
	// found in multiple contexts, the context with the highest priority will
	// be used.
	// Contexts will be prioritized in the order they are added to the system
	// (first added is highest priority).
	priorities      map[string]int
	currentPriority int
}

func NewSubstitutor() *Substitutor {
	return &Substitutor{
		substitutions: make(map[string]SubstitutionContext),
		priorities:    make(map[string]int),
	}
}

// AddContext adds an entire new context to the substitutor. The context is a map of
// key-value pairs that will be used for substitutions.
func (s *Substitutor) AddContext(name string, context SubstitutionContext, priority int) {
	s.substitutions[name] = context
	s.priorities[name] = priority
}

// AddSubstitution adds a single key-value pair to a context. If the context
// doesn't exist, it will be created.
func (s *Substitutor) AddSubstitution(ctx, key, value string) {
	context, ok := s.substitutions[ctx]
	if !ok {
		context = make(SubstitutionContext)
		s.substitutions[ctx] = context
	}
	context[key] = value
	s.currentPriority--
	s.priorities[ctx] = s.currentPriority
}

func (s *Substitutor) SetPriority(name string, priority int) {
	s.priorities[name] = priority
}

func (s *Substitutor) DoSubstitutions(input string) string {
	// All substitutions look like ${varname}, where varname might contain a
	// context prefix, e.g. ${context.varname}. If context exists, we want to
	// substitute the entire value including the ${} with the value of the key
	// in the context. If the context doesn't exist, we search all contexts for
	// a matching varname and substitute that.
	// If no substitution is found, we leave the string as-is.

	// First, find all the potential substitutions
	pat := regexp.MustCompile(`\${([^.}]+\.)?([^}]+)}`)
	matches := pat.FindAllStringSubmatchIndex(input, -1)
	if matches == nil {
		return input
	}

	// Next, iterate over the matches and do the substitutions
	// match[0,1] is the full match, match[2,3] is the context, match[4,5] is the key
	var output strings.Builder
	ix := 0
outer:
	for _, match := range matches {
		output.WriteString(input[ix:match[0]])
		// no matter what, at the end of this iteration, we want to skip to the end of the match
		ix = match[1]
		key := input[match[4]:match[5]]
		// first, see if we have a context that we know about
		// if so, we can just search that context
		context := ""
		if match[3]-match[2] != 0 {
			context = input[match[2] : match[3]-1] // skip the trailing .
			if ctx, ok := s.substitutions[context]; ok {
				if val, ok := ctx[key]; ok {
					output.WriteString(val)
					continue outer
				}
			}
			// we had a context, but it didn't match, so we put it back on the key
			key = context + "." + key
		}
		// if we get here, we didn't find a context, so we need to search all contextNames
		// for the entire key.
		// we want to search the contextNames in order of priority
		contextNames := slices.SortedFunc(maps.Keys(s.substitutions), func(a string, b string) int {
			return s.priorities[b] - s.priorities[a]
		})

		for _, ctxName := range contextNames {
			context := s.substitutions[ctxName]
			if val, ok := context[key]; ok {
				output.WriteString(val)
				continue outer
			}
		}

		// if we get here, we didn't find a substitution, so we leave the string as-is
		output.WriteString(input[match[0]:match[1]])
	}
	// write out the rest of the string
	output.WriteString(input[ix:])
	return output.String()
}
