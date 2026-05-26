package tenant

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type orgUsageAggregate struct {
	OrganizationID int32 `gorm:"column:organization_id"`
	UsedUsers      int64 `gorm:"column:used_users"`
	UsedCampuses   int64 `gorm:"column:used_campuses"`
}

type FeatureEntitlement struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type SubscriptionOverview struct {
	OrganizationID      int32                `json:"organizationId,omitempty"`
	OrganizationName    string               `json:"organizationName,omitempty"`
	PlanCode            string               `json:"planCode"`
	PlanName            string               `json:"planName"`
	Status              string               `json:"status"`
	StartsAt            *string              `json:"startsAt,omitempty"`
	EndsAt              *string              `json:"endsAt,omitempty"`
	RemainingDays       int32                `json:"remainingDays"`
	MaxUsers            int32                `json:"maxUsers"`
	UsedUsers           int64                `json:"usedUsers"`
	MaxCampuses         int32                `json:"maxCampuses"`
	UsedCampuses        int64                `json:"usedCampuses"`
	FeatureFlags        []string             `json:"featureFlags,omitempty"`
	FeatureEntitlements []FeatureEntitlement `json:"featureEntitlements,omitempty"`
	EditionName         string               `json:"editionName"`
	RenewalHint         string               `json:"renewalHint"`
	UpgradeHint         string               `json:"upgradeHint"`
}

type SubscriptionPipelineItem struct {
	OrganizationID     int32    `json:"organizationId"`
	OrganizationName   string   `json:"organizationName"`
	SubscriptionStatus string   `json:"subscriptionStatus"`
	PlanCode           string   `json:"planCode"`
	PlanName           string   `json:"planName"`
	EndsAt             *string  `json:"endsAt,omitempty"`
	RemainingDays      int32    `json:"remainingDays"`
	UsedUsers          int64    `json:"usedUsers"`
	MaxUsers           int32    `json:"maxUsers"`
	UsedCampuses       int64    `json:"usedCampuses"`
	MaxCampuses        int32    `json:"maxCampuses"`
	HealthLevel        string   `json:"healthLevel"`
	FollowUpStatus     string   `json:"followUpStatus"`
	FollowUpOwner      string   `json:"followUpOwner"`
	FollowUpNote       string   `json:"followUpNote"`
	LastContactAt      *string  `json:"lastContactAt,omitempty"`
	FollowUpRemark     string   `json:"followUpRemark"`
	FeatureFlags       []string `json:"featureFlags,omitempty"`
	EditionName        string   `json:"editionName"`
}

type UpdateSubscriptionFollowUpRequest struct {
	FollowUpStatus string `json:"followUpStatus" binding:"required"`
	FollowUpOwner  string `json:"followUpOwner"`
	FollowUpNote   string `json:"followUpNote"`
	LastContactAt  string `json:"lastContactAt"`
}

type SubscriptionQueryService interface {
	GetCurrentSubscription(ctx context.Context, actor proposal.SessionUserInfo) (*SubscriptionOverview, error)
	ListSubscriptionPipeline(ctx context.Context, actor proposal.SessionUserInfo) ([]SubscriptionPipelineItem, error)
	UpdateSubscriptionFollowUp(ctx context.Context, actor proposal.SessionUserInfo, organizationID int32, req *UpdateSubscriptionFollowUpRequest) error
}

type subscriptionQueryService struct {
	db mysql.Repo
}

func NewSubscriptionQuery(db mysql.Repo) SubscriptionQueryService {
	return &subscriptionQueryService{db: db}
}

func (s *subscriptionQueryService) GetCurrentSubscription(ctx context.Context, actor proposal.SessionUserInfo) (*SubscriptionOverview, error) {
	if actor.OrganizationID <= 0 {
		return nil, apperr.Forbidden("organization subscription is unavailable for current role")
	}
	return s.loadSubscriptionOverview(ctx, actor.OrganizationID)
}

