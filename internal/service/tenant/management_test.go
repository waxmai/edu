package tenant

import (
	"context"
	"path/filepath"
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"

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

func TestListCampusesScopesOrgAdminToOwnOrganization(t *testing.T) {
	svc, db := newManagementTestService(t)
	seedManagementCampuses(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}

	items, err := svc.ListCampuses(context.Background(), actor)
	if err != nil {
		t.Fatalf("ListCampuses() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].OrganizationID != 1 || items[0].ID != 10 {
		t.Fatalf("item = %+v, want org 1 campus 10", items[0])
	}
	if items[0].PlanCode != "pro" {
		t.Fatalf("PlanCode = %q, want pro", items[0].PlanCode)
	}
}

func TestListCampusesRequiresOrgScopeForOrgAdmin(t *testing.T) {
	svc, _ := newManagementTestService(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, DataScope: proposal.DataScopeOrg}

	_, err := svc.ListCampuses(context.Background(), actor)
	if err == nil {
		t.Fatal("expected missing organization scope to be rejected")
	}
}

func newManagementTestService(t *testing.T) (*managementService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.Organization{}, &model.Subscription{}, &model.Campus{}, &model.SysUser{}, &model.LessonPackage{}, &model.RescheduleRecord{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return &managementService{db: testDBRepo{db: db}}, db
}

func seedManagementCampuses(t *testing.T, db *gorm.DB) {
	t.Helper()
	organizations := []model.Organization{
		{ID: 1, OrgCode: "ORG1", OrgName: "org1", Status: "active", SubscriptionStatus: proposal.SubscriptionStatusActive, EditionCode: "standard", Timezone: "Asia/Shanghai"},
		{ID: 2, OrgCode: "ORG2", OrgName: "org2", Status: "active", SubscriptionStatus: proposal.SubscriptionStatusActive, EditionCode: "enterprise", Timezone: "Asia/Shanghai"},
	}
	if err := db.Create(&organizations).Error; err != nil {
		t.Fatalf("seed organizations error = %v", err)
	}
	subscriptions := []model.Subscription{
		{OrganizationID: 1, PlanCode: "pro", Status: proposal.SubscriptionStatusActive, MaxCampuses: 3, MaxUsers: 100},
		{OrganizationID: 2, PlanCode: "enterprise", Status: proposal.SubscriptionStatusActive, MaxCampuses: 10, MaxUsers: 500},
	}
	if err := db.Create(&subscriptions).Error; err != nil {
		t.Fatalf("seed subscriptions error = %v", err)
	}
	campuses := []model.Campus{
		{ID: 10, OrganizationID: 1, CampusCode: "C1", CampusName: "org1 campus", Status: "active"},
		{ID: 20, OrganizationID: 2, CampusCode: "C2", CampusName: "org2 campus", Status: "active"},
	}
	if err := db.Create(&campuses).Error; err != nil {
		t.Fatalf("seed campuses error = %v", err)
	}
}
