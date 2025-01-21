package tsdb

import "errors"

var (
	ErrEmptyEndpoint = errors.New("endpoint is empty")
	ErrEmptyOrg      = errors.New("org is empty")
	ErrEmptyBucket   = errors.New("bucket is empty")
	ErrUnhealthy     = errors.New("unhealthy")
)
