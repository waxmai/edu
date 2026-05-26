package lesson_record

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

func TestListScopesLessonRecordsToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeLessonRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("items = %#v, want only lesson record 1", items)
	}
}

func TestListScopesRelatedEntityFiltersToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeLessonRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx, dto.LessonRecordListQuery{StudentID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant student filter", len(items))
	}

	items, err = svc.List(ctx, dto.LessonRecordListQuery{TeacherID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant teacher filter", len(items))
	}
}

func TestGetByIDRejectsCrossTenantLessonRecord(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeLessonRecords(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	_, err := svc.GetByID(ctx, 2)
	if err == nil {
		t.Fatal("GetByID() error = nil, want cross-tenant access denied/not found")
	}
}

func newListScopeTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "lesson-record-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.LessonRecord{}, &model.Student{}, &model.SysUser{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: &testRepo{db: db}}, db
}

func seedListScopeLessonRecords(t *testing.T, db *gorm.DB) {
	t.Helper()
	org1, org2 := int32(1), int32(2)
	campus10, campus20 := int32(10), int32(20)
	students := []model.Student{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentName: "org1 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentName: "org2 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
	}
	if err := db.Create(&students).Error; err != nil {
		t.Fatalf("seed students error = %v", err)
	}
	teachers := []model.SysUser{
		{ID: 1, Username: "teacher1", RealName: "teacher1", RoleCode: proposal.RoleTeacher, OrganizationID: &org1, CampusID: &campus10, Status: proposal.UserStatusEnabled},
		{ID: 2, Username: "teacher2", RealName: "teacher2", RoleCode: proposal.RoleTeacher, OrganizationID: &org2, CampusID: &campus20, Status: proposal.UserStatusEnabled},
	}
	if err := db.Create(&teachers).Error; err != nil {
		t.Fatalf("seed teachers error = %v", err)
	}
	now := time.Now()
	records := []model.LessonRecord{
		{ID: 1, OrganizationID: 1, CampusID: 10, ScheduleID: 1, StudentID: 1, TeacherID: 1, AttendanceStatus: "present", RecordedAt: now},
		{ID: 2, OrganizationID: 2, CampusID: 20, ScheduleID: 2, StudentID: 2, TeacherID: 2, AttendanceStatus: "present", RecordedAt: now.Add(time.Hour)},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("seed lesson records error = %v", err)
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
