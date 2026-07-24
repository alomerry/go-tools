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
	// Always include the HTTP status and the top-level message Apollo returns.
	// Preserve the exception (stack trace / root cause) and timestamp fields so
	// callers can see the exact reason Apollo rejected the request, e.g. when a
	// token lacks publish permission Apollo returns a 403 with an exception that
	// would otherwise be silently dropped.
	msg := fmt.Sprintf("apollo: %d %s", e.Status, e.Message)
	if e.Exception != "" {
		msg += ", exception: " + e.Exception
	}
	if e.Timestamp != "" {
		msg += ", timestamp: " + e.Timestamp
	}
	return msg
}
