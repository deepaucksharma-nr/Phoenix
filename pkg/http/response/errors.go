package response

import (
	"net/http"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error sends an error response with the given status code and message
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

// ErrorWithCode sends an error response with an error code
func ErrorWithCode(w http.ResponseWriter, status int, message, code string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(w http.ResponseWriter, status int, message string, details map[string]interface{}) {
	JSON(w, status, ErrorResponse{
		Error:   message,
		Details: details,
	})
}

// Common error responses
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message)
}

func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}

func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, message)
}