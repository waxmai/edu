package lesson_record

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"edu-schedule-system/internal/service/serviceutil"
	"github.com/shopspring/decimal"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var lessonRecordAttendanceStatuses = map[string]struct{}{
	"present": {},
	"late":    {},
	"absent":  {},
	"leave":   {},
}

type Service interface {
	Create(ctx context.Context, req *dto.LessonRecordCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.LessonRecordListQuery) (dto.LessonRecordListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.LessonRecordResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.LessonRecordUpdateRequest) (int64, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }

func (s *service) Create(ctx context.Context, req *dto.LessonRecordCreateRequest) (int32, error) {
	item, deductDelta, err := buildLessonRecordModel(req)
	if err != nil {
		return 0, err
	}
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		scheduleItem, err := validateLessonRecordReferences(ctx, tx, item.ScheduleID, item.StudentID, item.TeacherID)
		if err != nil {
			return err
		}
		item.OrganizationID = scheduleItem.OrganizationID
		item.CampusID = scheduleItem.CampusID
		if _, err := tx.WithContext(ctx).LessonRecord.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.LessonRecord.ScheduleID.Eq(item.ScheduleID)).First(); err == nil {
			return apperr.Conflict("lesson record for schedule already exists")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if scheduleItem.ScheduleStatus != "completed" {
			if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(scheduleItem.ID)).Updates(map[string]interface{}{"schedule_status": "completed"}); err != nil {
				return err
			}
		}
		if deductDelta.GreaterThan(decimal.Zero) {
			if err := applyLessonDeduction(ctx, tx, scheduleItem.LessonPackageID, deductDelta); err != nil {
				return err
			}
		}
		return tx.WithContext(ctx).LessonRecord.Create(item)
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *service) List(ctx context.Context, query ...dto.LessonRecordListQuery) (dto.LessonRecordListResponse, error) {
	listQuery := dto.LessonRecordListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validateLessonRecordListQuery(listQuery); err != nil {
		return nil, err
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	do := readDB.WithContext(ctx).LessonRecord.Where(servicectx.ScopeLessonRecords(actor, readDB, buildLessonRecordConditions(readDB, listQuery))...).Order(readDB.LessonRecord.RecordedAt.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	studentMap, teacherMap, err := s.loadLessonRecordRelations(ctx, readDB, items)
	if err != nil {
		return nil, err
	}
	return toLessonRecordResponses(items, studentMap, teacherMap), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.LessonRecordResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("lesson record id must be positive")
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).LessonRecord.Where(servicectx.ScopeLessonRecords(actor, readDB, []gen.Condition{readDB.LessonRecord.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	studentMap, teacherMap, err := s.loadLessonRecordRelations(ctx, readDB, []*model.LessonRecord{item})
	if err != nil {
		return nil, err
	}
	return toLessonRecordResponse(item, studentMap, teacherMap), nil
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("lesson record id must be positive")
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		item, err := tx.WithContext(ctx).LessonRecord.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.LessonRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureLessonRecordAccess(servicectx.ActorFromContext(ctx), item); err != nil {
			return err
		}
		scheduleItem, err := tx.WithContext(ctx).Schedule.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.Schedule.ID.Eq(item.ScheduleID)).First()
		if err != nil {
			return err
		}
		if item.LessonDeducted && item.DeductLessonCount > 0 {
			if err := applyLessonDeduction(ctx, tx, scheduleItem.LessonPackageID, decimal.NewFromFloat(-item.DeductLessonCount)); err != nil {
				return err
			}
		}
		info, err := tx.WithContext(ctx).LessonRecord.Where(tx.LessonRecord.ID.Eq(id)).Delete()
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

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.LessonRecordUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("lesson record id must be positive")
	}
	updates, intent, err := sanitizeLessonRecordUpdates(req)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		current, err := tx.WithContext(ctx).LessonRecord.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.LessonRecord.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if err := serviceutil.EnsureLessonRecordAccess(servicectx.ActorFromContext(ctx), current); err != nil {
			return err
		}
		candidate, err := mergeLessonRecordUpdates(current, updates)
		if err != nil {
			return err
		}
		scheduleItem, err := validateLessonRecordReferences(ctx, tx, candidate.ScheduleID, candidate.StudentID, candidate.TeacherID)
		if err != nil {
			return err
		}
		if current.ScheduleID != candidate.ScheduleID {
			if _, err := tx.WithContext(ctx).LessonRecord.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.LessonRecord.ScheduleID.Eq(candidate.ScheduleID)).First(); err == nil {
				return apperr.Conflict("lesson record for schedule already exists")
			} else if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
		}
		if scheduleItem.ScheduleStatus != "completed" {
			if _, err := tx.WithContext(ctx).Schedule.Where(tx.Schedule.ID.Eq(scheduleItem.ID)).Updates(map[string]interface{}{"schedule_status": "completed"}); err != nil {
				return err
			}
		}
		currentDelta := decimal.Zero
		if current.LessonDeducted && current.DeductLessonCount > 0 {
			currentDelta = decimal.NewFromFloat(current.DeductLessonCount)
		}
		nextDelta := lessonRecordDeductDelta(candidate)
		delta := nextDelta.Sub(currentDelta)
		if !delta.IsZero() {
			if err := applyLessonDeduction(ctx, tx, scheduleItem.LessonPackageID, delta); err != nil {
				return err
			}
		}
		if intent.needDeductSpecified && !intent.needDeduct && !intent.deductCountSpecified {
			updates["deduct_lesson_count"] = 0.0
		}
		info, err := tx.WithContext(ctx).LessonRecord.Where(tx.LessonRecord.ID.Eq(id)).Updates(updates)
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

type updateIntent struct {
	needDeductSpecified  bool
	needDeduct           bool
	deductCountSpecified bool
}

func buildLessonRecordModel(req *dto.LessonRecordCreateRequest) (*model.LessonRecord, decimal.Decimal, error) {
	if req == nil {
		return nil, decimal.Zero, apperr.InvalidArgument("lesson record create request is required")
	}
	if req.ScheduleID <= 0 {
		return nil, decimal.Zero, apperr.InvalidArgument("scheduleId must be positive")
	}
	if req.StudentID <= 0 {
		return nil, decimal.Zero, apperr.InvalidArgument("studentId must be positive")
	}
	if req.TeacherID <= 0 {
		return nil, decimal.Zero, apperr.InvalidArgument("teacherId must be positive")
	}
	attendance := strings.TrimSpace(req.AttendanceStatus)
	if _, ok := lessonRecordAttendanceStatuses[attendance]; !ok {
		return nil, decimal.Zero, apperr.InvalidArgument("attendanceStatus is invalid")
	}
	lessonDeducted, deductCount, err := resolveDeductionFields(req.NeedDeductLesson, req.LessonDeducted, req.DeductLessonCount)
	if err != nil {
		return nil, decimal.Zero, err
	}
	item := &model.LessonRecord{
		ScheduleID:        req.ScheduleID,
		StudentID:         req.StudentID,
		TeacherID:         req.TeacherID,
		AttendanceStatus:  attendance,
		LessonContent:     optionalText(req.LessonContent),
		Homework:          optionalText(req.Homework),
		Feedback:          optionalText(req.Feedback),
		LessonDeducted:    lessonDeducted,
		DeductLessonCount: deductCount,
	}
	return item, lessonRecordDeductDelta(item), nil
}

func sanitizeLessonRecordUpdates(req dto.LessonRecordUpdateRequest) (map[string]interface{}, updateIntent, error) {
	if len(req) == 0 {
		return nil, updateIntent{}, apperr.InvalidArgument("lesson record update fields are required")
	}
	updates := make(map[string]interface{}, len(req))
	intent := updateIntent{}
	var needDeductPtr *bool
	var lessonDeductedPtr *bool
	var deductCount float64
	var deductCountSpecified bool
	for field, value := range req {
		switch field {
		case "scheduleId":
			v, err := int32Value(field, value, true)
			if err != nil {
				return nil, intent, err
			}
			updates["schedule_id"] = v
		case "studentId":
			v, err := int32Value(field, value, true)
			if err != nil {
				return nil, intent, err
			}
			updates["student_id"] = v
		case "teacherId":
			v, err := int32Value(field, value, true)
			if err != nil {
				return nil, intent, err
			}
			updates["teacher_id"] = v
		case "attendanceStatus":
			v, err := stringValue(field, value, true)
			if err != nil {
				return nil, intent, err
			}
			if _, ok := lessonRecordAttendanceStatuses[v]; !ok {
				return nil, intent, apperr.InvalidArgument("attendanceStatus is invalid")
			}
			updates["attendance_status"] = v
		case "lessonContent":
			v, err := stringValue(field, value, false)
			if err != nil {
				return nil, intent, err
			}
			updates["lesson_content"] = optionalText(v)
		case "homework":
			v, err := stringValue(field, value, false)
			if err != nil {
				return nil, intent, err
			}
			updates["homework"] = optionalText(v)
		case "feedback":
			v, err := stringValue(field, value, false)
			if err != nil {
				return nil, intent, err
			}
			updates["feedback"] = optionalText(v)
		case "needDeductLesson":
			v, err := boolValue(field, value)
			if err != nil {
				return nil, intent, err
			}
			needDeductPtr = &v
			intent.needDeductSpecified = true
			intent.needDeduct = v
		case "lessonDeducted":
			v, err := boolValue(field, value)
			if err != nil {
				return nil, intent, err
			}
			lessonDeductedPtr = &v
		case "deductLessonCount":
			v, err := nonNegativeFloat(field, value)
			if err != nil {
				return nil, intent, err
			}
			deductCount = v
			deductCountSpecified = true
			intent.deductCountSpecified = true
		default:
			return nil, intent, apperr.InvalidArgument("lesson record update field " + field + " is not allowed")
		}
	}
	if needDeductPtr != nil || lessonDeductedPtr != nil || deductCountSpecified {
		lessonDeducted, resolvedCount, err := resolveDeductionFields(needDeductPtr, lessonDeductedPtr, func() *float64 {
			if !deductCountSpecified {
				return nil
			}
			return &deductCount
		}())
		if err != nil {
			return nil, intent, err
		}
		updates["lesson_deducted"] = lessonDeducted
		updates["deduct_lesson_count"] = resolvedCount
	}
	return updates, intent, nil
}

func mergeLessonRecordUpdates(current *model.LessonRecord, updates map[string]interface{}) (*model.LessonRecord, error) {
	if current == nil {
		return nil, apperr.InvalidArgument("current lesson record is required")
	}
	copyItem := *current
	if v, ok := updates["schedule_id"].(int32); ok {
		copyItem.ScheduleID = v
	}
	if v, ok := updates["student_id"].(int32); ok {
		copyItem.StudentID = v
	}
	if v, ok := updates["teacher_id"].(int32); ok {
		copyItem.TeacherID = v
	}
	if v, ok := updates["attendance_status"].(string); ok {
		copyItem.AttendanceStatus = v
	}
	if v, ok := updates["lesson_content"].(*string); ok {
		copyItem.LessonContent = v
	}
	if v, ok := updates["homework"].(*string); ok {
		copyItem.Homework = v
	}
	if v, ok := updates["feedback"].(*string); ok {
		copyItem.Feedback = v
	}
	if v, ok := updates["lesson_deducted"].(bool); ok {
		copyItem.LessonDeducted = v
	}
	if v, ok := updates["deduct_lesson_count"].(float64); ok {
		copyItem.DeductLessonCount = v
	}
	if copyItem.LessonDeducted && copyItem.DeductLessonCount <= 0 {
		return nil, apperr.InvalidArgument("deductLessonCount must be greater than 0 when lesson is deducted")
	}
	if !copyItem.LessonDeducted && copyItem.DeductLessonCount < 0 {
		return nil, apperr.InvalidArgument("deductLessonCount must not be negative")
	}
	if !copyItem.LessonDeducted && copyItem.DeductLessonCount > 0 {
		return nil, apperr.InvalidArgument("lessonDeducted must be true when deductLessonCount is greater than 0")
	}
	return &copyItem, nil
}

func validateLessonRecordReferences(ctx context.Context, q *dao.Query, scheduleID, studentID, teacherID int32) (*model.Schedule, error) {
	scheduleItem, err := q.WithContext(ctx).Schedule.Clauses(clause.Locking{Strength: "UPDATE"}).Where(q.Schedule.ID.Eq(scheduleID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("schedule not found")
		}
		return nil, err
	}
	if err := serviceutil.EnsureScheduleAccess(servicectx.ActorFromContext(ctx), scheduleItem); err != nil {
		return nil, err
	}
	student, err := q.WithContext(ctx).Student.Where(q.Student.ID.Eq(studentID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("student not found")
		}
		return nil, err
	}
	if err := serviceutil.EnsureStudentAccess(servicectx.ActorFromContext(ctx), student); err != nil {
		return nil, err
	}
	teacher, err := q.WithContext(ctx).SysUser.Where(q.SysUser.ID.Eq(teacherID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("teacher not found")
		}
		return nil, err
	}
	if strings.TrimSpace(teacher.RoleCode) != "teacher" {
		return nil, apperr.InvalidArgument("teacherId must reference a teacher user")
	}
	teacherOrgID := serviceutil.Int32PtrValue(teacher.OrganizationID)
	teacherCampusID := serviceutil.Int32PtrValue(teacher.CampusID)
	if err := serviceutil.EnsureTenantAccess(servicectx.ActorFromContext(ctx), teacherOrgID, teacherCampusID); err != nil {
		return nil, err
	}
	if scheduleItem.OrganizationID != student.OrganizationID || scheduleItem.OrganizationID != teacherOrgID || scheduleItem.CampusID != student.CampusID || scheduleItem.CampusID != teacherCampusID {
		return nil, apperr.Conflict("lesson record references must belong to the same organization and campus")
	}
	if scheduleItem.StudentID != studentID {
		return nil, apperr.Conflict("studentId does not match schedule")
	}
	if scheduleItem.TeacherID != teacherID {
		return nil, apperr.Conflict("teacherId does not match schedule")
	}
	return scheduleItem, nil
}

func validateLessonRecordListQuery(query dto.LessonRecordListQuery) error {
	if query.StudentID < 0 || query.TeacherID < 0 {
		return apperr.InvalidArgument("studentId and teacherId must not be negative")
	}
	if status := strings.TrimSpace(query.AttendanceStatus); status != "" {
		if _, ok := lessonRecordAttendanceStatuses[status]; !ok {
			return apperr.InvalidArgument("attendanceStatus is invalid")
		}
	}
	start, end, err := parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return err
	}
	_ = start
	_ = end
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

func buildLessonRecordConditions(q *dao.Query, query dto.LessonRecordListQuery) []gen.Condition {
	conditions := make([]gen.Condition, 0, 5)
	if query.StudentID > 0 {
		conditions = append(conditions, q.LessonRecord.StudentID.Eq(query.StudentID))
	}
	if query.TeacherID > 0 {
		conditions = append(conditions, q.LessonRecord.TeacherID.Eq(query.TeacherID))
	}
	if status := strings.TrimSpace(query.AttendanceStatus); status != "" {
		conditions = append(conditions, q.LessonRecord.AttendanceStatus.Eq(status))
	}
	if start, end, err := parseDateRange(query.StartDate, query.EndDate); err == nil {
		if !start.IsZero() {
			conditions = append(conditions, q.LessonRecord.RecordedAt.Gte(start))
		}
		if !end.IsZero() {
			conditions = append(conditions, q.LessonRecord.RecordedAt.Lte(end.Add(24*time.Hour-time.Nanosecond)))
		}
	}
	return conditions
}

func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := parseDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must use format 2006-01-02")
	}
	end, err := parseDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("endDate must use format 2006-01-02")
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, apperr.InvalidArgument("startDate must be earlier than or equal to endDate")
	}
	return start, end, nil
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

