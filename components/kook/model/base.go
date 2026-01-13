package model

type BaseResponse struct {
	Message string         `json:"message"`
	Code    int32          `json:"code"`
	Data    map[string]any `json:"data"`
}
