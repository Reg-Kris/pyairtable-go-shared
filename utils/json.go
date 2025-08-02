package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

// ToJSON converts a struct to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

// ToJSONBytes converts a struct to JSON bytes
func ToJSONBytes(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return data, nil
}

// ToJSONIndent converts a struct to indented JSON string
func ToJSONIndent(v interface{}, prefix, indent string) (string, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON parses JSON string into a struct
func FromJSON(jsonStr string, v interface{}) error {
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// FromJSONBytes parses JSON bytes into a struct
func FromJSONBytes(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// FromJSONReader parses JSON from an io.Reader into a struct
func FromJSONReader(reader io.Reader, v interface{}) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(jsonStr string) bool {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp) == nil
}

// PrettyJSON formats JSON string with indentation
func PrettyJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	
	formatted, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	
	return string(formatted), nil
}

// CompactJSON removes unnecessary whitespace from JSON string
func CompactJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	
	compact, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to compact JSON: %w", err)
	}
	
	return string(compact), nil
}

// JSONSize returns the size of JSON representation in bytes
func JSONSize(v interface{}) (int, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return len(data), nil
}

// CloneViaJSON clones an object by marshaling to JSON and back
func CloneViaJSON(src, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}
	
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}
	
	return nil
}

// JSONPath extracts a value from JSON using a simple path (dot notation)
func JSONPath(jsonStr, path string) (interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	
	keys := splitJSONPath(path)
	current := interface{}(data)
	
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[key]; ok {
				current = val
			} else {
				return nil, fmt.Errorf("key '%s' not found", key)
			}
		default:
			return nil, fmt.Errorf("cannot navigate into non-object at key '%s'", key)
		}
	}
	
	return current, nil
}

// splitJSONPath splits a dot-notation path into individual keys
func splitJSONPath(path string) []string {
	if path == "" {
		return []string{}
	}
	// Simple implementation - doesn't handle escaped dots
	return strings.Split(path, ".")
}

// MergeJSON merges two JSON objects (second object takes precedence)
func MergeJSON(json1, json2 string) (string, error) {
	var obj1, obj2 map[string]interface{}
	
	if err := json.Unmarshal([]byte(json1), &obj1); err != nil {
		return "", fmt.Errorf("invalid first JSON: %w", err)
	}
	
	if err := json.Unmarshal([]byte(json2), &obj2); err != nil {
		return "", fmt.Errorf("invalid second JSON: %w", err)
	}
	
	merged := mergeObjects(obj1, obj2)
	
	result, err := json.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged JSON: %w", err)
	}
	
	return string(result), nil
}

// mergeObjects recursively merges two objects
func mergeObjects(obj1, obj2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy obj1
	for k, v := range obj1 {
		result[k] = v
	}
	
	// Merge obj2
	for k, v := range obj2 {
		if existing, ok := result[k]; ok {
			// If both are objects, merge recursively
			if existingMap, ok1 := existing.(map[string]interface{}); ok1 {
				if vMap, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = mergeObjects(existingMap, vMap)
					continue
				}
			}
		}
		// Otherwise, obj2 value takes precedence
		result[k] = v
	}
	
	return result
}