package tenant

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"gorm.io/gorm"
)

type orgAggregateRow struct {
	OrganizationID int32 `gorm:"column:organization_id"`
	UsedCampuses   int64 `gorm:"column:used_campuses"`
	UsedUsers      int64 `gorm:"column:used_users"`
}

type campusAggregateRow struct {
	CampusID   int32 `gorm:"column:campus_id"`
	CountValue int64 `gorm:"column:count_value"`
}

type OrgCampusSummary struct {
	ID                 int32      `json:"id"`
	OrganizationID     int32      `json:"organizationId,omitempty"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Status             string     `json:"status"`
	SubscriptionStatus string     `json:"subscriptionStatus,omitempty"`
	PlanCode           string     `json:"planCode,omitempty"`
	EditionCode        string     `json:"editionCode,omitempty"`
	EndsAt             *time.Time `json:"endsAt,omitempty"`
	DaysRemaining      int32      `json:"daysRemaining,omitempty"`
	MaxCampuses        int32      `json:"maxCampuses,omitempty"`
	MaxUsers           int32      `json:"maxUsers,omitempty"`
	UsedCampuses       int64      `json:"usedCampuses,omitempty"`
	UsedUsers          int64      `json:"usedUsers,omitempty"`
	FeatureFlags       []string   `json:"featureFlags,omitempty"`
	HealthLevel        string     `json:"healthLevel,omitempty"`
	FollowUpStatus     string     `json:"followUpStatus,omitempty"`
	FollowUpOwner      string     `json:"followUpOwner,omitempty"`
	FollowUpRemark     string     `json:"followUpRemark,omitempty"`
}

type CreateOrganizationRequest struct {
	OrgName            string `json:"orgName" binding:"required"`
	Status             string `json:"status"`
	SubscriptionStatus string `json:"subscriptionStatus"`
	PlanCode           string `json:"planCode"`
	EndsAt             string `json:"endsAt"`
	MaxCampuses        int32  `json:"maxCampuses"`
	MaxUsers           int32  `json:"maxUsers"`
	Timezone           string `json:"timezone"`
	Remark             string `json:"remark"`
}

type UpdateOrganizationRequest struct {
	OrgName            string `json:"orgName" binding:"required"`
	Status             string `json:"status"`
	SubscriptionStatus string `json:"subscriptionStatus"`
	PlanCode           string `json:"planCode"`
	EndsAt             string `json:"endsAt"`
	MaxCampuses        int32  `json:"maxCampuses"`
	MaxUsers           int32  `json:"maxUsers"`
	Timezone           string `json:"timezone"`
	Remark             string `json:"remark"`
}

type CreateCampusRequest struct {
	OrganizationID int32  `json:"organizationId"`
	CampusName     string `json:"campusName" binding:"required"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

type UpdateCampusRequest struct {
	OrganizationID int32  `json:"organizationId"`
	CampusName     string `json:"campusName" binding:"required"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

type ManagementService interface {
	ListOrganizations(ctx context.Context, actor proposal.SessionUserInfo) ([]OrgCampusSummary, error)
	CreateOrganization(ctx context.Context, actor proposal.SessionUserInfo, req *CreateOrganizationRequest) (int32, error)
	UpdateOrganization(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateOrganizationRequest) error
	ListCampuses(ctx context.Context, actor proposal.SessionUserInfo) ([]OrgCampusSummary, error)
	CreateCampus(ctx context.Context, actor proposal.SessionUserInfo, req *CreateCampusRequest) (int32, error)
	UpdateCampus(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateCampusRequest) error
	GetOrganizationSettings(ctx context.Context, actor proposal.SessionUserInfo) (*OrganizationSettings, error)
	UpdateOrganizationSettings(ctx context.Context, actor proposal.SessionUserInfo, req *UpdateOrganizationSettingsRequest) error
}

type managementService struct {
	db mysql.Repo
}

func NewManagement(db mysql.Repo) ManagementService {
	return &managementService{db: db}
}

func (s *managementService) ListOrganizations(ctx context.Context, actor proposal.SessionUserInfo) ([]OrgCampusSummary, error) {
	if !actor.IsPlatformRole() {
		return nil, apperr.Forbidden("platform permission is required")
	}
	var items []model.Organization
	if err := s.db.GetDbR().WithContext(ctx).Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	var subscriptions []model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	aggregates, err := s.loadOrgAggregates(ctx)
	if err != nil {
		return nil, err
	}
	subscriptionByOrg := make(map[int32]model.Subscription, len(subscriptions))
	for _, item := range subscriptions {
		subscriptionByOrg[item.OrganizationID] = item
	}
	result := make([]OrgCampusSummary, 0, len(items))
	for _, item := range items {
		agg := aggregates[item.ID]
		summary := OrgCampusSummary{ID: item.ID, Code: item.OrgCode, Name: item.OrgName, Status: item.Status, SubscriptionStatus: item.SubscriptionStatus, EditionCode: item.EditionCode, UsedCampuses: agg.UsedCampuses, UsedUsers: agg.UsedUsers}
		if subscription, ok := subscriptionByOrg[item.ID]; ok {
			summary.PlanCode = subscription.PlanCode
			summary.MaxCampuses = subscription.MaxCampuses
			summary.MaxUsers = subscription.MaxUsers
			summary.SubscriptionStatus = subscription.Status
			summary.EndsAt = subscription.EndsAt
			summary.DaysRemaining = calcDaysRemaining(subscription.EndsAt)
			summary.FeatureFlags = mergeFeatureFlags(planFeatureFlags(subscription.PlanCode), parseFeatureFlags(subscription.FeatureFlags))
			summary.HealthLevel = buildHealthLevel(subscription.Status, summary.DaysRemaining, summary.UsedUsers, summary.MaxUsers, summary.UsedCampuses, summary.MaxCampuses)
			summary.FollowUpStatus = fallbackString(subscription.FollowUpStatus, buildFollowUpStatus(subscription.Status, summary.DaysRemaining))
			summary.FollowUpOwner = strings.TrimSpace(subscription.FollowUpOwner)
			summary.FollowUpRemark = buildFollowUpRemark(subscription.Status, summary.DaysRemaining, summary.UsedUsers, summary.MaxUsers, summary.UsedCampuses, summary.MaxCampuses)
		} else {
			summary.PlanCode = fallbackString(strings.TrimSpace(item.EditionCode), "standard")
			summary.FeatureFlags = planFeatureFlags(summary.PlanCode)
			summary.HealthLevel = buildHealthLevel(summary.SubscriptionStatus, summary.DaysRemaining, summary.UsedUsers, summary.MaxUsers, summary.UsedCampuses, summary.MaxCampuses)
			summary.FollowUpStatus = buildFollowUpStatus(summary.SubscriptionStatus, summary.DaysRemaining)
			summary.FollowUpRemark = buildFollowUpRemark(summary.SubscriptionStatus, summary.DaysRemaining, summary.UsedUsers, summary.MaxUsers, summary.UsedCampuses, summary.MaxCampuses)
		}
		result = append(result, summary)
	}
	return result, nil
}

func (s *managementService) loadOrgAggregates(ctx context.Context) (map[int32]orgAggregateRow, error) {
	query := `
		SELECT o.id AS organization_id,
		       COALESCE(c.used_campuses, 0) AS used_campuses,
		       COALESCE(u.used_users, 0) AS used_users
		FROM organization o
		LEFT JOIN (
			SELECT organization_id, COUNT(*) AS used_campuses
			FROM campus
			GROUP BY organization_id
		) c ON c.organization_id = o.id
		LEFT JOIN (
			SELECT organization_id, COUNT(*) AS used_users
			FROM sys_user
			WHERE organization_id IS NOT NULL AND organization_id > 0
			GROUP BY organization_id
		) u ON u.organization_id = o.id
	`
	var rows []orgAggregateRow
	if err := s.db.GetDbR().WithContext(ctx).Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int32]orgAggregateRow, len(rows))
	for _, row := range rows {
		result[row.OrganizationID] = row
	}
	return result, nil
}

func (s *managementService) CreateOrganization(ctx context.Context, actor proposal.SessionUserInfo, req *CreateOrganizationRequest) (int32, error) {
	if !actor.IsPlatformAdmin() {
		return 0, apperr.Forbidden("platform admin permission is required")
	}

	planCode := normalizePlanCode(req.PlanCode)
	item := &model.Organization{
		OrgCode:            s.generateOrganizationCode(),
		OrgName:            req.OrgName,
		Status:             defaultStatus(req.Status),
		SubscriptionStatus: defaultSubscriptionStatus(req.SubscriptionStatus),
		EditionCode:        planCode,
		Timezone:           defaultString(req.Timezone, "Asia/Shanghai"),
		Remark:             req.Remark,
	}
	if err := s.db.GetDbW().WithContext(ctx).Create(item).Error; err != nil {
		return 0, err
	}

	endsAt, err := parseEndsAt(req.EndsAt)
	if err != nil {
		return 0, err
	}
	if err := s.db.GetDbW().WithContext(ctx).Create(&model.Subscription{
		OrganizationID: item.ID,
		PlanCode:       planCode,
		Status:         defaultSubscriptionStatus(req.SubscriptionStatus),
		EndsAt:         endsAt,
		FeatureFlags:   stringPtr(marshalFeatureFlags(planFeatureFlags(planCode))),
		MaxCampuses:    defaultPositive(req.MaxCampuses, defaultPlanMaxCampuses(planCode)),
		MaxUsers:       defaultPositive(req.MaxUsers, defaultPlanMaxUsers(planCode)),
		Remark:         req.Remark,
	}).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *managementService) UpdateOrganization(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateOrganizationRequest) error {
	if !actor.IsPlatformAdmin() {
		return apperr.Forbidden("platform admin permission is required")
	}
	planCode := normalizePlanCode(req.PlanCode)
	updates := map[string]any{
		"org_name":            req.OrgName,
		"status":              defaultStatus(req.Status),
		"edition_code":        planCode,
		"subscription_status": defaultSubscriptionStatus(req.SubscriptionStatus),
		"timezone":            defaultString(req.Timezone, "Asia/Shanghai"),
		"remark":              req.Remark,
	}
	if err := s.db.GetDbW().WithContext(ctx).Model(&model.Organization{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	endsAt, err := parseEndsAt(req.EndsAt)
	if err != nil {
		return err
	}
	subscriptionUpdates := map[string]any{
		"plan_code":     planCode,
		"status":        defaultSubscriptionStatus(req.SubscriptionStatus),
		"feature_flags": marshalFeatureFlags(planFeatureFlags(planCode)),
		"max_campuses":  defaultPositive(req.MaxCampuses, defaultPlanMaxCampuses(planCode)),
		"max_users":     defaultPositive(req.MaxUsers, defaultPlanMaxUsers(planCode)),
		"remark":        req.Remark,
	}
	if endsAt != nil {
		subscriptionUpdates["ends_at"] = endsAt
	}
	if err := s.db.GetDbW().WithContext(ctx).Model(&model.Subscription{}).Where("organization_id = ?", id).Updates(subscriptionUpdates).Error; err != nil {
		return err
	}
	return nil
}

func (s *managementService) ListCampuses(ctx context.Context, actor proposal.SessionUserInfo) ([]OrgCampusSummary, error) {
	var items []model.Campus
	query := s.db.GetDbR().WithContext(ctx).Order("id asc")
	orgFilter := []int32(nil)
	if actor.IsPlatformRole() {
		// no extra filter
	} else if actor.IsOrgAdmin() {
		if actor.OrganizationID <= 0 {
			return nil, apperr.Forbidden("actor organization scope is missing")
		}
		orgFilter = []int32{actor.OrganizationID}
		query = query.Where("organization_id = ?", actor.OrganizationID)
	} else if actor.IsCampusAdmin() {
		if actor.OrganizationID <= 0 || actor.CampusID <= 0 {
			return nil, apperr.Forbidden("actor campus scope is missing")
		}
		orgFilter = []int32{actor.OrganizationID}
		query = query.Where("organization_id = ? AND id = ?", actor.OrganizationID, actor.CampusID)
	} else {
		return nil, apperr.Forbidden("admin permission is required")
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []OrgCampusSummary{}, nil
	}
	if orgFilter == nil {
		orgSet := make(map[int32]struct{}, len(items))
		for _, item := range items {
			orgSet[item.OrganizationID] = struct{}{}
		}
		orgFilter = make([]int32, 0, len(orgSet))
		for orgID := range orgSet {
			orgFilter = append(orgFilter, orgID)
		}
	}
	var organizations []model.Organization
	if err := s.db.GetDbR().WithContext(ctx).Where("id IN ?", orgFilter).Find(&organizations).Error; err != nil {
		return nil, err
	}
	var subscriptions []model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Where("organization_id IN ?", orgFilter).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	activeUsersByCampus, lowLessonByCampus, arrearsByCampus, rescheduleByCampus, err := s.loadCampusAggregates(ctx, actor)
	if err != nil {
		return nil, err
	}
	orgMap := make(map[int32]model.Organization, len(organizations))
	for _, item := range organizations {
		orgMap[item.ID] = item
	}
	subscriptionByOrg := make(map[int32]model.Subscription, len(subscriptions))
	for _, item := range subscriptions {
		subscriptionByOrg[item.OrganizationID] = item
	}
	result := make([]OrgCampusSummary, 0, len(items))
	for _, item := range items {
		flags := []string(nil)
		planCode := "standard"
		if subscription, ok := subscriptionByOrg[item.OrganizationID]; ok {
			planCode = subscription.PlanCode
			flags = planFeatureFlags(subscription.PlanCode)
		} else if org, ok := orgMap[item.OrganizationID]; ok {
			planCode = org.EditionCode
			flags = planFeatureFlags(org.EditionCode)
		}
		activeUsers := activeUsersByCampus[item.ID]
		lowLessonCount := lowLessonByCampus[item.ID]
		arrearsCount := arrearsByCampus[item.ID]
		rescheduleCount := rescheduleByCampus[item.ID]
		riskCount := lowLessonCount + arrearsCount + rescheduleCount
		result = append(result, OrgCampusSummary{ID: item.ID, OrganizationID: item.OrganizationID, Code: item.CampusCode, Name: item.CampusName, Status: item.Status, FeatureFlags: flags, PlanCode: planCode, UsedUsers: activeUsers, DaysRemaining: int32(riskCount), HealthLevel: buildCampusHealthLevel(item.Status, int32(riskCount), activeUsers)})
	}
	return result, nil
}

func (s *managementService) loadCampusAggregates(ctx context.Context, actor proposal.SessionUserInfo) (map[int32]int64, map[int32]int64, map[int32]int64, map[int32]int64, error) {
	whereOrg := ""
	args := make([]any, 0, 1)
	if actor.IsOrgAdmin() || actor.IsCampusAdmin() {
		if actor.OrganizationID <= 0 {
			return nil, nil, nil, nil, apperr.Forbidden("actor organization scope is missing")
		}
		whereOrg = " AND organization_id = ?"
		args = append(args, actor.OrganizationID)
	}
	activeUsersByCampus, err := s.scanCampusAggregate(ctx, `
		SELECT campus_id, COUNT(*) AS count_value
		FROM sys_user
		WHERE campus_id IS NOT NULL AND campus_id > 0 AND status = ?`+whereOrg+`
		GROUP BY campus_id
	`, append([]any{proposal.UserStatusEnabled}, args...)...)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	lowLessonByCampus, err := s.scanCampusAggregate(ctx, `
		SELECT campus_id, COUNT(DISTINCT student_id) AS count_value
		FROM lesson_package
		WHERE campus_id > 0 AND remain_lessons <= low_lesson_threshold`+whereOrg+`
		GROUP BY campus_id
	`, args...)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	arrearsByCampus, err := s.scanCampusAggregate(ctx, `
		SELECT campus_id, COUNT(DISTINCT student_id) AS count_value
		FROM lesson_package
		WHERE campus_id > 0 AND paid_amount < total_amount`+whereOrg+`
		GROUP BY campus_id
	`, args...)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rescheduleByCampus, err := s.scanCampusAggregate(ctx, `
		SELECT campus_id, COUNT(*) AS count_value
		FROM reschedule_record
		WHERE campus_id > 0`+whereOrg+`
		GROUP BY campus_id
	`, args...)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return activeUsersByCampus, lowLessonByCampus, arrearsByCampus, rescheduleByCampus, nil
}

func (s *managementService) scanCampusAggregate(ctx context.Context, query string, args ...any) (map[int32]int64, error) {
	var rows []campusAggregateRow
	if err := s.db.GetDbR().WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int32]int64, len(rows))
	for _, row := range rows {
		result[row.CampusID] = row.CountValue
	}
	return result, nil
}

func (s *managementService) CreateCampus(ctx context.Context, actor proposal.SessionUserInfo, req *CreateCampusRequest) (int32, error) {
	if !actor.IsPlatformAdmin() && !actor.IsOrgAdmin() {
		return 0, apperr.Forbidden("admin permission is required")
	}
	orgID := req.OrganizationID
	if actor.IsOrgAdmin() {
		orgID = actor.OrganizationID
	}
	if orgID <= 0 {
		return 0, apperr.InvalidArgument("organizationId is required")
	}
	if err := EnsureTenantWriteAllowedForOrganization(ctx, New(s.db), orgID); err != nil {
		return 0, err
	}
	subscription, err := s.loadSubscription(ctx, orgID)
	if err != nil {
		return 0, err
	}
	if subscription != nil && subscription.MaxCampuses > 0 {
		var campusCount int64
		if err := s.db.GetDbR().WithContext(ctx).Model(&model.Campus{}).Where("organization_id = ?", orgID).Count(&campusCount).Error; err != nil {
			return 0, err
		}
		if campusCount >= int64(subscription.MaxCampuses) {
			return 0, apperr.Forbidden("当前套餐校区数已达上限，请升级套餐")
		}
	}
	item := &model.Campus{OrganizationID: orgID, CampusCode: s.generateCampusCode(orgID), CampusName: req.CampusName, Status: defaultStatus(req.Status), Remark: req.Remark}
	if err := s.db.GetDbW().WithContext(ctx).Create(item).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *managementService) UpdateCampus(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateCampusRequest) error {
	if !actor.IsPlatformAdmin() && !actor.IsOrgAdmin() {
		return apperr.Forbidden("admin permission is required")
	}
	var current model.Campus
	if err := s.db.GetDbR().WithContext(ctx).First(&current, id).Error; err != nil {
		return err
	}
	if actor.IsOrgAdmin() && current.OrganizationID != actor.OrganizationID {
		return apperr.Forbidden("organization scope mismatch")
	}
	updates := map[string]any{"organization_id": current.OrganizationID, "campus_name": req.CampusName, "status": defaultStatus(req.Status), "remark": req.Remark}
	if actor.IsPlatformAdmin() && req.OrganizationID > 0 {
		updates["organization_id"] = req.OrganizationID
	}
	return s.db.GetDbW().WithContext(ctx).Model(&model.Campus{}).Where("id = ?", id).Updates(updates).Error
}

func (s *managementService) loadSubscription(ctx context.Context, organizationID int32) (*model.Subscription, error) {
	if organizationID <= 0 {
		return nil, nil
	}
	var subscription model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Where("organization_id = ?", organizationID).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func defaultStatus(status string) string {
	if status == "inactive" {
		return status
	}
	return "active"
}

func defaultSubscriptionStatus(status string) string {
	switch status {
	case proposal.SubscriptionStatusTrial, proposal.SubscriptionStatusActive, proposal.SubscriptionStatusPastDue, proposal.SubscriptionStatusSuspended, proposal.SubscriptionStatusExpired:
		return status
	default:
		return proposal.SubscriptionStatusActive
	}
}

func normalizePlanCode(planCode string) string {
	switch strings.ToLower(strings.TrimSpace(planCode)) {
	case "trial", "standard", "pro", "enterprise", "platform":
		return strings.ToLower(strings.TrimSpace(planCode))
	default:
		return "standard"
	}
}

func planFeatureFlags(planCode string) []string {
	switch normalizePlanCode(planCode) {
	case "trial":
		return []string{"auth", "user_management", "student", "course", "lesson_package", "payment", "schedule", "lesson_record", "reschedule", "subscription_center"}
	case "pro":
		return []string{"auth", "user_management", "student", "course", "lesson_package", "payment", "schedule", "lesson_record", "reschedule", "subscription_center", "audit_export"}
	case "enterprise":
		return []string{"auth", "user_management", "student", "course", "lesson_package", "payment", "schedule", "lesson_record", "reschedule", "subscription_center", "audit_export", "platform_management", "recovery_ops"}
	case "platform":
		return []string{"auth", "user_management", "student", "course", "lesson_package", "payment", "schedule", "lesson_record", "reschedule", "subscription_center", "audit_export", "platform_management", "recovery_ops"}
	default:
		return []string{"auth", "user_management", "student", "course", "lesson_package", "payment", "schedule", "lesson_record", "reschedule", "subscription_center"}
	}
}

func defaultPlanMaxCampuses(planCode string) int32 {
	switch normalizePlanCode(planCode) {
	case "trial":
		return 1
	case "pro":
		return 3
	case "enterprise":
		return 10
	case "platform":
		return 100
	default:
		return 1
	}
}

func buildCampusHealthLevel(status string, riskCount int32, activeUsers int64) string {
	if status != "active" {
		return "high"
	}
	if riskCount >= 10 {
		return "high"
	}
	if riskCount >= 3 || activeUsers == 0 {
		return "medium"
	}
	return "healthy"
}

func defaultPlanMaxUsers(planCode string) int32 {
	switch normalizePlanCode(planCode) {
	case "trial":
		return 20
	case "pro":
		return 100
	case "enterprise":
		return 500
	case "platform":
		return 5000
	default:
		return 50
	}
}

func defaultPositive(value, fallback int32) int32 {
	if value > 0 {
		return value
	}
	return fallback
}

func parseEndsAt(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	endsAt, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
	if err != nil {
		endsAt, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
		if err != nil {
			return nil, apperr.InvalidArgument("endsAt format must be 2006-01-02 or 2006-01-02 15:04:05")
		}
	}
	return &endsAt, nil
}

func stringPtr(value string) *string {
	return &value
}

func (s *managementService) generateOrganizationCode() string {
	return "ORG" + time.Now().Format("20060102150405")
}

func (s *managementService) generateCampusCode(organizationID int32) string {
	return "CAMPUS" + time.Now().Format("20060102150405")
}

func calcDaysRemaining(endsAt *time.Time) int32 {
	if endsAt == nil {
		return 0
	}
	now := time.Now()
	if endsAt.Before(now) {
		return 0
	}
	duration := endsAt.Sub(now)
	days := int32(duration.Hours() / 24)
	if duration.Hours() > float64(days*24) {
		days++
	}
	return days
}
