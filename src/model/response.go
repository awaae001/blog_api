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
	ID             int    `json:"id"`
	WebsiteName    string `json:"website_name"`
	WebsiteURL     string `json:"website_url"`
	WebsiteIconURL string `json:"website_icon_url"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	Email          string `json:"email,omitempty"`
	Times          int    `json:"times,omitempty"`
	EnableRss      bool   `json:"enable_rss"`
	UpdatedAt      int64  `json:"updated_at"`
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
