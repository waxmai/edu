package notification

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
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

func TestDBRecorderWritesNotificationRecord(t *testing.T) {
	db := newNotificationTestDB(t, "notification.db")
	recorder := NewDBRecorder(testDBRepo{db: db})
	if recorder == nil {
		t.Fatal("NewDBRecorder() = nil")
	}
	now := time.Now()
	if err := recorder.Record(context.Background(), DeliveryRecord{Scene: ScenePasswordRecovery, Channel: ChannelEmail, TemplateCode: "password_recovery", OrganizationID: 12, Provider: "smtp", TargetMasked: "u***@example.com", Status: "accepted", IdempotencyKey: "challenge-1", RecordedAt: now}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	var item model.NotificationRecord
	if err := db.First(&item).Error; err != nil {
		t.Fatalf("load notification record error = %v", err)
	}
	if item.OrganizationID != 12 || item.Scene != ScenePasswordRecovery || item.TargetMasked != "u***@example.com" || item.RequestID != "challenge-1" {
		t.Fatalf("record = %#v", item)
	}
}

func TestQueryServiceListFiltersFailuresOnly(t *testing.T) {
	db := newNotificationTestDB(t, "notification-query.db")
	now := time.Now()
	rows := []model.NotificationRecord{
		{OrganizationID: 1, Channel: ChannelEmail, Scene: ScenePasswordRecovery, TemplateCode: "password_recovery", TargetMasked: "a***@x.com", Provider: "smtp", Status: "accepted", Attempts: 1, CreatedAt: now},
		{OrganizationID: 1, Channel: ChannelSMS, Scene: ScenePasswordRecovery, TemplateCode: "password_recovery", TargetMasked: "138****8000", Provider: "sms", Status: "error", ErrorCode: ErrorTimeout, ErrorMessage: "timeout", Attempts: 1, CreatedAt: now.Add(time.Second)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed notification records error = %v", err)
	}
	service := NewQueryService(testDBRepo{db: db})
	ctx := servicectx.WithActor(context.Background(), proposal.SessionUserInfo{RoleCode: proposal.RolePlatformAdmin, DataScope: proposal.DataScopeAll})
	resp, err := service.List(ctx, Query{FailuresOnly: true})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].Status != "error" || resp.Items[0].ErrorCode != ErrorTimeout {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestQueryServiceListScopesNonPlatformActorToOwnOrganization(t *testing.T) {
	db := newNotificationTestDB(t, "notification-scope.db")
	now := time.Now()
	rows := []model.NotificationRecord{
		{OrganizationID: 1, Channel: ChannelEmail, Scene: ScenePasswordRecovery, TemplateCode: "password_recovery", TargetMasked: "a***@x.com", Provider: "smtp", Status: "accepted", Attempts: 1, CreatedAt: now},
		{OrganizationID: 2, Channel: ChannelEmail, Scene: ScenePasswordRecovery, TemplateCode: "password_recovery", TargetMasked: "b***@x.com", Provider: "smtp", Status: "accepted", Attempts: 1, CreatedAt: now.Add(time.Second)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed notification records error = %v", err)
	}
	service := NewQueryService(testDBRepo{db: db})
	ctx := servicectx.WithActor(context.Background(), proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg})
	resp, err := service.List(ctx, Query{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].OrganizationID != 1 {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestQueryServiceListRejectsCrossOrganizationFilter(t *testing.T) {
	db := newNotificationTestDB(t, "notification-cross-org.db")
	service := NewQueryService(testDBRepo{db: db})
	ctx := servicectx.WithActor(context.Background(), proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg})
	_, err := service.List(ctx, Query{OrganizationID: 2})
	if err == nil {
		t.Fatal("expected cross-organization filter to be rejected")
	}
}

func newNotificationTestDB(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), filename)), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&model.NotificationRecord{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
