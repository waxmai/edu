package dto

import "time"

type LessonRecordCreateRequest struct {
	ScheduleID        int32   `json:"scheduleId" binding:"required"`
	StudentID         int32   `json:"studentId" binding:"required"`
	TeacherID         int32   `json:"teacherId" binding:"required"`
	AttendanceStatus  string  `json:"attendanceStatus" binding:"required"`
	LessonContent     string  `json:"lessonContent"`
	Homework          string  `json:"homework"`
	Feedback          string  `json:"feedback"`
	NeedDeductLesson  *bool   `json:"needDeductLesson"`
	LessonDeducted    *bool   `json:"lessonDeducted"`
	DeductLessonCount float64 `json:"deductLessonCount"`
}

type LessonRecordUpdateRequest map[string]interface{}

type LessonRecordListQuery struct {
	StudentID        int32  `form:"studentId" json:"studentId"`
	TeacherID        int32  `form:"teacherId" json:"teacherId"`
	AttendanceStatus string `form:"attendanceStatus" json:"attendanceStatus"`
	StartDate        string `form:"startDate" json:"startDate"`
	EndDate          string `form:"endDate" json:"endDate"`
	PageNum          int    `form:"pageNum" json:"pageNum"`
	PageSize         int    `form:"pageSize" json:"pageSize"`
}

type LessonRecordResponse struct {
	ID                int32     `json:"id"`
	ScheduleID        int32     `json:"scheduleId"`
	StudentID         int32     `json:"studentId"`
	StudentName       string    `json:"studentName"`
	TeacherID         int32     `json:"teacherId"`
	TeacherName       string    `json:"teacherName"`
	AttendanceStatus  string    `json:"attendanceStatus"`
	LessonContent     string    `json:"lessonContent"`
	Homework          string    `json:"homework"`
	Feedback          string    `json:"feedback"`
	LessonDeducted    bool      `json:"lessonDeducted"`
	DeductLessonCount float64   `json:"deductLessonCount"`
	RecordedAt        time.Time `json:"recordedAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type LessonRecordListResponse []LessonRecordResponse
