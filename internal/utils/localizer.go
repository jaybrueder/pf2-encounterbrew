package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Localizer struct {
	data    map[string]interface{}
	pattern *regexp.Regexp
}

var (
	instance *Localizer
	once     sync.Once
	initErr  error
)

// GetLocalizer returns the singleton instance of Localizer
func GetLocalizer(jsonPath string) (*Localizer, error) {
	once.Do(func() {
		instance, initErr = initLocalizer(jsonPath)
	})
	return instance, initErr
}

// initLocalizer is the actual initialization function
func initLocalizer(jsonPath string) (*Localizer, error) {
	// Read JSON file
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON into nested map
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	return &Localizer{
		data:    data,
		pattern: regexp.MustCompile(`@Localize\[([^\]]+)\]`),
	}, nil
}

// getValueFromPath retrieves a value from nested maps using a dot-separated path
func (l *Localizer) getValueFromPath(path string) (string, error) {
	parts := strings.Split(path, ".")
	current := l.data

	// Navigate through the nested maps
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - should be the actual string value
			if value, ok := current[part].(string); ok {
				return l.cleanText(value), nil
			}
			return "", fmt.Errorf("value at path %s is not a string", path)
		}

		// Navigate deeper into the nested structure
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return "", fmt.Errorf("invalid path at %s", strings.Join(parts[:i+1], "."))
		}
	}

	return "", fmt.Errorf("path not found: %s", path)
}

func (l *Localizer) cleanText(text string) string {
	// Convert \n to actual newlines
	text = strings.ReplaceAll(text, "\\n", "\n")

	// Trim spaces and remove extra newlines
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	return text
}

// ProcessText replaces all @Localize patterns with their corresponding text
func (l *Localizer) ProcessText(input string) string {
	return l.pattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract path from @Localize[...]
		path := l.pattern.FindStringSubmatch(match)[1]

		// Get the localized text
		text, err := l.getValueFromPath(path)
		if err != nil {
			// In case of error, return the original pattern
			// You might want to log the error here
			return match
		}

		return text
	})
}
