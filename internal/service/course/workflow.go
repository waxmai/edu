package course

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"
)

func (s *service) Create(ctx context.Context, req *dto.CourseCreateRequest) (int32, error) {
	if err := validateCourseCreateRequest(req); err != nil {
		return 0, err
	}
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), servicectx.ActorFromContext(ctx)); err != nil {
		return 0, err
	}
	item := &model.Course{
		OrganizationID:  servicectx.ActorFromContext(ctx).OrganizationID,
		CampusID:        servicectx.ActorFromContext(ctx).CampusID,
		CourseName:      req.CourseName,
		Subject:         req.Subject,
		CourseType:      req.CourseType,
		DurationMinutes: defaultCourseDuration(req.DurationMinutes),
		FeeStandard:     req.FeeStandard,
		Remark:          req.Remark,
		Status:          normalizeCourseStatus(req.Status),
	}
	writeDB := dao.Use(s.db.GetDbW())
	if err := writeDB.WithContext(ctx).Course.Create(item); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("course id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Course.Where(tx.Course.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureCourseAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		count, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.CourseID.Eq(id)).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return apperr.Conflict("course with schedules cannot be deleted")
		}
		info, err := tx.WithContext(ctx).Course.Where(tx.Course.ID.Eq(id)).Delete()
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

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.CourseUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("course id must be positive")
	}
	updates, err := sanitizeCourseUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Course.Where(tx.Course.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureCourseAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).Course.Where(tx.Course.ID.Eq(id)).Updates(updates)
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
