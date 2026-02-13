package model

type BaseResponse struct {
	Message string         `json:"message"`
	Code    int32          `json:"code"`
	Data    map[string]any `json:"data"`
}

type Meta struct {
	Page      int `json:"page"`
	PageTotal int `json:"page_total"`
	PageSize  int `json:"page_size"`
	Total     int `json:"total"`
}

type Sort struct {
	Id int `json:"id"`
}
