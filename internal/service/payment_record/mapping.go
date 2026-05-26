package payment_record

import (
	"fmt"
	"strings"

	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
)

func toPaymentRecordResponses(items []*model.PaymentRecord, studentMap map[int32]*model.Student, lessonPackageMap map[int32]*model.LessonPackage) dto.PaymentRecordListResponse {
	list := make(dto.PaymentRecordListResponse, 0, len(items))
	for _, item := range items {
		if resp := toPaymentRecordResponse(item, studentMap, lessonPackageMap); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toPaymentRecordResponse(item *model.PaymentRecord, studentMap map[int32]*model.Student, lessonPackageMap map[int32]*model.LessonPackage) *dto.PaymentRecordResponse {
	if item == nil {
		return nil
	}
	studentName := ""
	lessonPackageName := ""
	courseName := ""
	if student := studentMap[item.StudentID]; student != nil {
		studentName = student.StudentName
	}
	if lessonPackage := lessonPackageMap[item.LessonPackageID]; lessonPackage != nil {
		if lessonPackage.CourseID > 0 {
			courseName = fmt.Sprintf("课程%d", lessonPackage.CourseID)
		}
		parts := make([]string, 0, 4)
		if studentName != "" {
			parts = append(parts, studentName)
		}
		if courseName != "" {
			parts = append(parts, courseName)
		}
		parts = append(parts, fmt.Sprintf("课时包#%d", lessonPackage.ID))
		if remain := serviceutil.DecimalToFloat(lessonPackage.RemainLessons); remain > 0 {
			parts = append(parts, fmt.Sprintf("剩余%.0f课时", remain))
		}
		lessonPackageName = strings.Join(parts, " · ")
	}
	return &dto.PaymentRecordResponse{ID: item.ID, StudentID: item.StudentID, StudentName: studentName, LessonPackageID: item.LessonPackageID, LessonPackageName: lessonPackageName, CourseName: courseName, PaymentType: item.PaymentType, Amount: serviceutil.DecimalToFloat(item.Amount), PaymentMethod: item.PaymentMethod, PaymentTime: item.PaymentTime, PaymentStatus: item.PaymentStatus, Remark: item.Remark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
