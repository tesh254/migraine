package utils

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

func ExtractTemplateVars(content string) []string {
	var variables []string
	re := regexp.MustCompile(`{{([^}]+)}}`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Use map to deduplicate variables
	varsMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varsMap[match[1]] = true
		}
	}

	// Convert map keys to slice
	for v := range varsMap {
		variables = append(variables, v)
	}

	sort.Strings(variables) // Sort for consistent output
	return variables
}

func ExtractEnvVarsFromJSON(jsonStr string) ([]string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}

	envVars := make(map[string]bool)
	extractEnvVars(data, envVars)

	result := make([]string, 0, len(envVars))
	for v := range envVars {
		result = append(result, v)
	}

	return result, nil
}

func extractEnvVars(v interface{}, envVars map[string]bool) {
	switch value := v.(type) {
	case map[string]interface{}:
		for _, v := range value {
			extractEnvVars(v, envVars)
		}
	case []interface{}:
		for _, v := range value {
			extractEnvVars(v, envVars)
		}
	case string:
		if isEnvVar(value) {
			envVars[value] = true
		}
	}
}

func isEnvVar(s string) bool {
	return strings.ToUpper(s) == s && strings.Contains(s, "_") && !strings.Contains(s, " ")
}
