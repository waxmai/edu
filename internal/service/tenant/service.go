package tenant

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	"gorm.io/gorm"
)

type EffectiveProfile struct {
	OrganizationID     int32
	OrganizationName   string
	OrganizationStatus string
	CampusID           int32
	CampusName         string
	CampusStatus       string
	SubscriptionStatus string
	PlanCode           string
	MaxCampuses        int32
	MaxUsers           int32
	FeatureFlags       []string
}

type Service interface {
	ResolveEffectiveProfile(ctx context.Context, user *model.SysUser) (*EffectiveProfile, error)
	GetSubscription(ctx context.Context, organizationID int32) (*model.Subscription, error)
}

type service struct {
	db mysql.Repo
}

func New(db mysql.Repo) Service {
	return &service{db: db}
}

func (s *service) ResolveEffectiveProfile(ctx context.Context, user *model.SysUser) (*EffectiveProfile, error) {
	profile := &EffectiveProfile{SubscriptionStatus: proposal.SubscriptionStatusActive, PlanCode: "platform", MaxCampuses: 99999, MaxUsers: 99999}
	if user == nil {
		return profile, nil
	}
	if user.OrganizationID == nil || *user.OrganizationID <= 0 {
		profile.FeatureFlags = []string{
			proposal.FeatureAuth,
			proposal.FeatureUserManagement,
			proposal.FeatureStudent,
			proposal.FeatureCourse,
			proposal.FeatureLessonPackage,
			proposal.FeaturePayment,
			proposal.FeatureSchedule,
			proposal.FeatureLessonRecord,
			proposal.FeatureReschedule,
			proposal.FeatureSubscriptionCenter,
			proposal.FeatureAuditExport,
			proposal.FeatureRecoveryOps,
			proposal.FeaturePlatformManagement,
			proposal.FeatureCustomerSuccess,
			proposal.FeaturePlatformReport,
		}
		return profile, nil
	}

	orgID := *user.OrganizationID
	var org model.Organization
	if err := s.db.GetDbR().WithContext(ctx).First(&org, orgID).Error; err != nil {
		return nil, err
	}
	profile.OrganizationID = org.ID
	profile.OrganizationName = org.OrgName
	profile.OrganizationStatus = org.Status
	profile.SubscriptionStatus = org.SubscriptionStatus
	profile.FeatureFlags = append(profile.FeatureFlags, parseFeatureFlags(org.FeatureFlags)...)

	var subscription model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Where("organization_id = ?", orgID).First(&subscription).Error; err == nil {
		profile.SubscriptionStatus = subscription.Status
		profile.PlanCode = subscription.PlanCode
		profile.MaxCampuses = subscription.MaxCampuses
		profile.MaxUsers = subscription.MaxUsers
		profile.FeatureFlags = mergeFeatureFlags(profile.FeatureFlags, parseFeatureFlags(subscription.FeatureFlags))
		if subscription.EndsAt != nil && subscription.EndsAt.Before(time.Now()) {
			profile.SubscriptionStatus = proposal.SubscriptionStatusExpired
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if user.CampusID != nil && *user.CampusID > 0 {
		var campus model.Campus
		if err := s.db.GetDbR().WithContext(ctx).First(&campus, *user.CampusID).Error; err != nil {
			return nil, err
		}
		profile.CampusID = campus.ID
		profile.CampusName = campus.CampusName
		profile.CampusStatus = campus.Status
		profile.FeatureFlags = mergeFeatureFlags(profile.FeatureFlags, parseFeatureFlags(campus.FeatureFlags))
	}

	if len(profile.FeatureFlags) == 0 {
		profile.FeatureFlags = []string{
			proposal.FeatureAuth,
			proposal.FeatureUserManagement,
			proposal.FeatureStudent,
			proposal.FeatureCourse,
			proposal.FeatureLessonPackage,
			proposal.FeaturePayment,
			proposal.FeatureSchedule,
			proposal.FeatureLessonRecord,
			proposal.FeatureReschedule,
			proposal.FeatureSubscriptionCenter,
		}
	}
	return profile, nil
}

func (s *service) GetSubscription(ctx context.Context, organizationID int32) (*model.Subscription, error) {
	if organizationID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var subscription model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Where("organization_id = ?", organizationID).First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

func parseFeatureFlags(raw *string) []string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	var flags []string
	if err := json.Unmarshal([]byte(*raw), &flags); err != nil {
		return nil
	}
	return flags
}

func mergeFeatureFlags(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base)+len(extra))
	merged := make([]string, 0, len(base)+len(extra))
	for _, item := range append(append([]string{}, base...), extra...) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}
