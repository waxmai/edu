package dashboard

import (
	"context"
	"path/filepath"
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testDBRepo struct {
	db *gorm.DB
}

func (r testDBRepo) GetDbR() *gorm.DB { return r.db }
func (r testDBRepo) GetDbW() *gorm.DB { return r.db }
func (r testDBRepo) DbRClose() error  { return nil }
func (r testDBRepo) DbWClose() error  { return nil }
func (r testDBRepo) Ping(ctx context.Context) error {
	return nil
}
func (r testDBRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestLoadOperationalAlertsScopesByOrganization(t *testing.T) {
	svc, db := newDashboardTestService(t)
	seedDashboardAlerts(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}

	lowLessonCount, arrearsCount, inactiveCount, err := svc.loadOperationalAlerts(context.Background(), actor)
	if err != nil {
		t.Fatalf("loadOperationalAlerts() error = %v", err)
	}
	if lowLessonCount != 1 || arrearsCount != 1 || inactiveCount != 1 {
		t.Fatalf("counts = low:%d arrears:%d inactive:%d, want 1/1/1", lowLessonCount, arrearsCount, inactiveCount)
	}
}

func TestLoadOperationalAlertsRequiresOrganizationScope(t *testing.T) {
	svc, _ := newDashboardTestService(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, DataScope: proposal.DataScopeOrg}

	_, _, _, err := svc.loadOperationalAlerts(context.Background(), actor)
	if err == nil {
		t.Fatal("expected missing organization scope to be rejected")
	}
}

func newDashboardTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "dashboard.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.LessonPackage{}, &model.Student{}, &model.LessonRecord{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: testDBRepo{db: db}}, db
}

func seedDashboardAlerts(t *testing.T, db *gorm.DB) {
	t.Helper()
	students := []model.Student{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentName: "org1-active", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentName: "org2-active", Status: "active"},
	}
	if err := db.Create(&students).Error; err != nil {
		t.Fatalf("seed students error = %v", err)
	}
	packages := []model.LessonPackage{
		{OrganizationID: 1, CampusID: 10, StudentID: 1, CourseID: 1, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(1), LowLessonThreshold: decimal.NewFromInt(2), TotalAmount: decimal.NewFromInt(100), PaidAmount: decimal.NewFromInt(50), Status: "active"},
		{OrganizationID: 2, CampusID: 20, StudentID: 2, CourseID: 1, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(1), LowLessonThreshold: decimal.NewFromInt(2), TotalAmount: decimal.NewFromInt(100), PaidAmount: decimal.NewFromInt(50), Status: "active"},
	}
	if err := db.Create(&packages).Error; err != nil {
		t.Fatalf("seed lesson packages error = %v", err)
	}
}
