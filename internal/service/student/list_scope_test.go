package student

import (
	"context"
	"path/filepath"
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/dto"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListScopesStudentsToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeStudents(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("items = %#v, want only student 1", items)
	}
}

func TestListInactiveAlertScopesStudentsToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeStudents(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx, dto.StudentListQuery{InactiveAlert: true})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("items = %#v, want only student 1", items)
	}
}

func TestGetByIDRejectsCrossTenantStudent(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeStudents(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	_, err := svc.GetByID(ctx, 2)
	if err == nil {
		t.Fatal("GetByID() error = nil, want cross-tenant access denied/not found")
	}
}

func newListScopeTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "student-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.Student{}, &model.LessonRecord{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: &testRepo{db: db}}, db
}

func seedListScopeStudents(t *testing.T, db *gorm.DB) {
	t.Helper()
	students := []model.Student{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentName: "org1 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentName: "org2 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
	}
	if err := db.Create(&students).Error; err != nil {
		t.Fatalf("seed students error = %v", err)
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
