package dto

import "time"

// AdminCreateRequest 创建管理员请求参数
type AdminCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Mobile   string `json:"mobile" binding:"required"`
}

// AdminUpdateRequest 更新管理员请求参数
type AdminUpdateRequest map[string]interface{}

// AdminResponse 管理员响应数据，避免对外暴露生成的数据库模型。
type AdminResponse struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Mobile    string    `json:"mobile"`
	CreatedAt time.Time `json:"created_at"`
}

// AdminListResponse 管理员列表响应
type AdminListResponse []AdminResponse
