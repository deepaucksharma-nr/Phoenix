package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MaxRequestBodySize is the maximum allowed request body size (10MB)
const MaxRequestBodySize = 10 * 1024 * 1024

// DecodeJSON decodes JSON from request body with size limit
func DecodeJSON(r *http.Request, v interface{}) error {
	// Limit request body size
	r.Body = http.MaxBytesReader(nil, r.Body, MaxRequestBodySize)
	
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		return fmt.Errorf("content type must be application/json, got %s", contentType)
	}
	
	// Decode JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(v); err != nil {
		if err == io.EOF {
			return fmt.Errorf("request body is empty")
		}
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	
	// Check for extra data
	if decoder.More() {
		return fmt.Errorf("request body contains extra data")
	}
	
	return nil
}

// ValidateRequired checks if required fields are present
func ValidateRequired(fields map[string]interface{}) error {
	var missing []string
	
	for name, value := range fields {
		switch v := value.(type) {
		case string:
			if v == "" {
				missing = append(missing, name)
			}
		case []string:
			if len(v) == 0 {
				missing = append(missing, name)
			}
		case nil:
			missing = append(missing, name)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid email format")
	}
	
	if !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email domain")
	}
	
	return nil
}

// ValidateStringLength validates string length
func ValidateStringLength(name, value string, min, max int) error {
	length := len(value)
	
	if length < min {
		return fmt.Errorf("%s must be at least %d characters long", name, min)
	}
	
	if max > 0 && length > max {
		return fmt.Errorf("%s must be at most %d characters long", name, max)
	}
	
	return nil
}