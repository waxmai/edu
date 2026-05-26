package tenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
)

type OrganizationSettings struct {
	OrganizationID     int32    `json:"organizationId"`
	OrganizationName   string   `json:"organizationName"`
	Timezone           string   `json:"timezone"`
	BrandName          string   `json:"brandName"`
	NotificationEmail  string   `json:"notificationEmail"`
	SecurityPolicy     []string `json:"securityPolicy,omitempty"`
	Remark             string   `json:"remark"`
	FeatureFlags       []string `json:"featureFlags,omitempty"`
	PlanCode           string   `json:"planCode,omitempty"`
	SubscriptionStatus string   `json:"subscriptionStatus,omitempty"`
}

type UpdateOrganizationSettingsRequest struct {
	Timezone          string   `json:"timezone"`
	BrandName         string   `json:"brandName"`
	NotificationEmail string   `json:"notificationEmail"`
	SecurityPolicy    []string `json:"securityPolicy"`
	Remark            string   `json:"remark"`
}

type organizationSettingsRow struct {
	ID                 int32
	OrgName            string
	Timezone           string
	BrandName          string
	NotificationEmail  string
	SecurityPolicyRaw  sql.NullString
	Remark             string
	FeatureFlagsRaw    sql.NullString
	EditionCode        string
	SubscriptionStatus string
}

func (s *managementService) GetOrganizationSettings(ctx context.Context, actor proposal.SessionUserInfo) (*OrganizationSettings, error) {
	orgID := actor.OrganizationID
	if actor.IsPlatformAdmin() {
		return nil, apperr.InvalidArgument("platform admin should use tenant management pages for organization settings")
	}
	if orgID <= 0 {
		return nil, apperr.Forbidden("organization settings unavailable for current role")
	}
	row, err := s.loadOrganizationSettingsRow(ctx, orgID)
	if err != nil {
		return nil, err
	}
	settings := &OrganizationSettings{
		OrganizationID:     row.ID,
		OrganizationName:   row.OrgName,
		Timezone:           row.Timezone,
		BrandName:          row.BrandName,
		NotificationEmail:  row.NotificationEmail,
		SecurityPolicy:     parseJSONStringSlice(row.SecurityPolicyRaw.String),
		Remark:             row.Remark,
		FeatureFlags:       parseJSONStringSlice(row.FeatureFlagsRaw.String),
		PlanCode:           strings.TrimSpace(row.EditionCode),
		SubscriptionStatus: strings.TrimSpace(row.SubscriptionStatus),
	}
	if subscription, err := s.loadSubscription(ctx, orgID); err == nil && subscription != nil {
		settings.FeatureFlags = mergeFeatureFlags(settings.FeatureFlags, parseFeatureFlags(subscription.FeatureFlags))
		if strings.TrimSpace(subscription.PlanCode) != "" {
			settings.PlanCode = subscription.PlanCode
		}
		if strings.TrimSpace(subscription.Status) != "" {
			settings.SubscriptionStatus = subscription.Status
		}
	}
	return settings, nil
}

func (s *managementService) UpdateOrganizationSettings(ctx context.Context, actor proposal.SessionUserInfo, req *UpdateOrganizationSettingsRequest) error {
	if actor.OrganizationID <= 0 || actor.IsPlatformAdmin() || actor.IsCampusAdmin() {
		return apperr.Forbidden("organization admin permission is required")
	}
	if err := EnsureTenantWriteAllowedForOrganization(ctx, New(s.db), actor.OrganizationID); err != nil {
		return err
	}
	updates := map[string]any{
		"timezone":           defaultString(strings.TrimSpace(req.Timezone), "Asia/Shanghai"),
		"brand_name":         strings.TrimSpace(req.BrandName),
		"notification_email": strings.TrimSpace(req.NotificationEmail),
		"security_policy":    marshalFeatureFlags(req.SecurityPolicy),
		"remark":             strings.TrimSpace(req.Remark),
	}
	return s.db.GetDbW().WithContext(ctx).Model(&model.Organization{}).Where("id = ?", actor.OrganizationID).Updates(updates).Error
}

func (s *managementService) loadOrganizationSettingsRow(ctx context.Context, orgID int32) (*organizationSettingsRow, error) {
	row := &organizationSettingsRow{}
	result := s.db.GetDbR().WithContext(ctx).Raw(`
SELECT id, org_name, timezone,
       COALESCE(brand_name, '') AS brand_name,
       COALESCE(notification_email, '') AS notification_email,
       security_policy,
       remark,
       feature_flags,
       edition_code,
       subscription_status
FROM organization
WHERE id = ?
LIMIT 1`, orgID).Scan(row)
	if result.Error != nil {
		return nil, result.Error
	}
	if row.ID <= 0 {
		return nil, sql.ErrNoRows
	}
	return row, nil
}

func parseJSONStringSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}