func (s *subscriptionQueryService) ListSubscriptionPipeline(ctx context.Context, actor proposal.SessionUserInfo) ([]SubscriptionPipelineItem, error) {
	if !actor.IsPlatformRole() {
		return nil, apperr.Forbidden("platform permission is required")
	}

	var orgs []model.Organization
	if err := s.db.GetDbR().WithContext(ctx).Order("id asc").Find(&orgs).Error; err != nil {
		return nil, err
	}
	var subscriptions []model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	aggregates, err := s.loadOrgUsageAggregates(ctx)
	if err != nil {
		return nil, err
	}

	subscriptionByOrg := make(map[int32]model.Subscription, len(subscriptions))
	for _, item := range subscriptions {
		subscriptionByOrg[item.OrganizationID] = item
	}

	result := make([]SubscriptionPipelineItem, 0, len(orgs))
	for _, org := range orgs {
		subscription := subscriptionByOrg[org.ID]
		status := org.SubscriptionStatus
		if subscription.Status != "" {
			status = subscription.Status
		}
		agg := aggregates[org.ID]
		remainingDays := calcDaysRemaining(subscription.EndsAt)
		usedUsers := agg.UsedUsers
		usedCampuses := agg.UsedCampuses
		planCode := fallbackString(strings.TrimSpace(subscription.PlanCode), strings.TrimSpace(org.EditionCode), "standard")
		result = append(result, SubscriptionPipelineItem{
			OrganizationID:     org.ID,
			OrganizationName:   org.OrgName,
			SubscriptionStatus: status,
			PlanCode:           planCode,
			PlanName:           formatPlanName(planCode),
			EndsAt:             formatTimePointer(subscription.EndsAt),
			RemainingDays:      remainingDays,
			UsedUsers:          usedUsers,
			MaxUsers:           subscription.MaxUsers,
			UsedCampuses:       usedCampuses,
			MaxCampuses:        subscription.MaxCampuses,
			HealthLevel:        buildHealthLevel(status, remainingDays, usedUsers, subscription.MaxUsers, usedCampuses, subscription.MaxCampuses),
			FollowUpStatus:     fallbackString(subscription.FollowUpStatus, buildFollowUpStatus(status, remainingDays)),
			FollowUpOwner:      strings.TrimSpace(subscription.FollowUpOwner),
			FollowUpNote:       strings.TrimSpace(subscription.FollowUpNote),
			LastContactAt:      formatTimePointer(subscription.LastContactAt),
			FollowUpRemark:     buildFollowUpRemark(status, remainingDays, usedUsers, subscription.MaxUsers, usedCampuses, subscription.MaxCampuses),
			FeatureFlags:       mergeFeatureFlags(parseFeatureFlags(org.FeatureFlags), parseFeatureFlags(subscription.FeatureFlags)),
			EditionName:        formatEditionName(org.EditionCode),
		})
	}
	return result, nil
}

func (s *subscriptionQueryService) loadOrgUsageAggregates(ctx context.Context) (map[int32]orgUsageAggregate, error) {
	query := `
		SELECT o.id AS organization_id,
		       COALESCE(u.used_users, 0) AS used_users,
		       COALESCE(c.used_campuses, 0) AS used_campuses
		FROM organization o
		LEFT JOIN (
			SELECT organization_id, COUNT(*) AS used_users
			FROM sys_user
			WHERE organization_id IS NOT NULL AND organization_id > 0
			GROUP BY organization_id
		) u ON u.organization_id = o.id
		LEFT JOIN (
			SELECT organization_id, COUNT(*) AS used_campuses
			FROM campus
			GROUP BY organization_id
		) c ON c.organization_id = o.id
	`
	var rows []orgUsageAggregate
	if err := s.db.GetDbR().WithContext(ctx).Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int32]orgUsageAggregate, len(rows))
	for _, row := range rows {
		result[row.OrganizationID] = row
	}
	return result, nil
}

func (s *subscriptionQueryService) loadSubscriptionOverview(ctx context.Context, organizationID int32) (*SubscriptionOverview, error) {
	var org model.Organization
	if err := s.db.GetDbR().WithContext(ctx).First(&org, organizationID).Error; err != nil {
		return nil, err
	}
	var subscription model.Subscription
	if err := s.db.GetDbR().WithContext(ctx).Where("organization_id = ?", organizationID).First(&subscription).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	var usedUsers int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&model.SysUser{}).Where("organization_id = ?", organizationID).Count(&usedUsers).Error; err != nil {
		return nil, err
	}
	var usedCampuses int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&model.Campus{}).Where("organization_id = ?", organizationID).Count(&usedCampuses).Error; err != nil {
		return nil, err
	}

	planCode := fallbackString(strings.TrimSpace(subscription.PlanCode), strings.TrimSpace(org.EditionCode), "standard")
	status := fallbackString(strings.TrimSpace(subscription.Status), strings.TrimSpace(org.SubscriptionStatus), proposal.SubscriptionStatusActive)
	remainingDays := calcDaysRemaining(subscription.EndsAt)
	featureFlags := mergeFeatureFlags(parseFeatureFlags(org.FeatureFlags), parseFeatureFlags(subscription.FeatureFlags))
	return &SubscriptionOverview{
		OrganizationID:      org.ID,
		OrganizationName:    org.OrgName,
		PlanCode:            planCode,
		PlanName:            formatPlanName(planCode),
		Status:              status,
		StartsAt:            formatTimePointer(subscription.StartsAt),
		EndsAt:              formatTimePointer(subscription.EndsAt),
		RemainingDays:       remainingDays,
		MaxUsers:            subscription.MaxUsers,
		UsedUsers:           usedUsers,
		MaxCampuses:         subscription.MaxCampuses,
		UsedCampuses:        usedCampuses,
		FeatureFlags:        featureFlags,
		FeatureEntitlements: buildFeatureEntitlements(featureFlags),
		EditionName:         formatEditionName(org.EditionCode),
		RenewalHint:         buildRenewalHint(status, remainingDays),
		UpgradeHint:         buildUpgradeHint(usedUsers, subscription.MaxUsers, usedCampuses, subscription.MaxCampuses),
	}, nil
}

