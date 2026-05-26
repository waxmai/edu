package payment_record

import (
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	"github.com/shopspring/decimal"
	"gorm.io/gen"
)

func buildPaymentRecordModel(req *dto.PaymentRecordCreateRequest) (*model.PaymentRecord, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("payment record create request is required")
	}
	if req.StudentID <= 0 {
		return nil, apperr.InvalidArgument("studentId must be positive")
	}
	if req.LessonPackageID <= 0 {
		return nil, apperr.InvalidArgument("lessonPackageId must be positive")
	}
	if req.Amount <= 0 {
		return nil, apperr.InvalidArgument("amount must be greater than 0")
	}
	paymentType := strings.TrimSpace(req.PaymentType)
	if _, ok := paymentTypes[paymentType]; !ok {
		return nil, apperr.InvalidArgument("paymentType is invalid")
	}
	paymentMethod := strings.TrimSpace(req.PaymentMethod)
	if paymentMethod == "" {
		return nil, apperr.InvalidArgument("paymentMethod is required")
	}
	paymentTime, err := time.Parse(paymentTimeLayout, strings.TrimSpace(req.PaymentTime))
	if err != nil {
		return nil, apperr.InvalidArgument("paymentTime must use format 2006-01-02 15:04:05")
	}
	paymentStatus := strings.TrimSpace(req.PaymentStatus)
	if paymentStatus == "" {
		paymentStatus = "paid"
	}
	if _, ok := paymentStatuses[paymentStatus]; !ok {
		return nil, apperr.InvalidArgument("paymentStatus is invalid")
	}
	return &model.PaymentRecord{OrganizationID: req.OrganizationID, CampusID: req.CampusID, StudentID: req.StudentID, LessonPackageID: req.LessonPackageID, PaymentType: paymentType, Amount: serviceutil.DecimalFromFloat(req.Amount), PaymentMethod: paymentMethod, PaymentTime: paymentTime, PaymentStatus: paymentStatus, Remark: strings.TrimSpace(req.Remark)}, nil
}

func sanitizePaymentRecordUpdates(req dto.PaymentRecordUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("payment record update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "amount":
			number, ok := value.(float64)
			if !ok || number <= 0 {
				return nil, apperr.InvalidArgument("amount must be greater than 0")
			}
			updates["amount"] = serviceutil.DecimalFromFloat(number)
		case "paymentMethod":
			text, ok := value.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return nil, apperr.InvalidArgument("paymentMethod is required")
			}
			updates["payment_method"] = strings.TrimSpace(text)
		case "paymentTime":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("paymentTime must be a string")
			}
			parsed, err := time.Parse(paymentTimeLayout, strings.TrimSpace(text))
			if err != nil {
				return nil, apperr.InvalidArgument("paymentTime must use format 2006-01-02 15:04:05")
			}
			updates["payment_time"] = parsed
		case "paymentStatus":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("paymentStatus must be a string")
			}
			text = strings.TrimSpace(text)
			if _, ok := paymentStatuses[text]; !ok {
				return nil, apperr.InvalidArgument("paymentStatus is invalid")
			}
			updates["payment_status"] = text
		case "remark":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("remark must be a string")
			}
			updates["remark"] = strings.TrimSpace(text)
		default:
			return nil, apperr.InvalidArgument("payment record update field " + field + " is not allowed")
		}
	}
	return updates, nil
}

func validatePaymentRecordListQuery(query dto.PaymentRecordListQuery) error {
	if query.StudentID < 0 {
		return apperr.InvalidArgument("studentId must not be negative")
	}
	if query.LessonPackageID < 0 {
		return apperr.InvalidArgument("lessonPackageId must not be negative")
	}
	if paymentType := strings.TrimSpace(query.PaymentType); paymentType != "" {
		if _, ok := paymentTypes[paymentType]; !ok {
			return apperr.InvalidArgument("paymentType is invalid")
		}
	}
	if paymentStatus := strings.TrimSpace(query.PaymentStatus); paymentStatus != "" {
		if _, ok := paymentStatuses[paymentStatus]; !ok {
			return apperr.InvalidArgument("paymentStatus is invalid")
		}
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
	_, _, err := parseDateRange(query.StartDate, query.EndDate)
	return err
}

func normalizePaymentStatsQuery(query dto.PaymentIncomeStatsQuery) (time.Time, time.Time, string, error) {
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	groupBy := strings.TrimSpace(query.GroupBy)
	if groupBy == "" {
		groupBy = "day"
	}
	switch groupBy {
	case "day", "week", "month":
		return start, end, groupBy, nil
	default:
		return time.Time{}, time.Time{}, "", apperr.InvalidArgument("groupBy is invalid")
	}
}

func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	var start time.Time
	var end time.Time
	var err error
	if strings.TrimSpace(startDate) != "" {
		start, err = time.Parse(paymentDateLayout, strings.TrimSpace(startDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must use format 2006-01-02")
		}
	}
	if strings.TrimSpace(endDate) != "" {
		end, err = time.Parse(paymentDateLayout, strings.TrimSpace(endDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("endDate must use format 2006-01-02")
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must be earlier than or equal to endDate")
	}
	return start, end, nil
}

func buildPaymentRecordConditions(q *dao.Query, query dto.PaymentRecordListQuery) []gen.Condition {
	conditions := make([]gen.Condition, 0, 7)
	if query.StudentID > 0 {
		conditions = append(conditions, q.PaymentRecord.StudentID.Eq(query.StudentID))
	}
	if query.LessonPackageID > 0 {
		conditions = append(conditions, q.PaymentRecord.LessonPackageID.Eq(query.LessonPackageID))
	}
	if paymentType := strings.TrimSpace(query.PaymentType); paymentType != "" {
		conditions = append(conditions, q.PaymentRecord.PaymentType.Eq(paymentType))
	}
	if paymentStatus := strings.TrimSpace(query.PaymentStatus); paymentStatus != "" {
		conditions = append(conditions, q.PaymentRecord.PaymentStatus.Eq(paymentStatus))
	}
	if paymentMethod := strings.TrimSpace(query.PaymentMethod); paymentMethod != "" {
		conditions = append(conditions, q.PaymentRecord.PaymentMethod.Eq(paymentMethod))
	}
	if start, end, err := parseDateRange(query.StartDate, query.EndDate); err == nil {
		if !start.IsZero() {
			conditions = append(conditions, q.PaymentRecord.PaymentTime.Gte(start))
		}
		if !end.IsZero() {
			conditions = append(conditions, q.PaymentRecord.PaymentTime.Lt(end.Add(24*time.Hour)))
		}
	}
	return conditions
}

func derivePackageStatusFromPayment(current string, totalAmount, paidAmount, remainLessons decimal.Decimal, endDate *time.Time) string {
	if current == "closed" {
		return current
	}
	if endDate != nil && !endDate.IsZero() && endDate.Before(time.Now().Truncate(24*time.Hour)) {
		return "expired"
	}
	if remainLessons.LessThanOrEqual(decimal.Zero) {
		return "exhausted"
	}
	if paidAmount.GreaterThan(decimal.Zero) && totalAmount.GreaterThanOrEqual(paidAmount) {
		return "active"
	}
	return current
}

func paidAmountDelta(amount decimal.Decimal, status string) decimal.Decimal {
	if strings.TrimSpace(status) != "paid" {
		return decimal.Zero
	}
	return amount
}
