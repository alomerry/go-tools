package tsdb

import "errors"

var (
	ErrEmptyEndpoint = errors.New("endpoint is empty")
	ErrEmptyOrg      = errors.New("org is empty")
	ErrUnhealthy     = errors.New("unhealthy")
)
