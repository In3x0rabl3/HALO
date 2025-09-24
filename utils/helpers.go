package utils

import (
	"encoding/json"
)

// ==============================
// Generic Helpers
// ==============================

// Contains checks if a string slice contains a given target
func Contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// PrettyJSON indents any Go value into a human-readable JSON string
func PrettyJSON(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
