package entity

type ErrorResponse struct {
	Code    int32  `json:"code"  binding:"required"`
	Message string `json:"message" binding:"required"`
}
