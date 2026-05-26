package schedule

import (
	"context"
	"fmt"
	"strings"

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

func conflictMessage(teacherConflict, studentConflict bool) string {
	return fmt.Sprintf("当前时间段存在排课冲突 [teacher=%t student=%t]", teacherConflict, studentConflict)
}

func validateScheduleRefsAndConflicts(ctx context.Context, q *dao.Query, excludeID int32, item *model.Schedule) error {
	actor := servicectx.ActorFromContext(ctx)
	student, err := q.WithContext(ctx).Student.Where(q.Student.ID.Eq(item.StudentID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperr.NotFound("student not found")
		}
		return err
	}
	if err := serviceutil.EnsureStudentAccess(actor, student); err != nil {
		return err
	}
	if item.OrganizationID <= 0 {
		item.OrganizationID = student.OrganizationID
	}
	if item.CampusID <= 0 {
		item.CampusID = student.CampusID
	}
	course, err := q.WithContext(ctx).Course.Where(q.Course.ID.Eq(item.CourseID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperr.NotFound("course not found")
		}
		return err
	}
	if err := serviceutil.EnsureCourseAccess(actor, course); err != nil {
		return err
	}
	teacher, err := q.WithContext(ctx).SysUser.Where(q.SysUser.ID.Eq(item.TeacherID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperr.NotFound("teacher not found")
		}
		return err
	}
	if strings.TrimSpace(teacher.RoleCode) != "teacher" {
		return apperr.InvalidArgument("teacherId must reference a teacher user")
	}
	teacherOrgID := serviceutil.Int32PtrValue(teacher.OrganizationID)
	teacherCampusID := serviceutil.Int32PtrValue(teacher.CampusID)
	if err := serviceutil.EnsureTenantAccess(actor, teacherOrgID, teacherCampusID); err != nil {
		return err
	}
	if student.OrganizationID != course.OrganizationID || student.OrganizationID != teacherOrgID || student.CampusID != course.CampusID || student.CampusID != teacherCampusID {
		return apperr.Conflict("schedule references must belong to the same organization and campus")
	}
	if item.OrganizationID != student.OrganizationID || item.CampusID != student.CampusID {
		return apperr.Conflict("schedule scope does not match student scope")
	}
	if item.LessonPackageID != nil && *item.LessonPackageID > 0 {
		pkg, err := q.WithContext(ctx).LessonPackage.Where(q.LessonPackage.ID.Eq(*item.LessonPackageID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.NotFound("lesson package not found")
			}
			return err
		}
		if err := serviceutil.EnsureLessonPackageAccess(actor, pkg); err != nil {
			return err
		}
		if pkg.StudentID != item.StudentID {
			return apperr.Conflict("lesson package does not belong to student")
		}
		if pkg.OrganizationID != item.OrganizationID || pkg.CampusID != item.CampusID {
			return apperr.Conflict("lesson package scope does not match schedule scope")
		}
		if pkg.Status == "pending" {
			return apperr.Conflict("linked lesson package is not active yet")
		}
	}
	teacherConflict, studentConflict, err := detectConflicts(ctx, q, excludeID, item)
	if err != nil {
		return err
	}
	if teacherConflict || studentConflict {
		return apperr.Conflict(conflictMessage(teacherConflict, studentConflict))
	}
	return nil
}

func detectConflicts(ctx context.Context, q *dao.Query, excludeID int32, item *model.Schedule) (bool, bool, error) {
	base := []gen.Condition{
		q.Schedule.StartTime.Lt(item.EndTime),
		q.Schedule.EndTime.Gt(item.StartTime),
		q.Schedule.ScheduleStatus.NotIn("cancelled", "rescheduled"),
	}
	if excludeID > 0 {
		base = append(base, q.Schedule.ID.Neq(excludeID))
	}
	teacherQuery := q.WithContext(ctx).Schedule.Clauses(clause.Locking{Strength: "UPDATE"}).Where(append(base, q.Schedule.TeacherID.Eq(item.TeacherID))...)
	teacherCount, err := teacherQuery.Count()
	if err != nil {
		return false, false, err
	}
	studentQuery := q.WithContext(ctx).Schedule.Clauses(clause.Locking{Strength: "UPDATE"}).Where(append(base, q.Schedule.StudentID.Eq(item.StudentID))...)
	studentCount, err := studentQuery.Count()
	if err != nil {
		return false, false, err
	}
	return teacherCount > 0, studentCount > 0, nil
}

func buildScheduleConditions(q *dao.Query, query dto.ScheduleListQuery) []gen.Condition {
	conditions := make([]gen.Condition, 0, 6)
	if query.StudentID > 0 {
		conditions = append(conditions, q.Schedule.StudentID.Eq(query.StudentID))
	}
	if query.TeacherID > 0 {
		conditions = append(conditions, q.Schedule.TeacherID.Eq(query.TeacherID))
	}
	if query.LessonPackageID > 0 {
		conditions = append(conditions, q.Schedule.LessonPackageID.Eq(query.LessonPackageID))
	}
	if status := strings.TrimSpace(query.ScheduleStatus); status != "" {
		conditions = append(conditions, q.Schedule.ScheduleStatus.Eq(status))
	}
	if start, end, err := parseScheduleDateRange(query.StartDate, query.EndDate); err == nil {
		if !start.IsZero() {
			conditions = append(conditions, q.Schedule.ClassDate.Gte(start))
		}
		if !end.IsZero() {
			conditions = append(conditions, q.Schedule.ClassDate.Lte(end))
		}
	}
	return conditions
}

func applyLessonPackageUsageOnScheduleTransition(ctx context.Context, tx *dao.Query, current, next *model.Schedule) error {
	if current.LessonPackageID == nil && next.LessonPackageID == nil {
		return nil
	}
	wasCompleted := current.ScheduleStatus == "completed"
	willCompleted := next.ScheduleStatus == "completed"
	if wasCompleted == willCompleted && pointerInt32Equal(current.LessonPackageID, next.LessonPackageID) {
		return nil
	}
	if wasCompleted {
		if err := adjustLessonPackageLessons(ctx, tx, current.LessonPackageID, decimal.NewFromInt(-1)); err != nil {
			return err
		}
	}
	if willCompleted {
		if err := adjustLessonPackageLessons(ctx, tx, next.LessonPackageID, decimal.NewFromInt(1)); err != nil {
			if wasCompleted {
				_ = adjustLessonPackageLessons(ctx, tx, current.LessonPackageID, decimal.NewFromInt(1))
			}
			return err
		}
	}
	return nil
}

func adjustLessonPackageLessons(ctx context.Context, tx *dao.Query, lessonPackageID *int32, delta decimal.Decimal) error {
	if lessonPackageID == nil || *lessonPackageID <= 0 || delta.IsZero() {
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
		return apperr.Conflict("lesson package lessons are insufficient for completed schedule transition")
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
