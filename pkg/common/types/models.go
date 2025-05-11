package types

import "github.com/turtacn/dataseap/pkg/common/errors"

// PaginationRequest 分页请求参数结构
// PaginationRequest structure for pagination request parameters.
type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"omitempty,min=1"`         // 页码 Page number
	PageSize int `json:"pageSize" form:"pageSize" binding:"omitempty,min=1"` // 每页大小 Page size
}

// GetOffset 计算数据库查询的偏移量
// GetOffset calculates the offset for database queries.
func (pr *PaginationRequest) GetOffset() int {
	if pr.Page <= 0 {
		pr.Page = 1
	}
	if pr.PageSize <= 0 {
		// Use a default page size if not specified or invalid, e.g., from constants
		// pr.PageSize = constants.DefaultPageSize (assuming constants package exists)
		pr.PageSize = 10 // Placeholder
	}
	return (pr.Page - 1) * pr.PageSize
}

// GetLimit 获取查询的限制数量
// GetLimit retrieves the limit for queries.
func (pr *PaginationRequest) GetLimit() int {
	if pr.PageSize <= 0 {
		// pr.PageSize = constants.DefaultPageSize
		pr.PageSize = 10 // Placeholder
	}
	// Optionally cap page size to a maximum
	// if pr.PageSize > constants.MaxPageSize {
	//     pr.PageSize = constants.MaxPageSize
	// }
	return pr.PageSize
}

// PaginationResponse 分页响应参数结构
// PaginationResponse structure for pagination response parameters.
type PaginationResponse struct {
	Page     int   `json:"page"`     // 当前页码 Current page number
	PageSize int   `json:"pageSize"` // 每页大小 Page size
	Total    int64 `json:"total"`    // 总记录数 Total number of records
}

// APIResponse 通用API响应结构体
// APIResponse generic API response structure.
type APIResponse struct {
	Success bool         `json:"success"`         // 操作是否成功 Indicates if the operation was successful
	Code    string       `json:"code"`            // 业务状态码 Business status code (can be "OK" or an error code string)
	Message string       `json:"message"`         // 提示信息 Message
	Data    interface{}  `json:"data"`            // 响应数据 Response data
	Error   *ErrorDetail `json:"error,omitempty"` // 错误详情 (仅在 Success 为 false 时出现) Error details (only when Success is false)
}

// ErrorDetail API错误响应中的错误详情
// ErrorDetail provides detailed error information in API responses.
type ErrorDetail struct {
	Code       errors.ErrorCode `json:"code"`                 // 具体的错误码 Specific error code
	Message    string           `json:"message"`              // 错误信息 Error message
	Details    interface{}      `json:"details,omitempty"`    // 更详细的错误信息或验证错误 More detailed error information or validation errors
	StackTrace string           `json:"stackTrace,omitempty"` // 堆栈信息 (通常仅在开发模式下暴露) Stack trace (usually exposed only in development mode)
}

// NewSuccessAPIResponse 创建一个成功的API响应
// NewSuccessAPIResponse creates a successful API response.
func NewSuccessAPIResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Code:    "OK",
		Message: "Operation successful",
		Data:    data,
	}
}

// NewErrorAPIResponse 创建一个失败的API响应
// NewErrorAPIResponse creates a failed API response.
func NewErrorAPIResponse(appErr *errors.AppError, details ...interface{}) *APIResponse {
	errDetail := &ErrorDetail{
		Code:    appErr.Code,
		Message: appErr.Message,
		// StackTrace: appErr.StackTrace, // Be cautious about exposing stack traces
	}
	if len(details) > 0 {
		errDetail.Details = details[0]
	}

	return &APIResponse{
		Success: false,
		Code:    string(appErr.Code), // Use ErrorCode string directly
		Message: appErr.Message,
		Data:    nil,
		Error:   errDetail,
	}
}

// SortOrder 排序顺序
// SortOrder defines the sort order for queries.
type SortOrder string

const (
	// SortOrderAsc 升序
	// SortOrderAsc ascending order.
	SortOrderAsc SortOrder = "ASC"
	// SortOrderDesc 降序
	// SortOrderDesc descending order.
	SortOrderDesc SortOrder = "DESC"
)

// String returns the string representation of SortOrder.
func (so SortOrder) String() string {
	return string(so)
}

// IsValid checks if the SortOrder value is valid.
func (so SortOrder) IsValid() bool {
	switch so {
	case SortOrderAsc, SortOrderDesc:
		return true
	}
	return false
}

// SortField 定义排序字段和顺序
// SortField defines a field to sort by and the order.
type SortField struct {
	Field string    `json:"field"` // 排序字段 Sort field
	Order SortOrder `json:"order"` // 排序顺序 Sort order (ASC or DESC)
}
