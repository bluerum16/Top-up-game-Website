package model

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func SuccessResponse(data any) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func MessageResponse(msg string) APIResponse {
	return APIResponse{Success: true, Message: msg}
}

func ErrorResponse(msg string) APIResponse {
	return APIResponse{Success: false, Error: msg}
}