func (s *subscriptionQueryService) UpdateSubscriptionFollowUp(ctx context.Context, actor proposal.SessionUserInfo, organizationID int32, req *UpdateSubscriptionFollowUpRequest) error {
	if !actor.IsPlatformAdmin() {
		return apperr.Forbidden("platform admin permission is required")
	}
	if organizationID <= 0 {
		return apperr.InvalidArgument("organization id must be positive")
	}
	if req == nil || strings.TrimSpace(req.FollowUpStatus) == "" {
		return apperr.InvalidArgument("followUpStatus is required")
	}
	updates := map[string]any{
		"follow_up_status": strings.TrimSpace(req.FollowUpStatus),
		"follow_up_owner":  strings.TrimSpace(req.FollowUpOwner),
		"follow_up_note":   strings.TrimSpace(req.FollowUpNote),
	}
	if strings.TrimSpace(req.LastContactAt) == "" {
		updates["last_contact_at"] = nil
	} else {
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(req.LastContactAt), time.Local)
		if err != nil {
			parsed, err = time.ParseInLocation("2006-01-02 15:04", strings.TrimSpace(req.LastContactAt), time.Local)
			if err != nil {
				parsed, err = time.ParseInLocation("2006-01-02", strings.TrimSpace(req.LastContactAt), time.Local)
				if err != nil {
					return apperr.InvalidArgument("lastContactAt format must be 2006-01-02 or 2006-01-02 15:04[:05]")
				}
			}
		}
		updates["last_contact_at"] = parsed
	}
	if err := s.db.GetDbW().WithContext(ctx).Model(&model.Subscription{}).Where("organization_id = ?", organizationID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func buildFeatureEntitlements(enabledFlags []string) []FeatureEntitlement {
	enabled := make(map[string]bool, len(enabledFlags))
	for _, flag := range enabledFlags {
		enabled[strings.TrimSpace(flag)] = true
	}
	catalog := []FeatureEntitlement{
		{Code: "auth", Name: "认证登录", Description: "账号登录、会话管理与基础安全能力"},
		{Code: "user_management", Name: "用户管理", Description: "维护机构账号、校区管理员和教师账号"},
		{Code: "student", Name: "学员管理", Description: "维护学员档案、家长联系方式与学员状态"},
		{Code: "course", Name: "课程管理", Description: "配置课程定义、科目与课程类型"},
		{Code: "lesson_package", Name: "课时包管理", Description: "管理购课课包、课时余额与消耗"},
		{Code: "payment", Name: "收费管理", Description: "记录收费、欠费与支付流水"},
		{Code: "schedule", Name: "排课管理", Description: "维护课程安排、教师与学员排课"},
		{Code: "lesson_record", Name: "上课记录", Description: "记录上课出勤、课消与课堂结果"},
		{Code: "reschedule", Name: "调补课管理", Description: "处理请假、调课、补课等运营流程"},
		{Code: "subscription_center", Name: "订阅中心", Description: "查看当前套餐、有效期、配额与权益"},
		{Code: "audit_export", Name: "审计导出", Description: "查看并导出关键审计操作日志"},
		{Code: "recovery_ops", Name: "恢复治理", Description: "账号恢复失败告警与安全治理能力"},
		{Code: "platform_management", Name: "平台管理", Description: "平台级机构、套餐和运营治理能力"},
	}
	for i := range catalog {
		catalog[i].Enabled = enabled[catalog[i].Code]
	}
	return catalog
}

func formatTimePointer(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	value := t.Format("2006-01-02 15:04:05")
	return &value
}

func formatPlanName(planCode string) string {
	switch strings.ToLower(strings.TrimSpace(planCode)) {
	case "trial":
		return "试用版"
	case "standard":
		return "标准版"
	case "pro":
		return "专业版"
	case "enterprise":
		return "旗舰版"
	case "platform":
		return "平台版"
	default:
		if strings.TrimSpace(planCode) == "" {
			return "未设置"
		}
		return planCode
	}
}

func formatEditionName(editionCode string) string {
	if strings.TrimSpace(editionCode) == "" {
		return "标准交付版"
	}
	return formatPlanName(editionCode)
}

func buildRenewalHint(status string, remainingDays int32) string {
	switch status {
	case proposal.SubscriptionStatusExpired:
		return "订阅已过期，关键写操作可能已被限制，请优先续费恢复。"
	case proposal.SubscriptionStatusPastDue:
		return "订阅待续费，建议尽快完成续费，避免影响日常运营。"
	case proposal.SubscriptionStatusSuspended:
		return "机构当前已停用，请联系平台管理员恢复服务。"
	case proposal.SubscriptionStatusTrial:
		if remainingDays > 0 {
			return "当前处于试用期，建议尽早确认正式套餐与有效期。"
		}
		return "试用期即将结束，建议尽快转为正式套餐。"
	default:
		if remainingDays > 0 && remainingDays <= 15 {
			return "当前订阅即将到期，建议提前安排续费，避免影响业务连续性。"
		}
		return "当前订阅状态正常，可持续关注配额与到期时间。"
	}
}

func buildUpgradeHint(usedUsers int64, maxUsers int32, usedCampuses int64, maxCampuses int32) string {
	messages := make([]string, 0, 2)
	if nearLimit(usedUsers, maxUsers) {
		messages = append(messages, "账号数接近上限，可考虑升级更高账号配额")
	}
	if nearLimit(usedCampuses, maxCampuses) {
		messages = append(messages, "校区数接近上限，可考虑升级支持更多校区的套餐")
	}
	if len(messages) == 0 {
		return "当前配额充足，可继续按现有套餐稳定运营。"
	}
	return strings.Join(messages, "；")
}

func buildHealthLevel(status string, remainingDays int32, usedUsers int64, maxUsers int32, usedCampuses int64, maxCampuses int32) string {
	if status == proposal.SubscriptionStatusExpired || status == proposal.SubscriptionStatusSuspended {
		return "high"
	}
	if status == proposal.SubscriptionStatusPastDue || (remainingDays > 0 && remainingDays <= 7) {
		return "high"
	}
	if nearLimit(usedUsers, maxUsers) || nearLimit(usedCampuses, maxCampuses) || (remainingDays > 0 && remainingDays <= 15) {
		return "medium"
	}
	return "healthy"
}

func buildFollowUpStatus(status string, remainingDays int32) string {
	switch {
	case status == proposal.SubscriptionStatusExpired:
		return "已过期待恢复"
	case status == proposal.SubscriptionStatusPastDue:
		return "待催续费"
	case status == proposal.SubscriptionStatusSuspended:
		return "停用待处理"
	case status == proposal.SubscriptionStatusTrial && remainingDays <= 7:
		return "试用转化跟进"
	case remainingDays > 0 && remainingDays <= 15:
		return "即将到期跟进"
	default:
		return "正常维护"
	}
}

func buildFollowUpRemark(status string, remainingDays int32, usedUsers int64, maxUsers int32, usedCampuses int64, maxCampuses int32) string {
	parts := make([]string, 0, 4)
	if status == proposal.SubscriptionStatusPastDue || status == proposal.SubscriptionStatusExpired {
		parts = append(parts, "优先联系负责人确认续费")
	}
	if remainingDays > 0 && remainingDays <= 15 {
		parts = append(parts, "到期前完成续费提醒")
	}
	if nearLimit(usedUsers, maxUsers) {
		parts = append(parts, "账号配额接近上限")
	}
	if nearLimit(usedCampuses, maxCampuses) {
		parts = append(parts, "校区配额接近上限")
	}
	if len(parts) == 0 {
		return "当前机构状态稳定，保持常规跟进即可。"
	}
	return strings.Join(parts, "；")
}

func nearLimit(used int64, max int32) bool {
	if max <= 0 {
		return false
	}
	ratio := decimal.NewFromInt(used).Div(decimal.NewFromInt32(max))
	return ratio.GreaterThanOrEqual(decimal.NewFromFloat(0.8))
}

func fallbackString(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}
