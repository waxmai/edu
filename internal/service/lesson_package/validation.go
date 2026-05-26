package lesson_package

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	"github.com/shopspring/decimal"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func buildLessonPackageModel(req *dto.LessonPackageCreateRequest) (*model.LessonPackage, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("lesson package create request is required")
	}
	if req.StudentID <= 0 {
		return nil, apperr.InvalidArgument("studentId must be positive")
	}
	if req.CourseID <= 0 {
		return nil, apperr.InvalidArgument("courseId must be positive")
	}
	if req.TotalLessons <= 0 {
		return nil, apperr.InvalidArgument("totalLessons must be greater than 0")
	}
	if req.UsedLessons < 0 || req.PaidAmount < 0 || req.TotalAmount < 0 {
		return nil, apperr.InvalidArgument("amount and used lessons must not be negative")
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return nil, apperr.InvalidArgument("startDate must use format 2006-01-02")
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return nil, apperr.InvalidArgument("endDate must use format 2006-01-02")
	}
	if !startDate.IsZero() && !endDate.IsZero() && endDate.Before(startDate) {
		return nil, apperr.InvalidArgument("endDate must be on or after startDate")
	}
	status := normalizeLessonPackageStatus(req.Status)
	remainLessons := req.RemainLessons
	if remainLessons == 0 {
		remainLessons = req.TotalLessons - req.UsedLessons
	}
	if remainLessons < 0 {
		return nil, apperr.InvalidArgument("remainLessons must not be negative")
	}
	if req.UsedLessons > req.TotalLessons {
		return nil, apperr.InvalidArgument("usedLessons cannot exceed totalLessons")
	}
	threshold := req.LowLessonThreshold
	if threshold <= 0 {
		threshold = 3
	}
	status = deriveLessonPackageStatus(status, serviceutil.DecimalFromFloat(remainLessons), serviceutil.TimePtr(endDate))
	return &model.LessonPackage{
		OrganizationID:     req.OrganizationID,
		CampusID:           req.CampusID,
		StudentID:          req.StudentID,
		CourseID:           req.CourseID,
		TotalLessons:       serviceutil.DecimalFromFloat(req.TotalLessons),
		UsedLessons:        serviceutil.DecimalFromFloat(req.UsedLessons),
		RemainLessons:      serviceutil.DecimalFromFloat(remainLessons),
		TotalAmount:        serviceutil.DecimalFromFloat(req.TotalAmount),
		PaidAmount:         serviceutil.DecimalFromFloat(req.PaidAmount),
		StartDate:          serviceutil.TimePtr(startDate),
		EndDate:            serviceutil.TimePtr(endDate),
		Status:             status,
		LowLessonThreshold: serviceutil.DecimalFromFloat(threshold),
		Remark:             strings.TrimSpace(req.Remark),
	}, nil
}

func sanitizeLessonPackageUpdates(req dto.LessonPackageUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("lesson package update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "courseId":
			v, err := int32Value(field, value, true)
			if err != nil {
				return nil, err
			}
			updates["course_id"] = v
		case "totalLessons":
			v, err := positiveFloat(field, value, true)
			if err != nil {
				return nil, err
			}
			updates["total_lessons"] = serviceutil.DecimalFromFloat(v)
		case "usedLessons":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, err
			}
			updates["used_lessons"] = serviceutil.DecimalFromFloat(v)
		case "remainLessons":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, err
			}
			updates["remain_lessons"] = serviceutil.DecimalFromFloat(v)
		case "totalAmount":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, err
			}
			updates["total_amount"] = serviceutil.DecimalFromFloat(v)
		case "paidAmount":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, err
			}
			updates["paid_amount"] = serviceutil.DecimalFromFloat(v)
		case "startDate":
			v, err := dateStringValue(field, value)
			if err != nil {
				return nil, err
			}
			parsed, _ := parseDate(v)
			updates["start_date"] = serviceutil.TimePtr(parsed)
		case "endDate":
			v, err := dateStringValue(field, value)
			if err != nil {
				return nil, err
			}
			parsed, _ := parseDate(v)
			updates["end_date"] = serviceutil.TimePtr(parsed)
		case "status":
			v, err := stringValue(field, value, false)
			if err != nil {
				return nil, err
			}
			status := normalizeLessonPackageStatus(v)
			if _, ok := lessonPackageStatuses[status]; !ok {
				return nil, apperr.InvalidArgument("status is invalid")
			}
			updates["status"] = status
		case "lowLessonThreshold":
			v, err := positiveFloat(field, value, false)
			if err != nil {
				return nil, err
			}
			updates["low_lesson_threshold"] = serviceutil.DecimalFromFloat(v)
		case "remark":
			v, err := stringValue(field, value, false)
			if err != nil {
				return nil, err
			}
			updates["remark"] = v
		default:
			return nil, apperr.InvalidArgument("lesson package update field " + field + " is not allowed")
		}
	}
	return updates, nil
}

