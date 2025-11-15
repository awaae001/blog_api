package model

// ApiResponse 统一API响应结构
type ApiResponse struct {
	Code    int         `json:"code"`    // HTTP状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Items    interface{} `json:"items"`     // 数据列表
	Total    int         `json:"total"`     // 总数量
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页数量
}

// FriendLinkDTO 友链数据传输对象（不包含敏感字段times）
type FriendLinkDTO struct {
	ID     int    `json:"id"`     // 友链ID
	Name   string `json:"name"`   // 网站名称
	Link   string `json:"link"`   // 网站链接
	Avatar string `json:"avatar"` // 网站图标
	Info   string `json:"info"`   // 网站描述
	Status string `json:"status"` // 网站状态
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) ApiResponse {
	return ApiResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) ApiResponse {
	return ApiResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
