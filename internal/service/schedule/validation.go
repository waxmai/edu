package schedule

import (
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
)

func buildScheduleModel(req *dto.ScheduleCreateRequest) (*model.Schedule, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("schedule create request is required")
	}
	if req.StudentID <= 0 {
		return nil, apperr.InvalidArgument("studentId must be positive")
	}
	if req.CourseID <= 0 {
		return nil, apperr.InvalidArgument("courseId must be positive")
	}
	if req.TeacherID <= 0 {
		return nil, apperr.InvalidArgument("teacherId must be positive")
	}
	classDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.ClassDate))
	if err != nil {
		return nil, apperr.InvalidArgument("classDate must use format 2006-01-02")
	}
	startTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(req.StartTime))
	if err != nil {
		return nil, apperr.InvalidArgument("startTime must use format 2006-01-02 15:04:05")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(req.EndTime))
	if err != nil {
		return nil, apperr.InvalidArgument("endTime must use format 2006-01-02 15:04:05")
	}
	if !startTime.Before(endTime) {
		return nil, apperr.InvalidArgument("startTime must be earlier than endTime")
	}
	status := strings.TrimSpace(req.ScheduleStatus)
	if status == "" {
		status = "scheduled"
	}
	if _, ok := scheduleStatuses[status]; !ok {
		return nil, apperr.InvalidArgument("scheduleStatus is invalid")
	}
	return &model.Schedule{OrganizationID: req.OrganizationID, CampusID: req.CampusID, StudentID: req.StudentID, CourseID: req.CourseID, TeacherID: req.TeacherID, LessonPackageID: serviceutil.NilIfNonPositiveInt32(req.LessonPackageID), ClassDate: classDate, StartTime: startTime, EndTime: endTime, Classroom: strings.TrimSpace(req.Classroom), ScheduleStatus: status, OriginalScheduleID: serviceutil.NilIfNonPositiveInt32(req.OriginalScheduleID), IsMakeup: req.IsMakeup, Remark: strings.TrimSpace(req.Remark)}, nil
}

func sanitizeScheduleUpdates(req dto.ScheduleUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("schedule update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "teacherId", "studentId", "courseId":
			n, ok := value.(float64)
			if !ok {
				return nil, apperr.InvalidArgument(field + " must be a number")
			}
			if n < 0 {
				return nil, apperr.InvalidArgument(field + " must not be negative")
			}
			key := map[string]string{"teacherId": "teacher_id", "studentId": "student_id", "courseId": "course_id"}[field]
			updates[key] = int32(n)
		case "lessonPackageId", "originalScheduleId":
			n, ok := value.(float64)
			if !ok {
				return nil, apperr.InvalidArgument(field + " must be a number")
			}
			if n < 0 {
				return nil, apperr.InvalidArgument(field + " must not be negative")
			}
			key := map[string]string{"lessonPackageId": "lesson_package_id", "originalScheduleId": "original_schedule_id"}[field]
			updates[key] = serviceutil.NilIfNonPositiveInt32(int32(n))
		case "classDate":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("classDate must be a string")
			}
			parsed, err := time.Parse("2006-01-02", strings.TrimSpace(text))
			if err != nil {
				return nil, apperr.InvalidArgument("classDate must use format 2006-01-02")
			}
			updates["class_date"] = parsed
		case "startTime", "endTime":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument(field + " must be a string")
			}
			parsed, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(text))
			if err != nil {
				return nil, apperr.InvalidArgument(field + " must use format 2006-01-02 15:04:05")
			}
			key := map[string]string{"startTime": "start_time", "endTime": "end_time"}[field]
			updates[key] = parsed
		case "classroom", "remark":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument(field + " must be a string")
			}
			key := map[string]string{"classroom": "classroom", "remark": "remark"}[field]
			updates[key] = strings.TrimSpace(text)
		case "scheduleStatus":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("scheduleStatus must be a string")
			}
			text = strings.TrimSpace(text)
			if _, ok := scheduleStatuses[text]; !ok {
				return nil, apperr.InvalidArgument("scheduleStatus is invalid")
			}
			updates["schedule_status"] = text
		case "isMakeup":
			flag, ok := value.(bool)
			if !ok {
				return nil, apperr.InvalidArgument("isMakeup must be a boolean")
			}
			updates["is_makeup"] = flag
		default:
			return nil, apperr.InvalidArgument("schedule update field " + field + " is not allowed")
		}
	}
	return updates, nil
}

func mergeScheduleUpdates(item *model.Schedule, updates map[string]interface{}) (*model.Schedule, error) {
	copyItem := *item
	if v, ok := updates["student_id"].(int32); ok {
		copyItem.StudentID = v
	}
	if v, ok := updates["course_id"].(int32); ok {
		copyItem.CourseID = v
	}
	if v, ok := updates["teacher_id"].(int32); ok {
		copyItem.TeacherID = v
	}
	if v, ok := updates["lesson_package_id"].(*int32); ok {
		copyItem.LessonPackageID = v
	}
	if v, ok := updates["class_date"].(time.Time); ok {
		copyItem.ClassDate = v
	}
	if v, ok := updates["start_time"].(time.Time); ok {
		copyItem.StartTime = v
	}
	if v, ok := updates["end_time"].(time.Time); ok {
		copyItem.EndTime = v
	}
	if v, ok := updates["classroom"].(string); ok {
		copyItem.Classroom = v
	}
	if v, ok := updates["schedule_status"].(string); ok {
		copyItem.ScheduleStatus = v
	}
	if v, ok := updates["original_schedule_id"].(*int32); ok {
		copyItem.OriginalScheduleID = v
	}
	if v, ok := updates["is_makeup"].(bool); ok {
		copyItem.IsMakeup = v
	}
	if v, ok := updates["remark"].(string); ok {
		copyItem.Remark = v
	}
	if !copyItem.StartTime.Before(copyItem.EndTime) {
		return nil, apperr.InvalidArgument("startTime must be earlier than endTime")
	}
	return &copyItem, nil
}

func validateScheduleListQuery(query dto.ScheduleListQuery) error {
	if query.StudentID < 0 || query.TeacherID < 0 || query.LessonPackageID < 0 {
		return apperr.InvalidArgument("filter ids must not be negative")
	}
	if status := strings.TrimSpace(query.ScheduleStatus); status != "" {
		if _, ok := scheduleStatuses[status]; !ok {
			return apperr.InvalidArgument("scheduleStatus is invalid")
		}
	}
	if query.PageNum < 0 || query.PageSize < 0 {
		return apperr.InvalidArgument("pageNum and pageSize must not be negative")
	}
	if query.PageNum > 0 {
		if query.PageSize <= 0 {
			query.PageSize = 20
		}
		if query.PageSize > 100 {
			return apperr.InvalidArgument("pageSize must not exceed 100")
		}
	}
	_, _, err := parseScheduleDateRange(query.StartDate, query.EndDate)
	return err
}

func parseScheduleDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error
	if strings.TrimSpace(startDate) != "" {
		start, err = time.Parse("2006-01-02", strings.TrimSpace(startDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must use format 2006-01-02")
		}
	}
	if strings.TrimSpace(endDate) != "" {
		end, err = time.Parse("2006-01-02", strings.TrimSpace(endDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("endDate must use format 2006-01-02")
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must be earlier than or equal to endDate")
	}
	return start, end, nil
}

func pointerInt32Equal(a, b *int32) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func reqReason(reason string) string { return strings.TrimSpace(reason) }

func appendReason(action, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return action
	}
	return action + ": " + reason
}
