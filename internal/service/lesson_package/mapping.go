package lesson_package

import (
	"time"

	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	"github.com/shopspring/decimal"
)

func toLessonPackageResponses(items []*model.LessonPackage, studentMap map[int32]*model.Student, courseMap map[int32]*model.Course) dto.LessonPackageListResponse {
	list := make(dto.LessonPackageListResponse, 0, len(items))
	for _, item := range items {
		if resp := toLessonPackageResponse(item, studentMap, courseMap); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toLessonPackageResponse(item *model.LessonPackage, studentMap map[int32]*model.Student, courseMap map[int32]*model.Course) *dto.LessonPackageResponse {
	if item == nil {
		return nil
	}
	arrears := item.TotalAmount.Sub(item.PaidAmount)
	if arrears.IsNegative() {
		arrears = decimal.Zero
	}
	studentName := ""
	courseName := ""
	subject := ""
	if student := studentMap[item.StudentID]; student != nil {
		studentName = student.StudentName
		if subject == "" {
			subject = student.Subject
		}
	}
	if course := courseMap[item.CourseID]; course != nil {
		courseName = course.CourseName
		if course.Subject != "" {
			subject = course.Subject
		}
	}
	return &dto.LessonPackageResponse{
		ID: item.ID, StudentID: item.StudentID, StudentName: studentName, CourseID: item.CourseID, CourseName: courseName, Subject: subject,
		TotalLessons: serviceutil.DecimalToFloat(item.TotalLessons), UsedLessons: serviceutil.DecimalToFloat(item.UsedLessons), RemainLessons: serviceutil.DecimalToFloat(item.RemainLessons),
		TotalAmount: serviceutil.DecimalToFloat(item.TotalAmount), PaidAmount: serviceutil.DecimalToFloat(item.PaidAmount),
		StartDate: formatDate(item.StartDate), EndDate: formatDate(item.EndDate),
		Status: item.Status, LowLessonThreshold: serviceutil.DecimalToFloat(item.LowLessonThreshold), Remark: item.Remark,
		LowLessonAlert: item.RemainLessons.LessThanOrEqual(item.LowLessonThreshold),
		ArrearsAmount:  serviceutil.DecimalToFloat(arrears), CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func formatDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}
