package response

import "net/http"

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// Pagination 分页响应
func Pagination(items interface{}, total int64, page, pageSize int) Response {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return Response{
		Code:    0,
		Message: "success",
		Data: PaginationResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}
}

// HTTPStatus 获取HTTP状态码
func (r Response) HTTPStatus() int {
	if r.Code == 0 {
		return http.StatusOK
	}
	if r.Code >= 400 && r.Code < 500 {
		return r.Code
	}
	return http.StatusInternalServerError
}
