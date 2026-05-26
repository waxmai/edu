package payment_record

import (
	"context"
	"path/filepath"
	"testing"
	"time"

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
	seedListScopePayments(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	items, err := svc.List(ctx, dto.PaymentRecordListQuery{StudentID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant student filter", len(items))
	}

	items, err = svc.List(ctx, dto.PaymentRecordListQuery{LessonPackageID: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0 for cross-tenant lesson package filter", len(items))
	}
}

func TestIncomeStatisticsScopesToActorTenant(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopePayments(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	stats, err := svc.IncomeStatistics(ctx, dto.PaymentIncomeStatsQuery{})
	if err != nil {
		t.Fatalf("IncomeStatistics() error = %v", err)
	}
	if stats.RecordCount != 1 || stats.TotalAmount != 100 {
		t.Fatalf("stats = %+v, want org1 paid amount/count only", stats)
	}
}

func TestIncomeStatisticsRequiresActorOrganization(t *testing.T) {
	svc, db := newListScopeTestService(t)
	seedListScopePayments(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, DataScope: proposal.DataScopeOrg}
	ctx := servicectx.WithActor(context.Background(), actor)

	_, err := svc.IncomeStatistics(ctx, dto.PaymentIncomeStatsQuery{})
	if err == nil {
		t.Fatal("IncomeStatistics() error = nil, want forbidden")
	}
}

func newListScopeTestService(t *testing.T) (*service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "payment-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.PaymentRecord{}, &model.Student{}, &model.LessonPackage{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &service{db: &testRepo{db: db}}, db
}

func seedListScopePayments(t *testing.T, db *gorm.DB) {
	t.Helper()
	students := []model.Student{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentName: "org1 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentName: "org2 student", Subject: "math", TeachingType: "one_to_one", Status: "active"},
	}
	if err := db.Create(&students).Error; err != nil {
		t.Fatalf("seed students error = %v", err)
	}
	packages := []model.LessonPackage{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentID: 1, CourseID: 1, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(10), LowLessonThreshold: decimal.NewFromInt(2), Status: "active"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentID: 2, CourseID: 2, TotalLessons: decimal.NewFromInt(10), RemainLessons: decimal.NewFromInt(10), LowLessonThreshold: decimal.NewFromInt(2), Status: "active"},
	}
	if err := db.Create(&packages).Error; err != nil {
		t.Fatalf("seed lesson packages error = %v", err)
	}
	records := []model.PaymentRecord{
		{ID: 1, OrganizationID: 1, CampusID: 10, StudentID: 1, LessonPackageID: 1, PaymentType: "signup", Amount: decimal.NewFromInt(100), PaymentMethod: "cash", PaymentTime: time.Now(), PaymentStatus: "paid"},
		{ID: 2, OrganizationID: 2, CampusID: 20, StudentID: 2, LessonPackageID: 2, PaymentType: "signup", Amount: decimal.NewFromInt(100), PaymentMethod: "cash", PaymentTime: time.Now(), PaymentStatus: "paid"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("seed payment records error = %v", err)
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
