package student

import (
	"strings"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
)

func validateStudentCreateRequest(req *dto.StudentCreateRequest) error {
	if req == nil {
		return apperr.InvalidArgument("student create request is required")
	}
	if strings.TrimSpace(req.StudentName) == "" {
		return apperr.InvalidArgument("studentName is required")
	}
	if strings.TrimSpace(req.Subject) == "" {
		return apperr.InvalidArgument("subject is required")
	}
	if strings.TrimSpace(req.TeachingType) == "" {
		return apperr.InvalidArgument("teachingType is required")
	}
	return nil
}

func sanitizeStudentUpdates(req dto.StudentUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("student update fields are required")
	}

	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "studentName":
			text, err := sanitizeRequiredStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["student_name"] = text
		case "subject":
			text, err := sanitizeRequiredStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["subject"] = text
		case "teachingType":
			text, err := sanitizeRequiredStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["teaching_type"] = text
		case "gender":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["gender"] = text
		case "grade":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["grade"] = text
		case "phone":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["phone"] = text
		case "parentName":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["parent_name"] = text
		case "parentPhone":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["parent_phone"] = text
		case "status":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["status"] = text
		case "remark":
			text, err := sanitizeOptionalStringField(field, value)
			if err != nil {
				return nil, err
			}
			updates["remark"] = text
		default:
			return nil, apperr.InvalidArgument("student update field " + field + " is not allowed")
		}
	}

	return updates, nil
}

func validateStudentListQuery(query dto.StudentListQuery) error {
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
	return nil
}

func sanitizeRequiredStringField(field string, value interface{}) (string, error) {
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

func sanitizeOptionalStringField(field string, value interface{}) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", apperr.InvalidArgument(field + " must be a string")
	}
	return strings.TrimSpace(text), nil
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
