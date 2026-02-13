package sdk

import "fmt"

// ErrorResponse represents an error response from the Apollo Open API.
type ErrorResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Exception string `json:"exception"`
	Timestamp string `json:"timestamp"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("apollo: %d %s", e.Status, e.Message)
}
