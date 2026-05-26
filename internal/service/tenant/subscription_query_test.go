package tenant

import (
	"context"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"

	"gorm.io/gorm"
)

func TestListSubscriptionPipelineRequiresPlatformRole(t *testing.T) {
	svc, _ := newSubscriptionQueryTestService(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}

	_, err := svc.ListSubscriptionPipeline(context.Background(), actor)
	if err == nil {
		t.Fatal("ListSubscriptionPipeline() error = nil, want forbidden")
	}
}

func TestListSubscriptionPipelineReturnsAllOrganizationsForPlatform(t *testing.T) {
	svc, db := newSubscriptionQueryTestService(t)
	seedSubscriptionQueryData(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RolePlatformAdmin, DataScope: proposal.DataScopeAll}

	items, err := svc.ListSubscriptionPipeline(context.Background(), actor)
	if err != nil {
		t.Fatalf("ListSubscriptionPipeline() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].OrganizationID != 1 || items[1].OrganizationID != 2 {
		t.Fatalf("items = %#v, want organizations 1 and 2", items)
	}
}

func TestGetCurrentSubscriptionUsesActorOrganization(t *testing.T) {
	svc, db := newSubscriptionQueryTestService(t)
	seedSubscriptionQueryData(t, db)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}

	overview, err := svc.GetCurrentSubscription(context.Background(), actor)
	if err != nil {
		t.Fatalf("GetCurrentSubscription() error = %v", err)
	}
	if overview.OrganizationID != 1 || overview.OrganizationName != "org1" {
		t.Fatalf("overview = %+v, want org1 only", overview)
	}
}

func TestGetCurrentSubscriptionRequiresActorOrganization(t *testing.T) {
	svc, _ := newSubscriptionQueryTestService(t)
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, DataScope: proposal.DataScopeOrg}

	_, err := svc.GetCurrentSubscription(context.Background(), actor)
	if err == nil {
		t.Fatal("GetCurrentSubscription() error = nil, want forbidden")
	}
}

func newSubscriptionQueryTestService(t *testing.T) (*subscriptionQueryService, *gorm.DB) {
	t.Helper()
	_, db := newManagementTestService(t)
	return &subscriptionQueryService{db: testDBRepo{db: db}}, db
}

func seedSubscriptionQueryData(t *testing.T, db *gorm.DB) {
	t.Helper()
	organizations := []model.Organization{
		{ID: 1, OrgCode: "ORG1", OrgName: "org1", Status: "active", SubscriptionStatus: proposal.SubscriptionStatusActive, EditionCode: "standard", Timezone: "Asia/Shanghai"},
		{ID: 2, OrgCode: "ORG2", OrgName: "org2", Status: "active", SubscriptionStatus: proposal.SubscriptionStatusTrial, EditionCode: "enterprise", Timezone: "Asia/Shanghai"},
	}
	if err := db.Create(&organizations).Error; err != nil {
		t.Fatalf("seed organizations error = %v", err)
	}
	endsAt := time.Now().AddDate(0, 0, 30)
	subscriptions := []model.Subscription{
		{OrganizationID: 1, PlanCode: "pro", Status: proposal.SubscriptionStatusActive, MaxCampuses: 3, MaxUsers: 100, EndsAt: &endsAt},
		{OrganizationID: 2, PlanCode: "enterprise", Status: proposal.SubscriptionStatusTrial, MaxCampuses: 10, MaxUsers: 500, EndsAt: &endsAt},
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
