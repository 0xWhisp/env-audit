package parser

// CompareResult contains the comparison between target and example env files
type CompareResult struct {
	Missing []string // keys in example but not in target
	Extra   []string // keys in target but not in example
}

// Compare compares target env against example env
// Returns keys missing from target and extra keys in target
func Compare(target, example map[string]string) *CompareResult {
	result := &CompareResult{
		Missing: []string{},
		Extra:   []string{},
	}

	// Find keys in example but not in target (missing)
	for key := range example {
		if _, exists := target[key]; !exists {
			result.Missing = append(result.Missing, key)
		}
	}

	// Find keys in target but not in example (extra)
	for key := range target {
		if _, exists := example[key]; !exists {
			result.Extra = append(result.Extra, key)
		}
	}

	return result
}
