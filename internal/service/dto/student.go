package dto

import "time"

type StudentCreateRequest struct {
	StudentName  string `json:"studentName" binding:"required"`
	Gender       string `json:"gender"`
	Grade        string `json:"grade"`
	Phone        string `json:"phone"`
	ParentName   string `json:"parentName"`
	ParentPhone  string `json:"parentPhone"`
	Subject      string `json:"subject" binding:"required"`
	TeachingType string `json:"teachingType" binding:"required"`
	Status       string `json:"status"`
	Remark       string `json:"remark"`
}

type StudentUpdateRequest map[string]interface{}

type StudentListQuery struct {
	StudentName   string `form:"studentName" json:"studentName"`
	Subject       string `form:"subject" json:"subject"`
	Status        string `form:"status" json:"status"`
	ParentPhone   string `form:"parentPhone" json:"parentPhone"`
	InactiveAlert bool   `form:"inactiveAlert" json:"inactiveAlert"`
	PageNum       int    `form:"pageNum" json:"pageNum"`
	PageSize      int    `form:"pageSize" json:"pageSize"`
}

type StudentResponse struct {
	ID           int32     `json:"id"`
	StudentName  string    `json:"studentName"`
	Gender       string    `json:"gender"`
	Grade        string    `json:"grade"`
	Phone        string    `json:"phone"`
	ParentName   string    `json:"parentName"`
	ParentPhone  string    `json:"parentPhone"`
	Subject      string    `json:"subject"`
	TeachingType string    `json:"teachingType"`
	Status       string    `json:"status"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type StudentListResponse []StudentResponse
