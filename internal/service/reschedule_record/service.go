package reschedule_record

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	tenantservice "edu-schedule-system/internal/service/tenant"
	"gorm.io/gen"
	"gorm.io/gorm"
)

var operationTypes = map[string]struct{}{"leave": {}, "reschedule": {}, "makeup": {}, "cancel": {}}

const rescheduleDateLayout = "2006-01-02"

type Service interface {
	Create(ctx context.Context, req *dto.RescheduleRecordCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.RescheduleRecordListQuery) (dto.RescheduleRecordListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.RescheduleRecordResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.RescheduleRecordUpdateRequest) (int64, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }

func (s *service) Create(ctx context.Context, req *dto.RescheduleRecordCreateRequest) (int32, error) {
	actor := servicectx.ActorFromContext(ctx)
	if err := tenantservice.EnsureTenantWriteAllowed(ctx, tenantservice.New(s.db), actor); err != nil {
		return 0, err
	}
	item, status, err := buildRescheduleRecordModel(req)
	if err != nil {
		return 0, err
	}
	writeDB := dao.Use(s.db.GetDbW())
	var id int32
	err = writeDB.Transaction(func(tx *dao.Query) error {
		oldSchedule, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(item.OldScheduleID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.NotFound("old schedule not found")
			}
			return err
		}
		if err := serviceutil.EnsureScheduleAccess(actor, oldSchedule); err != nil {
			return err
		}
		item.OrganizationID = oldSchedule.OrganizationID
		item.CampusID = oldSchedule.CampusID
		if item.OperatorID == nil || *item.OperatorID <= 0 {
			item.OperatorID = serviceutil.Int32Ptr(actor.Id)
		}
		if item.NewScheduleID != nil && *item.NewScheduleID > 0 {
			newSchedule, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(*item.NewScheduleID)).First()
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return apperr.NotFound("new schedule not found")
				}
				return err
			}
			if err := serviceutil.EnsureScheduleAccess(actor, newSchedule); err != nil {
				return err
			}
			if newSchedule.OrganizationID != item.OrganizationID || newSchedule.CampusID != item.CampusID {
				return apperr.Conflict("reschedule records must stay inside the same organization and campus")
			}
		}
		if err := tx.WithContext(ctx).RescheduleRecord.Create(item); err != nil {
			return err
		}
		if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(item.OldScheduleID)).Updates(map[string]interface{}{"schedule_status": status}); err != nil {
			return err
		}
		if item.NewScheduleID != nil && *item.NewScheduleID > 0 {
			updates := map[string]interface{}{"original_schedule_id": item.OldScheduleID}
			if item.OperationType == "makeup" {
				updates["is_makeup"] = true
			}
			if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(*item.NewScheduleID)).Updates(updates); err != nil {
				return err
			}
		}
		id = item.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *service) List(ctx context.Context, query ...dto.RescheduleRecordListQuery) (dto.RescheduleRecordListResponse, error) {
	listQuery := dto.RescheduleRecordListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validateRescheduleRecordListQuery(listQuery); err != nil {
		return nil, err
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	conditions, err := buildRescheduleRecordConditions(ctx, readDB, listQuery)
	if err != nil {
		return nil, err
	}
	do := readDB.WithContext(ctx).RescheduleRecord.Where(servicectx.ScopeRescheduleRecords(actor, readDB, conditions)...).Order(readDB.RescheduleRecord.ID.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	scheduleMap, operatorMap, err := s.loadRescheduleRelations(ctx, readDB, items)
	if err != nil {
		return nil, err
	}
	return toRescheduleRecordResponses(items, scheduleMap, operatorMap), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.RescheduleRecordResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("reschedule record id must be positive")
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).RescheduleRecord.Where(servicectx.ScopeRescheduleRecords(actor, readDB, []gen.Condition{readDB.RescheduleRecord.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	scheduleMap, operatorMap, err := s.loadRescheduleRelations(ctx, readDB, []*model.RescheduleRecord{item})
	if err != nil {
		return nil, err
	}
	return toRescheduleRecordResponse(item, scheduleMap, operatorMap), nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("reschedule record id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).RescheduleRecord.Where(tx.RescheduleRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureRescheduleRecordAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(item.OldScheduleID)).Updates(map[string]interface{}{"schedule_status": "scheduled"}); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).RescheduleRecord.Where(tx.RescheduleRecord.ID.Eq(id)).Delete()
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

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.RescheduleRecordUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("reschedule record id must be positive")
	}
	updates, err := sanitizeRescheduleRecordUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).RescheduleRecord.Where(tx.RescheduleRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureRescheduleRecordAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		info, err := tx.WithContext(ctx).RescheduleRecord.Where(tx.RescheduleRecord.ID.Eq(id)).Updates(updates)
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

func buildRescheduleRecordModel(req *dto.RescheduleRecordCreateRequest) (*model.RescheduleRecord, string, error) {
	if req == nil {
		return nil, "", apperr.InvalidArgument("reschedule record create request is required")
	}
	if req.OldScheduleID <= 0 {
		return nil, "", apperr.InvalidArgument("oldScheduleId must be positive")
	}
	operationType := strings.TrimSpace(req.OperationType)
	if _, ok := operationTypes[operationType]; !ok {
		return nil, "", apperr.InvalidArgument("operationType is invalid")
	}
	if operationType == "reschedule" || operationType == "makeup" {
		if req.NewScheduleID <= 0 {
			return nil, "", apperr.InvalidArgument("newScheduleId must be positive for reschedule or makeup")
		}
	}
	status := map[string]string{"leave": "leave", "reschedule": "rescheduled", "makeup": "makeup_pending", "cancel": "cancelled"}[operationType]
	return &model.RescheduleRecord{OldScheduleID: req.OldScheduleID, NewScheduleID: serviceutil.NilIfNonPositiveInt32(req.NewScheduleID), OperationType: operationType, Reason: strings.TrimSpace(req.Reason), OperatorID: serviceutil.NilIfNonPositiveInt32(req.OperatorID)}, status, nil
}

func sanitizeRescheduleRecordUpdates(req dto.RescheduleRecordUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("reschedule record update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "newScheduleId", "operatorId":
			n, ok := value.(float64)
			if !ok || n < 0 {
				return nil, apperr.InvalidArgument(field + " must be a non-negative number")
			}
			key := map[string]string{"newScheduleId": "new_schedule_id", "operatorId": "operator_id"}[field]
			updates[key] = serviceutil.NilIfNonPositiveInt32(int32(n))
		case "operationType":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("operationType must be a string")
			}
			text = strings.TrimSpace(text)
			if _, ok := operationTypes[text]; !ok {
				return nil, apperr.InvalidArgument("operationType is invalid")
			}
			updates["operation_type"] = text
		case "reason":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument("reason must be a string")
			}
			updates["reason"] = strings.TrimSpace(text)
		default:
			return nil, apperr.InvalidArgument("reschedule record update field " + field + " is not allowed")
		}
	}
	return updates, nil
}

func validateRescheduleRecordListQuery(query dto.RescheduleRecordListQuery) error {
	if query.StudentID < 0 {
		return apperr.InvalidArgument("studentId must not be negative")
	}
	if operationType := strings.TrimSpace(query.OperationType); operationType != "" {
		if _, ok := operationTypes[operationType]; !ok {
			return apperr.InvalidArgument("operationType is invalid")
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
	_, _, err := parseRescheduleDateRange(query.StartDate, query.EndDate)
	return err
}

func parseRescheduleDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	var start time.Time
	var end time.Time
	var err error
	if strings.TrimSpace(startDate) != "" {
		start, err = time.Parse(rescheduleDateLayout, strings.TrimSpace(startDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must use format 2006-01-02")
		}
	}
	if strings.TrimSpace(endDate) != "" {
		end, err = time.Parse(rescheduleDateLayout, strings.TrimSpace(endDate))
		if err != nil {
			return time.Time{}, time.Time{}, apperr.InvalidArgument("endDate must use format 2006-01-02")
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must be earlier than or equal to endDate")
	}
	return start, end, nil
}

func buildRescheduleRecordConditions(ctx context.Context, q *dao.Query, query dto.RescheduleRecordListQuery) ([]gen.Condition, error) {
	conditions := make([]gen.Condition, 0, 5)
	if studentID := query.StudentID; studentID > 0 {
		schedules, err := q.WithContext(ctx).Schedule.Select(q.Schedule.ID).Where(q.Schedule.StudentID.Eq(studentID)).Find()
		if err != nil {
			return nil, err
		}
		ids := make([]int32, 0, len(schedules))
		for _, schedule := range schedules {
			if schedule != nil {
				ids = append(ids, schedule.ID)
			}
		}
		if len(ids) == 0 {
			conditions = append(conditions, q.RescheduleRecord.ID.Eq(-1))
		} else {
			conditions = append(conditions, q.RescheduleRecord.OldScheduleID.In(ids...))
		}
	}
	if operationType := strings.TrimSpace(query.OperationType); operationType != "" {
		conditions = append(conditions, q.RescheduleRecord.OperationType.Eq(operationType))
	}
	if start, end, err := parseRescheduleDateRange(query.StartDate, query.EndDate); err == nil {
		if !start.IsZero() {
			conditions = append(conditions, q.RescheduleRecord.CreatedAt.Gte(start))
		}
		if !end.IsZero() {
			conditions = append(conditions, q.RescheduleRecord.CreatedAt.Lt(end.Add(24*time.Hour)))
		}
	}
	return conditions, nil
}

func toRescheduleRecordResponses(items []*model.RescheduleRecord, scheduleMap map[int32]*model.Schedule, operatorMap map[int32]*model.SysUser) dto.RescheduleRecordListResponse {
	list := make(dto.RescheduleRecordListResponse, 0, len(items))
	for _, item := range items {
		if resp := toRescheduleRecordResponse(item, scheduleMap, operatorMap); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toRescheduleRecordResponse(item *model.RescheduleRecord, scheduleMap map[int32]*model.Schedule, operatorMap map[int32]*model.SysUser) *dto.RescheduleRecordResponse {
	if item == nil {
		return nil
	}
	oldSummary := buildScheduleSummary(scheduleMap[item.OldScheduleID])
	newSummary := buildScheduleSummary(scheduleMap[serviceutil.DecimalPtrToInt32Value(item.NewScheduleID)])
	operatorName := ""
	operatorID := serviceutil.DecimalPtrToInt32Value(item.OperatorID)
	if operator := operatorMap[operatorID]; operator != nil {
		operatorName = operator.RealName
		if operatorName == "" {
			operatorName = operator.Username
		}
	}
	return &dto.RescheduleRecordResponse{ID: item.ID, OldScheduleID: item.OldScheduleID, OldScheduleSummary: oldSummary, NewScheduleID: serviceutil.DecimalPtrToInt32Value(item.NewScheduleID), NewScheduleSummary: newSummary, OperationType: item.OperationType, Reason: item.Reason, OperatorID: operatorID, OperatorName: operatorName, CreatedAt: item.CreatedAt}
}

func (s *service) loadRescheduleRelations(ctx context.Context, readDB *dao.Query, items []*model.RescheduleRecord) (map[int32]*model.Schedule, map[int32]*model.SysUser, error) {
	scheduleMap, err := s.loadScheduleMap(ctx, readDB, items)
	if err != nil {
		return nil, nil, err
	}
	operatorIDs := make([]int32, 0, len(items))
	seenOperators := map[int32]struct{}{}
	for _, item := range items {
		if item == nil {
			continue
		}
		operatorID := serviceutil.DecimalPtrToInt32Value(item.OperatorID)
		if operatorID <= 0 {
			continue
		}
		if _, ok := seenOperators[operatorID]; ok {
			continue
		}
		seenOperators[operatorID] = struct{}{}
		operatorIDs = append(operatorIDs, operatorID)
	}
	operatorMap := map[int32]*model.SysUser{}
	if len(operatorIDs) == 0 {
		return scheduleMap, operatorMap, nil
	}
	operators, err := readDB.WithContext(ctx).SysUser.Where(readDB.SysUser.ID.In(operatorIDs...)).Find()
	if err != nil {
		return nil, nil, err
	}
	for _, item := range operators {
		operatorMap[item.ID] = item
	}
	return scheduleMap, operatorMap, nil
}

func (s *service) loadScheduleMap(ctx context.Context, readDB *dao.Query, items []*model.RescheduleRecord) (map[int32]*model.Schedule, error) {
	ids := make([]int32, 0, len(items)*2)
	seen := map[int32]struct{}{}
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seen[item.OldScheduleID]; !ok {
			seen[item.OldScheduleID] = struct{}{}
			ids = append(ids, item.OldScheduleID)
		}
		newID := serviceutil.DecimalPtrToInt32Value(item.NewScheduleID)
		if newID > 0 {
			if _, ok := seen[newID]; !ok {
				seen[newID] = struct{}{}
				ids = append(ids, newID)
			}
		}
	}
	result := map[int32]*model.Schedule{}
	if len(ids) == 0 {
		return result, nil
	}
	schedules, err := readDB.WithContext(ctx).Schedule.Where(readDB.Schedule.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range schedules {
		result[item.ID] = item
	}
	return result, nil
}

func buildScheduleSummary(item *model.Schedule) string {
	if item == nil {
		return "-"
	}
	return item.ClassDate.Format("2006-01-02") + " " + item.StartTime.Format("15:04")
}
