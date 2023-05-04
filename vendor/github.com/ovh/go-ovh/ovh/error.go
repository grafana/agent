package ovh

import "fmt"

// APIError represents an error that can occurred while calling the API.
type APIError struct {
	// Error class
	Class string `json:"class,omitempty"`
	// Error message.
	Message string `json:"message"`
	// Error details
	Details map[string]string `json:"details,omitempty"`
	// HTTP code.
	Code int
	// ID of the request
	QueryID string
}

func (err *APIError) Error() string {
	if err.Class == "" {
		return fmt.Sprintf("HTTP Error %d: %q", err.Code, err.Message)
	}

	return fmt.Sprintf("HTTP Error %d: %s: %q (X-OVH-Query-Id: %s)", err.Code, err.Class, err.Message, err.QueryID)
}
