package dto

import "time"

type PaymentRecordCreateRequest struct {
	OrganizationID  int32   `json:"-"`
	CampusID        int32   `json:"-"`
	StudentID       int32   `json:"studentId" binding:"required"`
	LessonPackageID int32   `json:"lessonPackageId" binding:"required"`
	PaymentType     string  `json:"paymentType" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	PaymentMethod   string  `json:"paymentMethod" binding:"required"`
	PaymentTime     string  `json:"paymentTime" binding:"required"`
	PaymentStatus   string  `json:"paymentStatus"`
	Remark          string  `json:"remark"`
}

type PaymentRecordUpdateRequest map[string]interface{}

type PaymentRecordListQuery struct {
	StudentID       int32  `form:"studentId" json:"studentId"`
	LessonPackageID int32  `form:"lessonPackageId" json:"lessonPackageId"`
	PaymentType     string `form:"paymentType" json:"paymentType"`
	PaymentStatus   string `form:"paymentStatus" json:"paymentStatus"`
	PaymentMethod   string `form:"paymentMethod" json:"paymentMethod"`
	StartDate       string `form:"startDate" json:"startDate"`
	EndDate         string `form:"endDate" json:"endDate"`
	PageNum         int    `form:"pageNum" json:"pageNum"`
	PageSize        int    `form:"pageSize" json:"pageSize"`
}

type PaymentIncomeStatsQuery struct {
	StartDate string `form:"startDate" json:"startDate"`
	EndDate   string `form:"endDate" json:"endDate"`
	GroupBy   string `form:"groupBy" json:"groupBy"`
}

type PaymentRecordResponse struct {
	ID                int32     `json:"id"`
	StudentID         int32     `json:"studentId"`
	StudentName       string    `json:"studentName"`
	LessonPackageID   int32     `json:"lessonPackageId"`
	LessonPackageName string    `json:"lessonPackageName"`
	CourseName        string    `json:"courseName"`
	PaymentType       string    `json:"paymentType"`
	Amount            float64   `json:"amount"`
	PaymentMethod     string    `json:"paymentMethod"`
	PaymentTime       time.Time `json:"paymentTime"`
	PaymentStatus     string    `json:"paymentStatus"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type PaymentRecordListResponse []PaymentRecordResponse

type PaymentIncomeStatsResponse struct {
	TotalAmount float64 `json:"totalAmount"`
	RecordCount int     `json:"recordCount"`
}