func boolValue(field string, value interface{}) (bool, error) {
	v, ok := value.(bool)
	if !ok {
		return false, apperr.InvalidArgument(field + " must be a boolean")
	}
	return v, nil
}

func resolveDeductionFields(needDeductLesson, lessonDeducted *bool, deductLessonCount interface{}) (bool, float64, error) {
	wantsDeduct := false
	if needDeductLesson != nil {
		wantsDeduct = *needDeductLesson
	}
	if lessonDeducted != nil {
		wantsDeduct = wantsDeduct || *lessonDeducted
	}
	count := 0.0
	if deductLessonCount != nil {
		switch v := deductLessonCount.(type) {
		case float64:
			count = v
		case *float64:
			if v != nil {
				count = *v
			}
		default:
			return false, 0, apperr.InvalidArgument("deductLessonCount must be a number")
		}
	}
	if count < 0 {
		return false, 0, apperr.InvalidArgument("deductLessonCount must not be negative")
	}
	if !wantsDeduct {
		if count > 0 {
			return false, 0, apperr.InvalidArgument("needDeductLesson must be true when deductLessonCount is greater than 0")
		}
		return false, 0, nil
	}
	if count <= 0 {
		count = 1
	}
	return true, count, nil
}

func lessonRecordDeductDelta(item *model.LessonRecord) decimal.Decimal {
	if item == nil || !item.LessonDeducted || item.DeductLessonCount <= 0 {
		return decimal.Zero
	}
	return decimal.NewFromFloat(item.DeductLessonCount)
}

