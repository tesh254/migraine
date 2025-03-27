package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func ExtractTemplateVars(content string) []string {
	re := regexp.MustCompile(`{{([^}]+)}}`)
	matches := re.FindAllStringSubmatch(content, -1)

	// If no matches found, return nil instead of empty slice
	if len(matches) == 0 {
		return nil
	}

	// Use map to deduplicate variables
	varsMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varsMap[match[1]] = true
		}
	}

	// Convert map keys to slice
	variables := make([]string, 0, len(varsMap))
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

func ReplaceVariables(content string, values map[string]string) (string, error) {
	variables := ExtractTemplateVars(content)

	missingVars := []string{}
	for _, v := range variables {
		if _, exists := values[v]; !exists {
			missingVars = append(missingVars, v)
		}
	}

	if len(missingVars) > 0 {
		return "", fmt.Errorf("Missing required variables: %s", strings.Join(missingVars, ", "))
	}

	result := content
	for variable, value := range values {
		pattern := fmt.Sprintf("{{%s}}", regexp.QuoteMeta(variable))
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, value)
	}

	return result, nil
}

func ApplyVariablesToCommand(command string, variables map[string]string) (string, error) {
	if command == "" {
		return "", nil
	}

	replacedCommand, err := ReplaceVariables(command, variables)
	if err != nil {
		return "", fmt.Errorf("failed to replace variables in command: %w", err)
	}

	return replacedCommand, nil
}

func ValidateVariables(required []string, values map[string]string) error {
	missingVars := []string{}

	for _, req := range required {
		if _, exists := values[req]; !exists {
			missingVars = append(missingVars, req)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

// SanitizeVariableValue sanitizes a variable value to prevent command injection
func SanitizeVariableValue(value string) string {
	unsafe := []string{";", "&", "|", ">", "<", "`", "$", "(", ")", "{", "}", "[", "]", "\"", "'", "\n", "\r"}
	result := value

	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "")
	}

	return result
}

func PreprocessVariables(variables map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range variables {
		result[key] = SanitizeVariableValue(value)
	}

	return result
}
