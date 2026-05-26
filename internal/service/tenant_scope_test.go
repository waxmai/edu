package service

import (
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/dao"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestScopeStudentsDenyWhenOrgScopeMissing(t *testing.T) {
	q := newScopeTestQuery(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, DataScope: proposal.DataScopeOrg}

	conds := ScopeStudents(actor, q, nil)
	if len(conds) != 1 {
		t.Fatalf("len(conds) = %d, want deny condition only", len(conds))
	}
}

func TestScopeCampusOwnedResourcesTreatSelfScopeAsCampusScoped(t *testing.T) {
	q := newScopeTestQuery(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleTeacher, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeSelf}

	tests := []struct {
		name string
		got  int
	}{
		{name: "students", got: len(ScopeStudents(actor, q, nil))},
		{name: "courses", got: len(ScopeCourses(actor, q, nil))},
		{name: "lesson packages", got: len(ScopeLessonPackages(actor, q, nil))},
		{name: "payment records", got: len(ScopePaymentRecords(actor, q, nil))},
		{name: "schedules", got: len(ScopeSchedules(actor, q, nil))},
		{name: "lesson records", got: len(ScopeLessonRecords(actor, q, nil))},
		{name: "reschedule records", got: len(ScopeRescheduleRecords(actor, q, nil))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got < 2 {
				t.Fatalf("condition count = %d, want at least org + campus scope", tt.got)
			}
		})
	}
}

func newScopeTestQuery(t *testing.T) *dao.Query {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	return dao.Use(db)
}
