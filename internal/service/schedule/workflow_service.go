package schedule

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func (s *service) Create(ctx context.Context, req *dto.ScheduleCreateRequest) (int32, error) {
	actor := servicectx.ActorFromContext(ctx)
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), actor); err != nil {
		return 0, err
	}
	if req != nil {
		req.OrganizationID = actor.OrganizationID
		req.CampusID = actor.CampusID
	}
	item, err := buildScheduleModel(req)
	if err != nil {
		return 0, err
	}
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		if err := validateScheduleRefsAndConflicts(ctx, tx, 0, item); err != nil {
			return err
		}
		return tx.WithContext(ctx).Schedule.Create(item)
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if strings.Contains(strings.ToLower(err.Error()), "duplicate entry") || (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) {
			teacherConflict, studentConflict, detectErr := detectConflicts(ctx, writeDB, 0, item)
			if detectErr == nil && (teacherConflict || studentConflict) {
				return 0, apperr.Conflict(conflictMessage(teacherConflict, studentConflict))
			}
			return 0, apperr.Conflict("当前时间段存在排课冲突")
		}
		return 0, err
	}
	return item.ID, nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("schedule id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		if item.ScheduleStatus == "completed" {
			return apperr.Conflict("completed schedule cannot be deleted")
		}
		childCount, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.OriginalScheduleID.Eq(id)).Count()
		if err != nil {
			return err
		}
		if childCount > 0 {
			return apperr.Conflict("schedule with linked reschedule or makeup records cannot be deleted")
		}
		info, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).Delete()
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

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.ScheduleUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("schedule id must be positive")
	}
	updates, err := sanitizeScheduleUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		candidate, err := mergeScheduleUpdates(item, updates)
		if err != nil {
			return err
		}
		if err := validateScheduleRefsAndConflicts(ctx, tx, id, candidate); err != nil {
			return err
		}
		if err := applyLessonPackageUsageOnScheduleTransition(ctx, tx, item, candidate); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).Updates(updates)
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

func (s *service) Cancel(ctx context.Context, id int32, req *dto.ScheduleCancelRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("schedule id must be positive")
	}
	updates := dto.ScheduleUpdateRequest{"scheduleStatus": "cancelled", "remark": appendReason("cancel", reqReason(req.Reason))}
	return s.UpdateByID(ctx, id, updates)
}

func (s *service) Leave(ctx context.Context, id int32, req *dto.ScheduleLeaveRequest) (int32, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("schedule id must be positive")
	}
	writeDB := dao.Use(s.db.GetDbW())
	var recordID int32
	err := writeDB.Transaction(func(tx *dao.Query) error {
		scheduleItem, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), scheduleItem); err != nil {
			return err
		}
		if scheduleItem.ScheduleStatus == "completed" || scheduleItem.ScheduleStatus == "cancelled" {
			return apperr.Conflict("schedule status does not allow leave")
		}
		if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).Updates(map[string]interface{}{
			"schedule_status": "leave",
			"remark":          appendReason("leave", reqReason(req.Reason)),
		}); err != nil {
			return err
		}
		record := &model.RescheduleRecord{OldScheduleID: id, OperationType: "leave", Reason: reqReason(req.Reason)}
		if err := tx.WithContext(ctx).RescheduleRecord.Create(record); err != nil {
			return err
		}
		recordID = record.ID
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	return recordID, nil
}

func (s *service) Reschedule(ctx context.Context, id int32, req *dto.ScheduleRescheduleRequest) (int32, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("schedule id must be positive")
	}
	if req == nil {
		return 0, apperr.InvalidArgument("schedule reschedule request is required")
	}
	writeDB := dao.Use(s.db.GetDbW())
	var recordID int32
	err := writeDB.Transaction(func(tx *dao.Query) error {
		origin, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), origin); err != nil {
			return err
		}
		newReq := &dto.ScheduleCreateRequest{
			OrganizationID:     origin.OrganizationID,
			CampusID:           origin.CampusID,
			StudentID:          origin.StudentID,
			CourseID:           origin.CourseID,
			TeacherID:          origin.TeacherID,
			LessonPackageID:    serviceutil.DecimalPtrToInt32Value(origin.LessonPackageID),
			ClassDate:          req.NewClassDate,
			StartTime:          req.NewStartTime,
			EndTime:            req.NewEndTime,
			Classroom:          req.Classroom,
			ScheduleStatus:     "scheduled",
			OriginalScheduleID: origin.ID,
			Remark:             strings.TrimSpace(req.Reason),
		}
		newSchedule, err := buildScheduleModel(newReq)
		if err != nil {
			return err
		}
		if err := validateScheduleRefsAndConflicts(ctx, tx, 0, newSchedule); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Schedule.Create(newSchedule); err != nil {
			return err
		}
		if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(id)).Updates(map[string]interface{}{
			"schedule_status": "rescheduled",
			"remark":          appendReason("reschedule", req.Reason),
		}); err != nil {
			return err
		}
		record := &model.RescheduleRecord{OldScheduleID: id, NewScheduleID: serviceutil.Int32Ptr(newSchedule.ID), OperationType: "reschedule", Reason: strings.TrimSpace(req.Reason)}
		if err := tx.WithContext(ctx).RescheduleRecord.Create(record); err != nil {
			return err
		}
		recordID = record.ID
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	return recordID, nil
}

func (s *service) CreateMakeup(ctx context.Context, req *dto.MakeupScheduleCreateRequest) (int32, error) {
	if req == nil {
		return 0, apperr.InvalidArgument("makeup schedule create request is required")
	}
	writeDB := dao.Use(s.db.GetDbW())
	var recordID int32
	err := writeDB.Transaction(func(tx *dao.Query) error {
		origin, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(req.OriginalScheduleID)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), origin); err != nil {
			return err
		}
		newReq := &dto.ScheduleCreateRequest{
			OrganizationID:     origin.OrganizationID,
			CampusID:           origin.CampusID,
			StudentID:          req.StudentID,
			CourseID:           req.CourseID,
			TeacherID:          req.TeacherID,
			LessonPackageID:    req.LessonPackageID,
			ClassDate:          req.ClassDate,
			StartTime:          req.StartTime,
			EndTime:            req.EndTime,
			Classroom:          req.Classroom,
			ScheduleStatus:     "scheduled",
			OriginalScheduleID: req.OriginalScheduleID,
			IsMakeup:           true,
			Remark:             req.Remark,
		}
		newSchedule, err := buildScheduleModel(newReq)
		if err != nil {
			return err
		}
		if err := validateScheduleRefsAndConflicts(ctx, tx, 0, newSchedule); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Schedule.Create(newSchedule); err != nil {
			return err
		}
		if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(origin.ID)).Updates(map[string]interface{}{"schedule_status": "makeup_pending"}); err != nil {
			return err
		}
		record := &model.RescheduleRecord{OldScheduleID: origin.ID, NewScheduleID: serviceutil.Int32Ptr(newSchedule.ID), OperationType: "makeup", Reason: strings.TrimSpace(req.Remark)}
		if err := tx.WithContext(ctx).RescheduleRecord.Create(record); err != nil {
			return err
		}
		recordID = record.ID
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	return recordID, nil
}
