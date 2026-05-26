package tenant

import (
	"encoding/json"
	"strings"
)

func marshalFeatureFlags(flags []string) string {
	clean := make([]string, 0, len(flags))
	seen := map[string]struct{}{}
	for _, item := range flags {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		clean = append(clean, item)
	}
	if len(clean) == 0 {
		return "[]"
	}
	buf, _ := json.Marshal(clean)
	return string(buf)
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
