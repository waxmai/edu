package student

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"
)

func (s *service) Create(ctx context.Context, req *dto.StudentCreateRequest) (int32, error) {
	if err := validateStudentCreateRequest(req); err != nil {
		return 0, err
	}
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), actorFromContext(ctx)); err != nil {
		return 0, err
	}

	item := &model.Student{
		OrganizationID: actorFromContext(ctx).OrganizationID,
		CampusID:       actorFromContext(ctx).CampusID,
		StudentName:    defaultString(req.StudentName, ""),
		Gender:         defaultString(req.Gender, "unknown"),
		Grade:          req.Grade,
		Phone:          req.Phone,
		ParentName:     req.ParentName,
		ParentPhone:    req.ParentPhone,
		Subject:        req.Subject,
		TeachingType:   req.TeachingType,
		Status:         defaultString(req.Status, "active"),
		Remark:         req.Remark,
	}

	writeDB := dao.Use(s.db.GetDbW())
	if err := writeDB.WithContext(ctx).Student.Create(item); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.StudentUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("student id must be positive")
	}
	updates, err := sanitizeStudentUpdates(req)
	if err != nil {
		return 0, err
	}

	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Student.Where(tx.Student.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureStudentAccess(actorFromContext(ctx), item); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).Student.Where(tx.Student.ID.Eq(id)).Updates(updates)
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

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("student id must be positive")
	}

	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Student.Where(tx.Student.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureStudentAccess(actorFromContext(ctx), item); err != nil {
			return err
		}
		lessonPackageCount, err := tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.StudentID.Eq(id)).Count()
		if err != nil {
			return err
		}
		if lessonPackageCount > 0 {
			return apperr.Conflict("student with lesson packages cannot be deleted")
		}
		scheduleCount, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.StudentID.Eq(id)).Count()
		if err != nil {
			return err
		}
		if scheduleCount > 0 {
			return apperr.Conflict("student with schedules cannot be deleted")
		}
		info, err := tx.WithContext(ctx).Student.Where(tx.Student.ID.Eq(id)).Delete()
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
