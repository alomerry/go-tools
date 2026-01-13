package resty

import (
	"net/http"

	"resty.dev/v3"
)

func DefaultRetryCondition(r *resty.Response, err error) bool {
	if err != nil {
		// 网络错误重试
		return true
	}

	// 特定状态码重试
	statusCode := r.StatusCode()
	retryCodes := map[int]bool{
		http.StatusRequestTimeout:      true,
		http.StatusTooManyRequests:     true,
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}

	if _, ok := retryCodes[statusCode]; ok {
		return true
	}

	return false
}
