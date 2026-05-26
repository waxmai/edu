package dto

import "time"

type ScheduleCreateRequest struct {
	OrganizationID     int32  `json:"-"`
	CampusID           int32  `json:"-"`
	StudentID          int32  `json:"studentId" binding:"required"`
	CourseID           int32  `json:"courseId" binding:"required"`
	TeacherID          int32  `json:"teacherId" binding:"required"`
	LessonPackageID    int32  `json:"lessonPackageId"`
	ClassDate          string `json:"classDate" binding:"required"`
	StartTime          string `json:"startTime" binding:"required"`
	EndTime            string `json:"endTime" binding:"required"`
	Classroom          string `json:"classroom"`
	ScheduleStatus     string `json:"scheduleStatus"`
	OriginalScheduleID int32  `json:"originalScheduleId"`
	IsMakeup           bool   `json:"isMakeup"`
	Remark             string `json:"remark"`
}

type ScheduleUpdateRequest map[string]interface{}

type ScheduleListQuery struct {
	StudentID       int32  `form:"studentId" json:"studentId"`
	TeacherID       int32  `form:"teacherId" json:"teacherId"`
	LessonPackageID int32  `form:"lessonPackageId" json:"lessonPackageId"`
	ScheduleStatus  string `form:"scheduleStatus" json:"scheduleStatus"`
	StartDate       string `form:"startDate" json:"startDate"`
	EndDate         string `form:"endDate" json:"endDate"`
	PageNum         int    `form:"pageNum" json:"pageNum"`
	PageSize        int    `form:"pageSize" json:"pageSize"`
}

type ScheduleCancelRequest struct {
	Reason string `json:"reason"`
}

type ScheduleLeaveRequest struct {
	Reason string `json:"reason"`
}

type ScheduleRescheduleRequest struct {
	Reason       string `json:"reason"`
	NewClassDate string `json:"newClassDate" binding:"required"`
	NewStartTime string `json:"newStartTime" binding:"required"`
	NewEndTime   string `json:"newEndTime" binding:"required"`
	Classroom    string `json:"classroom"`
}

type MakeupScheduleCreateRequest struct {
	OriginalScheduleID int32  `json:"originalScheduleId" binding:"required"`
	StudentID          int32  `json:"studentId" binding:"required"`
	TeacherID          int32  `json:"teacherId" binding:"required"`
	CourseID           int32  `json:"courseId" binding:"required"`
	LessonPackageID    int32  `json:"lessonPackageId"`
	ClassDate          string `json:"classDate" binding:"required"`
	StartTime          string `json:"startTime" binding:"required"`
	EndTime            string `json:"endTime" binding:"required"`
	Classroom          string `json:"classroom"`
	Remark             string `json:"remark"`
}

type ScheduleResponse struct {
	ID                 int32     `json:"id"`
	StudentID          int32     `json:"studentId"`
	StudentName        string    `json:"studentName"`
	CourseID           int32     `json:"courseId"`
	CourseName         string    `json:"courseName"`
	TeacherID          int32     `json:"teacherId"`
	TeacherName        string    `json:"teacherName"`
	LessonPackageID    int32     `json:"lessonPackageId"`
	LessonPackageName  string    `json:"lessonPackageName"`
	ClassDate          string    `json:"classDate"`
	StartTime          time.Time `json:"startTime"`
	EndTime            time.Time `json:"endTime"`
	Classroom          string    `json:"classroom"`
	ScheduleStatus     string    `json:"scheduleStatus"`
	OriginalScheduleID int32     `json:"originalScheduleId"`
	IsMakeup           bool      `json:"isMakeup"`
	Remark             string    `json:"remark"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type ScheduleListResponse []ScheduleResponse

type ScheduleConflictResponse struct {
	TeacherConflict bool `json:"teacherConflict"`
	StudentConflict bool `json:"studentConflict"`
}