func applyLessonDeduction(ctx context.Context, tx *dao.Query, lessonPackageID *int32, delta decimal.Decimal) error {
	if delta.IsZero() {
		return nil
	}
	if lessonPackageID == nil || *lessonPackageID <= 0 {
		if delta.GreaterThan(decimal.Zero) {
			return apperr.Conflict("schedule has no lesson package for deduction")
		}
		return nil
	}
	pkg, err := tx.WithContext(ctx).LessonPackage.Clauses(clause.Locking{Strength: "UPDATE"}).Where(tx.LessonPackage.ID.Eq(*lessonPackageID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperr.NotFound("lesson package not found")
		}
		return err
	}
	usedLessons := pkg.UsedLessons.Add(delta)
	remainLessons := pkg.RemainLessons.Sub(delta)
	if usedLessons.IsNegative() || remainLessons.IsNegative() || usedLessons.GreaterThan(pkg.TotalLessons) {
		return apperr.Conflict("lesson package lessons are insufficient for lesson deduction")
	}
	status := pkg.Status
	if remainLessons.LessThanOrEqual(decimal.Zero) {
		status = "exhausted"
	} else if status == "exhausted" {
		status = "active"
	}
	_, err = tx.WithContext(ctx).LessonPackage.Where(tx.LessonPackage.ID.Eq(*lessonPackageID)).Updates(map[string]interface{}{
		"used_lessons":   usedLessons,
		"remain_lessons": remainLessons,
		"status":         status,
	})
	return err
}

