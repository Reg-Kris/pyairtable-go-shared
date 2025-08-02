package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty and contains non-whitespace characters
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// TrimAndClean trims whitespace and removes extra spaces
func TrimAndClean(s string) string {
	// Trim leading and trailing whitespace
	s = strings.TrimSpace(s)
	
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

// ToSnakeCase converts camelCase or PascalCase to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	
	return result.String()
}

// ToCamelCase converts snake_case to camelCase
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return s
	}
	
	var result strings.Builder
	result.WriteString(strings.ToLower(parts[0]))
	
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result.WriteString(strings.ToUpper(string(parts[i][0])))
			if len(parts[i]) > 1 {
				result.WriteString(strings.ToLower(parts[i][1:]))
			}
		}
	}
	
	return result.String()
}

// ToPascalCase converts snake_case to PascalCase
func ToPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	
	return result.String()
}

// Truncate truncates a string to the specified length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	
	if length <= 3 {
		return s[:length]
	}
	
	return s[:length-3] + "..."
}

// Contains checks if a slice of strings contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if a slice of strings contains a specific string (case insensitive)
func ContainsIgnoreCase(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// IsValidEmail checks if a string is a valid email address
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// MaskEmail masks an email address for privacy
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 2 {
		return email
	}
	
	masked := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	return masked + "@" + domain
}

// SanitizeString removes potentially harmful characters from a string
func SanitizeString(s string) string {
	// Remove control characters
	var result strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			continue
		}
		result.WriteRune(r)
	}
	
	return result.String()
}

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	
	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	
	// Remove leading and trailing hyphens
	s = strings.Trim(s, "-")
	
	return s
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// WordCount counts the number of words in a string
func WordCount(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// Ellipsis adds ellipsis to a string if it exceeds the maximum length
func Ellipsis(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	if maxLength <= 3 {
		return s[:maxLength]
	}
	
	return s[:maxLength-3] + "..."
}