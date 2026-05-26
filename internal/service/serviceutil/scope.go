package serviceutil

import (
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"gorm.io/gorm"
)

// EnsureSameOrganization rejects cross-tenant access for non-platform actors.
func EnsureSameOrganization(actor proposal.SessionUserInfo, organizationID int32) error {
	if actor.IsPlatformAdmin() {
		return nil
	}
	if actor.OrganizationID <= 0 {
		return apperr.Forbidden("actor organization scope is missing")
	}
	if organizationID <= 0 || organizationID != actor.OrganizationID {
		return apperr.Forbidden("resource does not belong to current organization")
	}
	return nil
}

// EnsureSameCampus rejects cross-campus access for campus-scoped actors.
func EnsureSameCampus(actor proposal.SessionUserInfo, campusID int32) error {
	if actor.IsPlatformAdmin() || (actor.DataScope != proposal.DataScopeCampus && actor.DataScope != proposal.DataScopeSelf) {
		return nil
	}
	if actor.CampusID <= 0 {
		return apperr.Forbidden("actor campus scope is missing")
	}
	if campusID <= 0 || campusID != actor.CampusID {
		return apperr.Forbidden("resource does not belong to current campus")
	}
	return nil
}

// EnsureTenantAccess validates organization and campus data scope for a tenant-owned resource.
func EnsureTenantAccess(actor proposal.SessionUserInfo, organizationID, campusID int32) error {
	if err := EnsureSameOrganization(actor, organizationID); err != nil {
		return err
	}
	return EnsureSameCampus(actor, campusID)
}

// ApplyTenantGormScope applies organization/campus restrictions to a GORM query.
// Non-platform actors must have an organization scope; campus/self scoped actors must
// also have a campus scope when the target resource is campus-owned.
func ApplyTenantGormScope(actor proposal.SessionUserInfo, db *gorm.DB, organizationColumn, campusColumn string) (*gorm.DB, error) {
	if db == nil || actor.IsPlatformAdmin() {
		return db, nil
	}
	if actor.OrganizationID <= 0 {
		return nil, apperr.Forbidden("actor organization scope is missing")
	}
	if organizationColumn == "" {
		organizationColumn = "organization_id"
	}
	db = db.Where(organizationColumn+" = ?", actor.OrganizationID)
	if actor.DataScope == proposal.DataScopeCampus || actor.DataScope == proposal.DataScopeSelf {
		if actor.CampusID <= 0 {
			return nil, apperr.Forbidden("actor campus scope is missing")
		}
		if campusColumn == "" {
			campusColumn = "campus_id"
		}
		db = db.Where(campusColumn+" = ?", actor.CampusID)
	}
	return db, nil
}

func EnsureStudentAccess(actor proposal.SessionUserInfo, item *model.Student) error {
	if item == nil {
		return apperr.NotFound("student not found")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsureCourseAccess(actor proposal.SessionUserInfo, item *model.Course) error {
	if item == nil {
		return apperr.NotFound("course not found")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsureLessonPackageAccess(actor proposal.SessionUserInfo, item *model.LessonPackage) error {
	if item == nil {
		return apperr.NotFound("lesson package not found")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsurePaymentRecordAccess(actor proposal.SessionUserInfo, item *model.PaymentRecord) error {
	if item == nil {
		return apperr.NotFound("payment record not found")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsureScheduleAccess(actor proposal.SessionUserInfo, item *model.Schedule) error {
	if item == nil {
		return apperr.NotFound("schedule not found")
	}
	if actor.IsTeacher() && item.TeacherID != actor.Id {
		return apperr.Forbidden("schedule does not belong to current teacher")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsureLessonRecordAccess(actor proposal.SessionUserInfo, item *model.LessonRecord) error {
	if item == nil {
		return apperr.NotFound("lesson record not found")
	}
	if actor.IsTeacher() && item.TeacherID != actor.Id {
		return apperr.Forbidden("lesson record does not belong to current teacher")
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}

func EnsureRescheduleRecordAccess(actor proposal.SessionUserInfo, item *model.RescheduleRecord) error {
	if item == nil {
		return apperr.NotFound("reschedule record not found")
	}
	if actor.IsTeacher() {
		if item.OperatorID == nil || *item.OperatorID != actor.Id {
			return apperr.Forbidden("reschedule record does not belong to current teacher")
		}
		return nil
	}
	return EnsureTenantAccess(actor, item.OrganizationID, item.CampusID)
}