func optionalText(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func toLessonRecordResponses(items []*model.LessonRecord, studentMap map[int32]*model.Student, teacherMap map[int32]*model.SysUser) dto.LessonRecordListResponse {
	list := make(dto.LessonRecordListResponse, 0, len(items))
	for _, item := range items {
		if resp := toLessonRecordResponse(item, studentMap, teacherMap); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toLessonRecordResponse(item *model.LessonRecord, studentMap map[int32]*model.Student, teacherMap map[int32]*model.SysUser) *dto.LessonRecordResponse {
	if item == nil {
		return nil
	}
	studentName := ""
	teacherName := ""
	if student := studentMap[item.StudentID]; student != nil {
		studentName = student.StudentName
	}
	if teacher := teacherMap[item.TeacherID]; teacher != nil {
		teacherName = teacher.RealName
		if teacherName == "" {
			teacherName = teacher.Username
		}
	}
	return &dto.LessonRecordResponse{
		ID:                item.ID,
		ScheduleID:        item.ScheduleID,
		StudentID:         item.StudentID,
		StudentName:       studentName,
		TeacherID:         item.TeacherID,
		TeacherName:       teacherName,
		AttendanceStatus:  item.AttendanceStatus,
		LessonContent:     derefString(item.LessonContent),
		Homework:          derefString(item.Homework),
		Feedback:          derefString(item.Feedback),
		LessonDeducted:    item.LessonDeducted,
		DeductLessonCount: item.DeductLessonCount,
		RecordedAt:        item.RecordedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func (s *service) loadLessonRecordRelations(ctx context.Context, readDB *dao.Query, items []*model.LessonRecord) (map[int32]*model.Student, map[int32]*model.SysUser, error) {
	studentIDs := make([]int32, 0, len(items))
	teacherIDs := make([]int32, 0, len(items))
	seenStudents := map[int32]struct{}{}
	seenTeachers := map[int32]struct{}{}
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seenStudents[item.StudentID]; !ok {
			seenStudents[item.StudentID] = struct{}{}
			studentIDs = append(studentIDs, item.StudentID)
		}
		if _, ok := seenTeachers[item.TeacherID]; !ok {
			seenTeachers[item.TeacherID] = struct{}{}
			teacherIDs = append(teacherIDs, item.TeacherID)
		}
	}
	studentMap := map[int32]*model.Student{}
	teacherMap := map[int32]*model.SysUser{}
	if len(studentIDs) > 0 {
		students, err := readDB.WithContext(ctx).Student.Where(readDB.Student.ID.In(studentIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range students {
			studentMap[item.ID] = item
		}
	}
	if len(teacherIDs) > 0 {
		teachers, err := readDB.WithContext(ctx).SysUser.Where(readDB.SysUser.ID.In(teacherIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range teachers {
			teacherMap[item.ID] = item
		}
	}
	return studentMap, teacherMap, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

var _ = serviceutil.Int32Ptr
