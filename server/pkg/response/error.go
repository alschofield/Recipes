package response

import (
	"encoding/json"
	"net/http"
)

// APIError is the standard error response format.
type APIError struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, status int, message, code string) {
	WriteJSON(w, status, APIError{
		Error: message,
		Code:  code,
	})
}

// WriteErrorDetails writes a standardized error response with extra context.
func WriteErrorDetails(w http.ResponseWriter, status int, message, code, details string) {
	WriteJSON(w, status, APIError{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// BadRequest is a shorthand for 400 errors.
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message, "BAD_REQUEST")
}

// Unauthorized is a shorthand for 401 errors.
func Unauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, message, "UNAUTHORIZED")
}

// Forbidden is a shorthand for 403 errors.
func Forbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, message, "FORBIDDEN")
}

// NotFound is a shorthand for 404 errors.
func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message, "NOT_FOUND")
}

// Conflict is a shorthand for 409 errors.
func Conflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message, "CONFLICT")
}

// InternalError is a shorthand for 500 errors.
func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message, "INTERNAL_ERROR")
}