func mergeLessonPackageUpdates(item *model.LessonPackage, updates map[string]interface{}) (map[string]interface{}, error) {
	totalLessons := item.TotalLessons
	usedLessons := item.UsedLessons
	remainLessons := item.RemainLessons
	status := item.Status
	startDate := item.StartDate
	endDate := item.EndDate

	if v, ok := updates["total_lessons"].(decimal.Decimal); ok {
		totalLessons = v
	}
	if v, ok := updates["used_lessons"].(decimal.Decimal); ok {
		usedLessons = v
	}
	if v, ok := updates["remain_lessons"].(decimal.Decimal); ok {
		remainLessons = v
	}
	if v, ok := updates["status"].(string); ok {
		status = v
	}
	if v, ok := updates["start_date"].(*time.Time); ok {
		startDate = v
	}
	if v, ok := updates["end_date"].(*time.Time); ok {
		endDate = v
	}
	if _, changedTotal := updates["total_lessons"]; remainLessons.Equal(item.RemainLessons) && (changedTotal || updates["used_lessons"] != nil) {
		remainLessons = totalLessons.Sub(usedLessons)
		updates["remain_lessons"] = remainLessons
	}
	if usedLessons.GreaterThan(totalLessons) {
		return nil, apperr.InvalidArgument("usedLessons cannot exceed totalLessons")
	}
	if remainLessons.IsNegative() {
		return nil, apperr.InvalidArgument("remainLessons must not be negative")
	}
	if !serviceutil.TimeValue(startDate).IsZero() && !serviceutil.TimeValue(endDate).IsZero() && serviceutil.TimeValue(endDate).Before(serviceutil.TimeValue(startDate)) {
		return nil, apperr.InvalidArgument("endDate must be on or after startDate")
	}
	updates["status"] = deriveLessonPackageStatus(status, remainLessons, endDate)
	return updates, nil
}

func ensureStudentExists(ctx context.Context, q *dao.Query, studentID int32) (*model.Student, error) {
	student, err := q.WithContext(ctx).Student.Where(q.Student.ID.Eq(studentID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("student not found")
		}
		return nil, err
	}
	return student, nil
}

func validateLessonPackageListQuery(query dto.LessonPackageListQuery) error {
	if query.StudentID < 0 {
		return apperr.InvalidArgument("studentId must not be negative")
	}
	if query.CourseID < 0 {
		return apperr.InvalidArgument("courseId must not be negative")
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		if _, ok := lessonPackageStatuses[status]; !ok {
			return apperr.InvalidArgument("status is invalid")
		}
	}
	if paymentStatus := strings.TrimSpace(query.PaymentStatus); paymentStatus != "" && paymentStatus != "arrears" {
		return apperr.InvalidArgument("paymentStatus is invalid")
	}
	if query.PageNum < 0 {
		return apperr.InvalidArgument("pageNum must not be negative")
	}
	if query.PageSize < 0 {
		return apperr.InvalidArgument("pageSize must not be negative")
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

func buildLessonPackageConditions(q *dao.Query, query dto.LessonPackageListQuery) []gen.Condition {
	conditions := make([]gen.Condition, 0, 3)
	if query.StudentID > 0 {
		conditions = append(conditions, q.LessonPackage.StudentID.Eq(query.StudentID))
	}
	if query.CourseID > 0 {
		conditions = append(conditions, q.LessonPackage.CourseID.Eq(query.CourseID))
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		conditions = append(conditions, q.LessonPackage.Status.Eq(status))
	}
	if query.LowLessonAlert {
		conditions = append(conditions, q.LessonPackage.RemainLessons.LteCol(q.LessonPackage.LowLessonThreshold))
	}
	if strings.TrimSpace(query.PaymentStatus) == "arrears" {
		conditions = append(conditions, q.LessonPackage.PaidAmount.LtCol(q.LessonPackage.TotalAmount))
	}
	return conditions
}

func normalizeLessonPackageStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "pending"
	}
	return status
}

func deriveLessonPackageStatus(status string, remainLessons decimal.Decimal, endDate *time.Time) string {
	if status == "closed" {
		return status
	}
	if endDate != nil && !endDate.IsZero() && endDate.Before(time.Now().Truncate(24*time.Hour)) {
		return "expired"
	}
	if remainLessons.LessThanOrEqual(decimal.Zero) {
		return "exhausted"
	}
	if status == "pending" || status == "active" {
		return status
	}
	return status
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

func stringValue(field string, value interface{}, required bool) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", apperr.InvalidArgument(field + " must be a string")
	}
	text = strings.TrimSpace(text)
	if required && text == "" {
		return "", apperr.InvalidArgument(field + " is required")
	}
	return text, nil
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

func positiveFloat(field string, value interface{}, allowZero bool) (float64, error) {
	number, ok := value.(float64)
	if !ok {
		return 0, apperr.InvalidArgument(field + " must be a number")
	}
	if allowZero {
		if number < 0 {
			return 0, apperr.InvalidArgument(field + " must not be negative")
		}
	} else if number <= 0 {
		return 0, apperr.InvalidArgument(field + " must be greater than 0")
	}
	return number, nil
}

func dateStringValue(field string, value interface{}) (string, error) {
	text, err := stringValue(field, value, false)
	if err != nil {
		return "", err
	}
	if _, err := parseDate(text); err != nil {
		return "", apperr.InvalidArgument(field + " must use format 2006-01-02")
	}
	return text, nil
}
