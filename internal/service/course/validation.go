package course

import (
	"strings"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
)

func validateCourseCreateRequest(req *dto.CourseCreateRequest) error {
	if req == nil {
		return apperr.InvalidArgument("course create request is required")
	}
	if strings.TrimSpace(req.CourseName) == "" {
		return apperr.InvalidArgument("courseName is required")
	}
	if strings.TrimSpace(req.Subject) == "" {
		return apperr.InvalidArgument("subject is required")
	}
	if strings.TrimSpace(req.CourseType) == "" {
		return apperr.InvalidArgument("courseType is required")
	}
	if req.DurationMinutes < 0 {
		return apperr.InvalidArgument("durationMinutes must not be negative")
	}
	if req.FeeStandard < 0 {
		return apperr.InvalidArgument("feeStandard must not be negative")
	}
	if _, ok := courseStatuses[normalizeCourseStatus(req.Status)]; !ok {
		return apperr.InvalidArgument("status is invalid")
	}
	return nil
}

func sanitizeCourseUpdates(req dto.CourseUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("course update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "courseName":
			v, err := requiredString(field, value)
			if err != nil {
				return nil, err
			}
			updates["course_name"] = v
		case "subject":
			v, err := requiredString(field, value)
			if err != nil {
				return nil, err
			}
			updates["subject"] = v
		case "courseType":
			v, err := requiredString(field, value)
			if err != nil {
				return nil, err
			}
			updates["course_type"] = v
		case "durationMinutes":
			v, err := int32Value(field, value, false)
			if err != nil {
				return nil, err
			}
			if v <= 0 {
				return nil, apperr.InvalidArgument("durationMinutes must be positive")
			}
			updates["duration_minutes"] = v
		case "feeStandard":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, err
			}
			updates["fee_standard"] = v
		case "remark":
			v, err := optionalString(field, value)
			if err != nil {
				return nil, err
			}
			updates["remark"] = v
		case "status":
			v, err := optionalString(field, value)
			if err != nil {
				return nil, err
			}
			v = normalizeCourseStatus(v)
			if _, ok := courseStatuses[v]; !ok {
				return nil, apperr.InvalidArgument("status is invalid")
			}
			updates["status"] = v
		default:
			return nil, apperr.InvalidArgument("course update field " + field + " is not allowed")
		}
	}
	return updates, nil
}

func requiredString(field string, value interface{}) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", apperr.InvalidArgument(field + " must be a string")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", apperr.InvalidArgument(field + " is required")
	}
	return text, nil
}

func optionalString(field string, value interface{}) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", apperr.InvalidArgument(field + " must be a string")
	}
	return strings.TrimSpace(text), nil
}

func int32Value(field string, value interface{}, positive bool) (int32, error) {
	number, ok := value.(float64)
	if !ok {
		return 0, apperr.InvalidArgument(field + " must be a number")
	}
	result := int32(number)
	if float64(result) != number {
		return 0, apperr.InvalidArgument(field + " must be an integer")
	}
	if positive && result <= 0 {
		return 0, apperr.InvalidArgument(field + " must be positive")
	}
	return result, nil
}

func nonNegativeFloat(field string, value interface{}) (float64, error) {
	number, ok := value.(float64)
	if !ok {
		return 0, apperr.InvalidArgument(field + " must be a number")
	}
	if number < 0 {
		return 0, apperr.InvalidArgument(field + " must not be negative")
	}
	return number, nil
}

func normalizeCourseStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "enabled"
	}
	return status
}

func defaultCourseDuration(v int32) int32 {
	if v <= 0 {
		return 60
	}
	return v
}
