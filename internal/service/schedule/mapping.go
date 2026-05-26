package schedule

import (
	"fmt"
	"strings"

	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
)

func toScheduleResponses(items []*model.Schedule, studentMap map[int32]*model.Student, courseMap map[int32]*model.Course, teacherMap map[int32]*model.SysUser) dto.ScheduleListResponse {
	list := make(dto.ScheduleListResponse, 0, len(items))
	for _, item := range items {
		if resp := toScheduleResponse(item, studentMap, courseMap, teacherMap); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toScheduleResponse(item *model.Schedule, studentMap map[int32]*model.Student, courseMap map[int32]*model.Course, teacherMap map[int32]*model.SysUser) *dto.ScheduleResponse {
	if item == nil {
		return nil
	}
	studentName := ""
	courseName := ""
	teacherName := ""
	lessonPackageName := ""
	if student := studentMap[item.StudentID]; student != nil {
		studentName = student.StudentName
	}
	if course := courseMap[item.CourseID]; course != nil {
		courseName = course.CourseName
	}
	if teacher := teacherMap[item.TeacherID]; teacher != nil {
		teacherName = teacher.RealName
		if teacherName == "" {
			teacherName = teacher.Username
		}
	}
	lessonPackageID := serviceutil.DecimalPtrToInt32Value(item.LessonPackageID)
	if lessonPackageID > 0 {
		parts := make([]string, 0, 3)
		if studentName != "" {
			parts = append(parts, studentName)
		}
		if courseName != "" {
			parts = append(parts, courseName)
		}
		parts = append(parts, fmt.Sprintf("课时包#%d", lessonPackageID))
		lessonPackageName = strings.Join(parts, " · ")
	}
	return &dto.ScheduleResponse{ID: item.ID, StudentID: item.StudentID, StudentName: studentName, CourseID: item.CourseID, CourseName: courseName, TeacherID: item.TeacherID, TeacherName: teacherName, LessonPackageID: lessonPackageID, LessonPackageName: lessonPackageName, ClassDate: item.ClassDate.Format("2006-01-02"), StartTime: item.StartTime, EndTime: item.EndTime, Classroom: item.Classroom, ScheduleStatus: item.ScheduleStatus, OriginalScheduleID: serviceutil.DecimalPtrToInt32Value(item.OriginalScheduleID), IsMakeup: item.IsMakeup, Remark: item.Remark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
