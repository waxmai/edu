package lesson_package

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"
)

func (s *service) Create(ctx context.Context, req *dto.LessonPackageCreateRequest) (int32, error) {
	actor := servicectx.ActorFromContext(ctx)
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), actor); err != nil {
		return 0, err
	}
	if req != nil {
		req.OrganizationID = actor.OrganizationID
		req.CampusID = actor.CampusID
	}
	item, err := buildLessonPackageModel(req)
	if err != nil {
		return 0, err
	}

	writeDB := dao.Use(s.db.GetDbW())
	student, err := ensureStudentExists(ctx, writeDB, item.StudentID)
	if err != nil {
		return 0, err
	}
	if err := serviceutil.EnsureStudentAccess(actor, student); err != nil {
		return 0, err
	}
	if item.OrganizationID <= 0 {
		item.OrganizationID = student.OrganizationID
	}
	if item.CampusID <= 0 {
		item.CampusID = student.CampusID
	}
	if err := writeDB.WithContext(ctx).LessonPackage.Create(item); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("lesson package id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureLessonPackageAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		if !item.UsedLessons.IsZero() || !item.PaidAmount.IsZero() {
			return apperr.Conflict("lesson package with used lessons or payment records cannot be deleted")
		}
		scheduleCount, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.LessonPackageID.Eq(id)).Count()
		if err != nil {
			return err
		}
		if scheduleCount > 0 {
			return apperr.Conflict("lesson package with schedules cannot be deleted")
		}
		info, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(id)).Delete()
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

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.LessonPackageUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("lesson package id must be positive")
	}
	updates, err := sanitizeLessonPackageUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureLessonPackageAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		merged, err := mergeLessonPackageUpdates(item, updates)
		if err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(id)).Updates(merged)
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
