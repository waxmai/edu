package service

import (
	"context"

	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/service/apperr"
	"gorm.io/gen"
)

func ensureTenantScope(actor proposal.SessionUserInfo, campusOwned bool) error {
	if actor.IsPlatformAdmin() {
		return nil
	}
	if actor.OrganizationID <= 0 {
		return apperr.Forbidden("actor organization scope is missing")
	}
	if campusOwned && (actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf) && actor.CampusID <= 0 {
		return apperr.Forbidden("actor campus scope is missing")
	}
	return nil
}

func ScopeStudents(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.Student.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.Student.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.Student.CampusID.Eq(actor.CampusID))
		}
	}
	return conds
}

func ScopeCourses(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.Course.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.Course.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.Course.CampusID.Eq(actor.CampusID))
		}
	}
	return conds
}

func ScopeLessonPackages(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.LessonPackage.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.LessonPackage.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.LessonPackage.CampusID.Eq(actor.CampusID))
		}
	}
	return conds
}

func ScopePaymentRecords(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.PaymentRecord.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.PaymentRecord.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.PaymentRecord.CampusID.Eq(actor.CampusID))
		}
	}
	return conds
}

func ScopeSchedules(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.Schedule.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.Schedule.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.Schedule.CampusID.Eq(actor.CampusID))
		}
	}
	if actor.IsTeacher() {
		conds = append(conds, q.Schedule.TeacherID.Eq(actor.Id))
	}
	return conds
}

func ScopeLessonRecords(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.LessonRecord.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.LessonRecord.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.LessonRecord.CampusID.Eq(actor.CampusID))
		}
	}
	if actor.IsTeacher() {
		conds = append(conds, q.LessonRecord.TeacherID.Eq(actor.Id))
	}
	return conds
}

func ScopeRescheduleRecords(actor proposal.SessionUserInfo, q *dao.Query, base []gen.Condition) []gen.Condition {
	conds := append([]gen.Condition{}, base...)
	if err := ensureTenantScope(actor, true); err != nil {
		conds = append(conds, q.RescheduleRecord.ID.Eq(-1))
		return conds
	}
	if !actor.IsPlatformAdmin() {
		conds = append(conds, q.RescheduleRecord.OrganizationID.Eq(actor.OrganizationID))
		if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
			conds = append(conds, q.RescheduleRecord.CampusID.Eq(actor.CampusID))
		}
	}
	if actor.IsTeacher() {
		conds = append(conds, q.RescheduleRecord.OperatorID.Eq(actor.Id))
	}
	return conds
}

func WithActor(ctx context.Context, actor proposal.SessionUserInfo) context.Context {
	return core.WithActor(ctx, actor)
}

func ActorFromContext(ctx context.Context) proposal.SessionUserInfo {
	return core.ActorFromContext(ctx)
}
