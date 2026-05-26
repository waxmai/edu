package payment_record

import (
	"context"
	"database/sql"
	"time"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *service) Create(ctx context.Context, req *dto.PaymentRecordCreateRequest) (int32, error) {
	actor := servicectx.ActorFromContext(ctx)
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), actor); err != nil {
		return 0, err
	}
	if req != nil {
		req.OrganizationID = actor.OrganizationID
		req.CampusID = actor.CampusID
	}
	item, err := buildPaymentRecordModel(req)
	if err != nil {
		return 0, err
	}
	writeDB := dao.Use(s.db.GetDbW())
	var id int32
	err = writeDB.Transaction(func(tx *dao.Query) error {
		pkg, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(item.LessonPackageID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.NotFound("lesson package not found")
			}
			return err
		}
		if err := serviceutil.EnsureLessonPackageAccess(actor, pkg); err != nil {
			return err
		}
		if item.OrganizationID <= 0 {
			item.OrganizationID = pkg.OrganizationID
		}
		if item.CampusID <= 0 {
			item.CampusID = pkg.CampusID
		}
		if pkg.StudentID != item.StudentID {
			return apperr.Conflict("lesson package does not belong to student")
		}
		student, err := tx.WithContext(ctx).Student.Where(tx.Student.ID.Eq(item.StudentID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.NotFound("student not found")
			}
			return err
		}
		if err := serviceutil.EnsureStudentAccess(actor, student); err != nil {
			return err
		}
		if student.OrganizationID != item.OrganizationID || student.CampusID != item.CampusID {
			return apperr.Conflict("payment record scope does not match student scope")
		}
		if err := tx.WithContext(ctx).PaymentRecord.Create(item); err != nil {
			return err
		}
		newPaidAmount := pkg.PaidAmount
		if item.PaymentStatus == "paid" {
			newPaidAmount = newPaidAmount.Add(item.Amount)
		}
		newStatus := derivePackageStatusFromPayment(pkg.Status, pkg.TotalAmount, newPaidAmount, pkg.RemainLessons, pkg.EndDate)
		if _, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(pkg.ID)).Updates(map[string]interface{}{
			"paid_amount": newPaidAmount,
			"status":      newStatus,
		}); err != nil {
			return err
		}
		id = item.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("payment record id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).PaymentRecord.Where(tx.PaymentRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsurePaymentRecordAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		if item.PaymentStatus == "refunded" {
			return apperr.Conflict("refunded payment record cannot be deleted")
		}
		pkg, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(item.LessonPackageID)).First()
		if err != nil {
			return err
		}
		newPaidAmount := pkg.PaidAmount
		if item.PaymentStatus == "paid" {
			newPaidAmount = newPaidAmount.Sub(item.Amount)
		}
		if newPaidAmount.IsNegative() {
			newPaidAmount = decimal.Zero
		}
		status := derivePackageStatusFromPayment(pkg.Status, pkg.TotalAmount, newPaidAmount, pkg.RemainLessons, pkg.EndDate)
		if _, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(pkg.ID)).Updates(map[string]interface{}{"paid_amount": newPaidAmount, "status": status}); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).PaymentRecord.Where(tx.PaymentRecord.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}
		rowsAffected = info.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.PaymentRecordUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("payment record id must be positive")
	}
	updates, err := sanitizePaymentRecordUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).PaymentRecord.Where(tx.PaymentRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsurePaymentRecordAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		oldAmount := paidAmountDelta(item.Amount, item.PaymentStatus)
		newAmount := oldAmount
		if v, ok := updates["amount"].(decimal.Decimal); ok {
			newAmount = paidAmountDelta(v, item.PaymentStatus)
		}
		if v, ok := updates["payment_status"].(string); ok {
			baseAmount := item.Amount
			if rawAmount, ok := updates["amount"].(decimal.Decimal); ok {
				baseAmount = rawAmount
			}
			newAmount = paidAmountDelta(baseAmount, v)
		}
		info, err := tx.WithContext(ctx).PaymentRecord.Where(tx.PaymentRecord.ID.Eq(id)).Updates(updates)
		if err != nil {
			return err
		}
		rowsAffected = info.RowsAffected
		if !newAmount.Equal(oldAmount) {
			pkg, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(item.LessonPackageID)).First()
			if err != nil {
				return err
			}
			paidAmount := pkg.PaidAmount.Sub(oldAmount).Add(newAmount)
			if paidAmount.IsNegative() {
				paidAmount = decimal.Zero
			}
			status := derivePackageStatusFromPayment(pkg.Status, pkg.TotalAmount, paidAmount, pkg.RemainLessons, pkg.EndDate)
			if _, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(pkg.ID)).Updates(map[string]interface{}{"paid_amount": paidAmount, "status": status}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *service) IncomeStatistics(ctx context.Context, query dto.PaymentIncomeStatsQuery) (*dto.PaymentIncomeStatsResponse, error) {
	_, _, _, err := normalizePaymentStatsQuery(query)
	if err != nil {
		return nil, err
	}
	readDB := s.db.GetDbR().WithContext(ctx)
	start, end, _, _ := normalizePaymentStatsQuery(query)
	dbq, err := serviceutil.ApplyTenantGormScope(servicectx.ActorFromContext(ctx), readDB.Model(&model.PaymentRecord{}), "organization_id", "campus_id")
	if err != nil {
		return nil, err
	}
	dbq = dbq.Where("payment_status = ?", "paid")
	if !start.IsZero() {
		dbq = dbq.Where("payment_time >= ?", start)
	}
	if !end.IsZero() {
		dbq = dbq.Where("payment_time < ?", end.Add(24*time.Hour))
	}
	stats := &dto.PaymentIncomeStatsResponse{}
	var totalAmount sql.NullFloat64
	row := dbq.Select("COALESCE(SUM(amount), 0) AS total_amount, COUNT(*) AS record_count").Row()
	if err := row.Scan(&totalAmount, &stats.RecordCount); err != nil {
		return nil, err
	}
	stats.TotalAmount = totalAmount.Float64
	return stats, nil
}
