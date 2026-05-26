package dto

import "time"

type LessonPackageCreateRequest struct {
	OrganizationID     int32   `json:"-"`
	CampusID           int32   `json:"-"`
	StudentID          int32   `json:"studentId" binding:"required"`
	CourseID           int32   `json:"courseId" binding:"required"`
	TotalLessons       float64 `json:"totalLessons" binding:"required"`
	UsedLessons        float64 `json:"usedLessons"`
	RemainLessons      float64 `json:"remainLessons"`
	TotalAmount        float64 `json:"totalAmount"`
	PaidAmount         float64 `json:"paidAmount"`
	StartDate          string  `json:"startDate"`
	EndDate            string  `json:"endDate"`
	Status             string  `json:"status"`
	LowLessonThreshold float64 `json:"lowLessonThreshold"`
	Remark             string  `json:"remark"`
}

type LessonPackageUpdateRequest map[string]interface{}

type LessonPackageListQuery struct {
	StudentID      int32  `form:"studentId" json:"studentId"`
	CourseID       int32  `form:"courseId" json:"courseId"`
	Status         string `form:"status" json:"status"`
	LowLessonAlert bool   `form:"lowLessonAlert" json:"lowLessonAlert"`
	PaymentStatus  string `form:"paymentStatus" json:"paymentStatus"`
	PageNum        int    `form:"pageNum" json:"pageNum"`
	PageSize       int    `form:"pageSize" json:"pageSize"`
}

type LessonPackageResponse struct {
	ID                 int32     `json:"id"`
	StudentID          int32     `json:"studentId"`
	StudentName        string    `json:"studentName"`
	CourseID           int32     `json:"courseId"`
	CourseName         string    `json:"courseName"`
	Subject            string    `json:"subject"`
	TotalLessons       float64   `json:"totalLessons"`
	UsedLessons        float64   `json:"usedLessons"`
	RemainLessons      float64   `json:"remainLessons"`
	TotalAmount        float64   `json:"totalAmount"`
	PaidAmount         float64   `json:"paidAmount"`
	StartDate          string    `json:"startDate"`
	EndDate            string    `json:"endDate"`
	Status             string    `json:"status"`
	LowLessonThreshold float64   `json:"lowLessonThreshold"`
	Remark             string    `json:"remark"`
	LowLessonAlert     bool      `json:"lowLessonAlert"`
	ArrearsAmount      float64   `json:"arrearsAmount"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type LessonPackageListResponse []LessonPackageResponse
