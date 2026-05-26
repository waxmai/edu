package lesson_package

import (
	"context"
	"path/filepath"
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/dto"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListScopesRelatedEntityFiltersToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopeLessonPackages(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx, dto.LessonPackageListQuery{StudentID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant student filter", len(items))
	}

	items, err = svc.List(ctx, dto.LessonPackageListQuery{CourseID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant course filter", len(items))
	}
}

func newListScopeTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "lesson-package-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.LessonPackage{}, &model.Student{}, &model.Course{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: &testRepo{db: db}}, db
}

func seedListScopeLessonPackages(t *testing.T, db *gorm.DB) {
	t.Helper()
	students := []model.Student{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentName: "org1 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentName: "org2 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
	}
	if err := db.Create(&students).Error; err != nil {
		t.Fatalf("seed students error = %v", err)
	}
	courses := []model.Course{
		{ID: 1, OrganizationID: 1, CampusID: 10, CourseName: "org1 course", Subject: "math", CourseType: "regular", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, CourseName: "org2 course", Subject: "math", CourseType: "regular", Status: "active"},
	}
	if err := db.Create(&courses).Error; err != nil {
		t.Fatalf("seed courses error = %v", err)
	}
	packages := []model.LessonPackage{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentID: 1, CourseID: 1, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(10), LowLessonThreshold: decimal.NewFromInt(2), Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentID: 2, CourseID: 2, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(10), LowLessonThreshold: decimal.NewFromInt(2), Status: "active"},
	}
	if err := db.Create(&packages).Error; err != nil {
		t.Fatalf("seed lesson packages error = %v", err)
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
