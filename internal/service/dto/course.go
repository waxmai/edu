package dto

import "time"

type CourseCreateRequest struct {
	CourseName      string  `json:"courseName" binding:"required"`
	Subject         string  `json:"subject" binding:"required"`
	CourseType      string  `json:"courseType" binding:"required"`
	DurationMinutes int32   `json:"durationMinutes"`
	FeeStandard     float64 `json:"feeStandard"`
	Remark          string  `json:"remark"`
	Status          string  `json:"status"`
}

type CourseUpdateRequest map[string]interface{}

type CourseResponse struct {
	ID              int32     `json:"id"`
	CourseName      string    `json:"courseName"`
	Subject         string    `json:"subject"`
	CourseType      string    `json:"courseType"`
	DurationMinutes int32     `json:"durationMinutes"`
	FeeStandard     float64   `json:"feeStandard"`
	Remark          string    `json:"remark"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CourseListResponse []CourseResponse
