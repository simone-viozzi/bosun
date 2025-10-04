package dockerlabels

import "strings"

// FilterByPrefixes filters a map of labels by allowed prefixes and drops empty values.
// It returns a new map containing only labels whose keys start with any of the provided prefixes,
// excluding any labels with empty or whitespace-only values.
// If no prefixes are provided, returns an empty map.
func FilterByPrefixes(in map[string]string, prefixes []string) map[string]string {
	if len(prefixes) == 0 {
		return make(map[string]string)
	}

	out := make(map[string]string)
	for k, v := range in {
		if strings.TrimSpace(v) == "" {
			continue
		}
		for _, p := range prefixes {
			if strings.HasPrefix(k, p) {
				out[k] = v
				break
			}
		}
	}
	return out
}
