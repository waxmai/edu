package dto

import "time"

type RescheduleRecordCreateRequest struct {
	OldScheduleID int32  `json:"oldScheduleId" binding:"required"`
	NewScheduleID int32  `json:"newScheduleId"`
	OperationType string `json:"operationType" binding:"required"`
	Reason        string `json:"reason"`
	OperatorID    int32  `json:"operatorId"`
}

type RescheduleRecordUpdateRequest map[string]interface{}

type RescheduleRecordListQuery struct {
	StudentID     int32  `form:"studentId" json:"studentId"`
	OperationType string `form:"operationType" json:"operationType"`
	StartDate     string `form:"startDate" json:"startDate"`
	EndDate       string `form:"endDate" json:"endDate"`
	PageNum       int    `form:"pageNum" json:"pageNum"`
	PageSize      int    `form:"pageSize" json:"pageSize"`
}

type RescheduleRecordResponse struct {
	ID                 int32     `json:"id"`
	OldScheduleID      int32     `json:"oldScheduleId"`
	OldScheduleSummary string    `json:"oldScheduleSummary"`
	NewScheduleID      int32     `json:"newScheduleId"`
	NewScheduleSummary string    `json:"newScheduleSummary"`
	OperationType      string    `json:"operationType"`
	Reason             string    `json:"reason"`
	OperatorID         int32     `json:"operatorId"`
	OperatorName       string    `json:"operatorName"`
	CreatedAt          time.Time `json:"createdAt"`
}

type RescheduleRecordListResponse []RescheduleRecordResponse
