package loader

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
)

// TODO we really need this? can't we use a lib or rely on json.Unmarshal or yaml parsing?

// parseBool parses a string as a boolean value.
// Accepts: "true", "false", "1", "0", "yes", "no" (case-insensitive).
func parseBool(s string) (bool, error) {
	return strconv.ParseBool(strings.ToLower(s))
}

// parseInt parses a string as an integer.
func parseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

// parseDuration parses a string as a time.Duration.
// Uses Go's standard time.ParseDuration format (e.g., "30s", "5m", "1h30m").
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// parseSize parses a string as a byte size.
// Uses docker/go-units format (e.g., "10GB", "500MB", "1024").
// Returns the size in bytes as int64.
func parseSize(s string) (int64, error) {
	return units.RAMInBytes(s)
}

// parseEnum validates that a string value is in the allowed list.
// Returns the value if valid, otherwise returns an error.
func parseEnum(s string, allowed []string) (string, bool) {
	for _, a := range allowed {
		if s == a {
			return s, true
		}
	}
	return "", false
}

// parseList parses a string as a list of strings.
// Supports two formats:
//   - JSON array: ["a", "b", "c"]
//   - CSV: a,b,c
//
// The format is auto-detected based on whether the string starts with '['.
func parseList(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Try JSON first if it looks like JSON
	if strings.HasPrefix(s, "[") {
		var result []string
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return nil, err
		}
		return result, nil
	}

	// Fall back to CSV
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result, nil
}
