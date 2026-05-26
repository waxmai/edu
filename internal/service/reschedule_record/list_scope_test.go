package reschedule_record

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/dto"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListScopesRescheduleRecordsToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeRescheduleRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("items = %#v, want only reschedule record 1", items)
	}
}

func TestListScopesStudentFilterToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeRescheduleRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx, dto.RescheduleRecordListQuery{StudentID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant student filter", len(items))
	}
}

func TestGetByIDRejectsCrossTenantRescheduleRecord(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeRescheduleRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	_, err := svc.GetByID(ctx, 2)
	if err == nil {
		t.Fatal("GetByID() error = nil, want cross-tenant access denied/not found")
	}
}

func newListScopeTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "reschedule-record-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.RescheduleRecord{}, &model.Schedule{}, &model.SysUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: &testRepo{db: db}}, db
}

func seedListScopeRescheduleRecords(t *testing.T, db *gorm.DB) {
	t.Helper()
	org1, org2 := int32(1), int32(2)
	campus10, campus20 := int32(10), int32(20)
	operators := []model.SysUser{
		{ID: 1, Username: "operator1", RealName: "operator1", RoleCode: proposal.RoleOrgAdmin, OrganizationID: &org1, CampusID: &campus10, Status: proposal.UserStatusEnabled},
		{ID: 2, Username: "operator2", RealName: "operator2", RoleCode: proposal.RoleOrgAdmin, OrganizationID: &org2, CampusID: &campus20, Status: proposal.UserStatusEnabled},
	}
	if err := db.Create(&operators).Error; err != nil {
		t.Fatalf("seed operators error = %v", err)
	}
	now := time.Now()
	schedules := []model.Schedule{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentID: 1, CourseID: 1, TeacherID: 1, ClassDate: now, StartTime: now, EndTime: now.Add(time.Hour), ScheduleStatus: "scheduled"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentID: 2, CourseID: 2, TeacherID: 2, ClassDate: now, StartTime: now, EndTime: now.Add(time.Hour), ScheduleStatus: "scheduled"},
	}
	if err := db.Create(&schedules).Error; err != nil {
		t.Fatalf("seed schedules error = %v", err)
	}
	records := []model.RescheduleRecord{
		{ID: 1, OrganizationID: 1, CampusID: 10, OldScheduleID: 1, OperationType: "leave", Reason: "org1", OperatorID: &org1},
		{ID: 2, OrganizationID: 2, CampusID: 20, OldScheduleID: 2, OperationType: "leave", Reason: "org2", OperatorID: &org2},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("seed reschedule records error = %v", err)
	}
}

type testRepo struct{ db *gorm.DB }

func (r *testRepo) GetDbR() *gorm.DB           { return r.db }
func (r *testRepo) GetDbW() *gorm.DB           { return r.db }
func (r *testRepo) DbRClose() error            { return nil }
func (r *testRepo) DbWClose() error            { return nil }
func (r *testRepo) Ping(context.Context) error { return nil }
func (r *testRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
