package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/dto"
	lessonrecordsvc "edu-schedule-system/internal/service/lesson_record"
	paymentsvc "edu-schedule-system/internal/service/payment_record"
	reschedulesvc "edu-schedule-system/internal/service/reschedule_record"
	schedulesvc "edu-schedule-system/internal/service/schedule"
	"edu-schedule-system/internal/service/serviceutil"
	studentsvc "edu-schedule-system/internal/service/student"
	tenantservice "edu-schedule-system/internal/service/tenant"
)

type Service interface {
	Overview(ctx context.Context, actor proposal.SessionUserInfo, query dto.DashboardQuery) (*dto.DashboardResponse, error)
}

type service struct {
	db                mysql.Repo
	students          studentsvc.Service
	payments          paymentsvc.Service
	schedules         schedulesvc.Service
	lessonRecords     lessonrecordsvc.Service
	reschedules       reschedulesvc.Service
	subscriptionQuery tenantservice.SubscriptionQueryService
	tenantManagement  tenantservice.ManagementService
}

func New(
	db mysql.Repo,
	students studentsvc.Service,
	payments paymentsvc.Service,
	schedules schedulesvc.Service,
	lessonRecords lessonrecordsvc.Service,
	reschedules reschedulesvc.Service,
	subscriptionQuery tenantservice.SubscriptionQueryService,
	tenantManagement tenantservice.ManagementService,
) Service {
	return &service{
		db:                db,
		students:          students,
		payments:          payments,
		schedules:         schedules,
		lessonRecords:     lessonRecords,
		reschedules:       reschedules,
		subscriptionQuery: subscriptionQuery,
		tenantManagement:  tenantManagement,
	}
}

func (s *service) Overview(ctx context.Context, actor proposal.SessionUserInfo, query dto.DashboardQuery) (*dto.DashboardResponse, error) {
	ctx = servicectx.WithActor(ctx, actor)

	if actor.IsPlatformRole() {
		return s.platformOverview(ctx, actor)
	}
	if actor.IsTeacher() {
		return s.teacherOverview(ctx, actor)
	}
	return s.adminOverview(ctx, actor, query)
}

func (s *service) teacherOverview(ctx context.Context, actor proposal.SessionUserInfo) (*dto.DashboardResponse, error) {
	schedules, err := s.schedules.List(ctx, dto.ScheduleListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	lessonRecords, err := s.lessonRecords.List(ctx, dto.LessonRecordListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	reschedules, err := s.reschedules.List(ctx, dto.RescheduleRecordListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}

	resp := &dto.DashboardResponse{Role: actor.RoleCode}
	resp.SummaryCards = []dto.DashboardSummaryCard{
		{Label: "本周课程", Value: fmt.Sprintf("%d", len(schedules)), Desc: "当前账号可见课程数量"},
		{Label: "课堂记录", Value: fmt.Sprintf("%d", len(lessonRecords)), Desc: "最近已记录课堂次数"},
		{Label: "调课申请", Value: fmt.Sprintf("%d", len(reschedules)), Desc: "最近调课与请假处理量"},
		{Label: "登录身份", Value: "教师", Desc: "个人教学工作台"},
	}
	resp.QuickActions = []dto.DashboardQuickAction{{Title: "我的排课", Desc: "查看本周课程安排", Path: "/my-schedules"}, {Title: "上课记录", Desc: "查看最近课堂记录", Path: "/my-lesson-records"}, {Title: "调课申请", Desc: "处理调课与请假", Path: "/my-reschedules"}}
	resp.Schedules = mapSchedules(schedules)
	resp.LessonRecords = mapLessonRecords(lessonRecords)
	return resp, nil
}

func (s *service) platformOverview(ctx context.Context, actor proposal.SessionUserInfo) (*dto.DashboardResponse, error) {
	orgs, err := s.tenantManagement.ListOrganizations(ctx, actor)
	if err != nil {
		return nil, err
	}
	subs, err := s.subscriptionQuery.ListSubscriptionPipeline(ctx, actor)
	if err != nil {
		return nil, err
	}
	riskItems := make([]dto.DashboardPlatformRiskItem, 0)
	trialCount := 0
	totalUsers := int64(0)
	for _, item := range subs {
		if item.SubscriptionStatus == proposal.SubscriptionStatusTrial {
			trialCount++
		}
		if item.HealthLevel == "high" {
			riskItems = append(riskItems, dto.DashboardPlatformRiskItem{
				OrganizationName: item.OrganizationName,
				FollowUpStatus:   item.FollowUpStatus,
				RemainingDays:    item.RemainingDays,
				UsedUsers:        item.UsedUsers,
				MaxUsers:         item.MaxUsers,
				UsedCampuses:     item.UsedCampuses,
				MaxCampuses:      item.MaxCampuses,
				HealthLevel:      item.HealthLevel,
			})
		}
	}
	sortPlatformRisks(riskItems)
	for _, item := range orgs {
		totalUsers += item.UsedUsers
	}
	resp := &dto.DashboardResponse{Role: actor.RoleCode}
	if actor.IsPlatformAdmin() || actor.RoleCode == proposal.RolePlatformOps || actor.RoleCode == proposal.RolePlatformFinance {
		resp.SummaryCards = []dto.DashboardSummaryCard{
			{Label: "机构总数", Value: fmt.Sprintf("%d", len(orgs)), Desc: "当前平台已开通机构数"},
			{Label: "高风险机构", Value: fmt.Sprintf("%d", len(riskItems)), Desc: "续费或恢复需优先处理"},
			{Label: "试用机构", Value: fmt.Sprintf("%d", trialCount), Desc: "重点推进试用转化"},
			{Label: "平台账号", Value: fmt.Sprintf("%d", totalUsers), Desc: "全平台已开通账号数"},
		}
		resp.PlatformRisks = riskItems
		resp.MetricItems = []dto.DashboardMetricItem{{Label: "机构总数", Value: fmt.Sprintf("%d", len(orgs))}, {Label: "总账号数", Value: fmt.Sprintf("%d", totalUsers)}, {Label: "总校区数", Value: fmt.Sprintf("%d", sumUsedCampuses(orgs))}, {Label: "试用机构", Value: fmt.Sprintf("%d", trialCount)}}
	} else {
		resp.SummaryCards = []dto.DashboardSummaryCard{
			{Label: "平台角色", Value: platformRoleName(actor.RoleCode), Desc: "按当前岗位展示可处理事项"},
			{Label: "可见机构", Value: fmt.Sprintf("%d", len(orgs)), Desc: "当前账号可查看的机构范围"},
			{Label: "高风险机构", Value: fmt.Sprintf("%d", len(riskItems)), Desc: "与当前岗位相关的风险线索"},
			{Label: "平台账号", Value: fmt.Sprintf("%d", totalUsers), Desc: "全平台已开通账号数"},
		}
		resp.MetricItems = []dto.DashboardMetricItem{{Label: "岗位", Value: platformRoleName(actor.RoleCode)}, {Label: "可见机构", Value: fmt.Sprintf("%d", len(orgs))}, {Label: "高风险机构", Value: fmt.Sprintf("%d", len(riskItems))}, {Label: "试用机构", Value: fmt.Sprintf("%d", trialCount)}}
		if actor.RoleCode == proposal.RolePlatformAuditor {
			resp.MetricItems = append(resp.MetricItems, dto.DashboardMetricItem{Label: "主责", Value: "审计 / 恢复告警"})
		} else if actor.RoleCode == proposal.RolePlatformSupport {
			resp.MetricItems = append(resp.MetricItems, dto.DashboardMetricItem{Label: "主责", Value: "客户支持 / 恢复处理"})
		}
	}
	if actor.RoleCode == proposal.RolePlatformAuditor || actor.RoleCode == proposal.RolePlatformSupport {
		resp.PlatformRisks = riskItems
	}
	resp.QuickActions = platformQuickActions(actor)
	if len(resp.PlatformRisks) > 8 {
		resp.PlatformRisks = resp.PlatformRisks[:8]
	}
	return resp, nil
}

func (s *service) adminOverview(ctx context.Context, actor proposal.SessionUserInfo, query dto.DashboardQuery) (*dto.DashboardResponse, error) {
	students, err := s.students.List(ctx, dto.StudentListQuery{PageNum: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	payments, err := s.payments.List(ctx, dto.PaymentRecordListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	schedules, err := s.schedules.List(ctx, dto.ScheduleListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	lessonRecords, err := s.lessonRecords.List(ctx, dto.LessonRecordListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	reschedules, err := s.reschedules.List(ctx, dto.RescheduleRecordListQuery{PageNum: 1, PageSize: 8})
	if err != nil {
		return nil, err
	}
	subscriptionDays := int32(0)
	if actor.OrganizationID > 0 {
		current, err := s.subscriptionQuery.GetCurrentSubscription(ctx, actor)
		if err == nil && current != nil {
			subscriptionDays = current.RemainingDays
		}
	}
	lowLessonCount, arrearsCount, inactiveCount, err := s.loadOperationalAlerts(ctx, actor)
	if err != nil {
		return nil, err
	}
	organizationCampusCount := 0
	organizationUserCount := 0
	if actor.IsOrgAdmin() {
		campuses, err := s.tenantManagement.ListCampuses(ctx, actor)
		if err == nil {
			organizationCampusCount = len(campuses)
		}
		if actor.OrganizationID > 0 {
			var enabledUserCount int64
			if err := s.db.GetDbR().WithContext(ctx).Model(&model.SysUser{}).Where("organization_id = ? AND status = ?", actor.OrganizationID, proposal.UserStatusEnabled).Count(&enabledUserCount).Error; err == nil {
				organizationUserCount = int(enabledUserCount)
			}
		}
	}
	recentRevenue := 0.0
	for _, item := range payments {
		recentRevenue += item.Amount
	}
	resp := &dto.DashboardResponse{Role: actor.RoleCode, SubscriptionDays: subscriptionDays}
	if actor.IsOrgAdmin() {
		riskCount := lowLessonCount + arrearsCount + inactiveCount + len(reschedules)
		resp.SummaryCards = []dto.DashboardSummaryCard{
			{Label: "校区总数", Value: fmt.Sprintf("%d", organizationCampusCount), Desc: "当前机构已配置校区数量"},
			{Label: "启用账号数", Value: fmt.Sprintf("%d", organizationUserCount), Desc: "当前机构处于启用状态的账号"},
			{Label: "订阅剩余天数", Value: fmt.Sprintf("%d", subscriptionDays), Desc: "建议提前安排续费与配额规划"},
			{Label: "高风险事项数", Value: fmt.Sprintf("%d", riskCount), Desc: "需优先关注的机构级异常信号"},
		}
	} else {
		resp.SummaryCards = []dto.DashboardSummaryCard{
			{Label: "当前学员", Value: fmt.Sprintf("%d", len(students)), Desc: "当前校区管理范围内学员数"},
			{Label: "最近实收", Value: formatCurrency(recentRevenue), Desc: "最近一批收费记录合计"},
			{Label: "待执行排课", Value: fmt.Sprintf("%d", len(schedules)), Desc: "当前校区课程安排"},
			{Label: "待处理调补课", Value: fmt.Sprintf("%d", len(reschedules)), Desc: "请假、调课和补课待处理量"},
		}
	}
	resp.Alerts = []dto.DashboardAlert{
		{Label: "低课时学员", Value: fmt.Sprintf("%d", lowLessonCount), Level: chooseLevel(lowLessonCount), Path: chooseByRole(actor.IsOrgAdmin(), "/org/lesson-packages?status=active&lowLessonAlert=1", "/campus/lesson-packages?status=active&lowLessonAlert=1")},
		{Label: "欠费学员", Value: fmt.Sprintf("%d", arrearsCount), Level: chooseLevel(arrearsCount), Path: chooseByRole(actor.IsOrgAdmin(), "/org/lesson-packages?paymentStatus=arrears", "/campus/lesson-packages?paymentStatus=arrears")},
		{Label: "长期未上课", Value: fmt.Sprintf("%d", inactiveCount), Level: chooseLevel(inactiveCount), Path: "/students?inactiveAlert=1"},
		{Label: "待处理调补课", Value: fmt.Sprintf("%d", len(reschedules)), Level: chooseLevel(len(reschedules)), Path: chooseByRole(actor.IsOrgAdmin(), "/org/reschedules", "/campus/reschedules")},
	}
	if actor.IsOrgAdmin() {
		resp.CampusRisks = s.loadOrgCampusRisks(ctx, actor)
	}
	if actor.IsOrgAdmin() {
		resp.QuickActions = []dto.DashboardQuickAction{
			{Title: "校区管理", Desc: "维护机构下校区结构与启停状态", Path: "/platform/campuses"},
			{Title: "用户管理", Desc: "维护机构账号、角色与权限分工", Path: "/org/users"},
			{Title: "学员管理", Desc: "查看跨校区学员档案与运营状态", Path: "/students"},
			{Title: "排课管理", Desc: "统筹机构内校区课程安排", Path: "/org/schedules"},
		}
	} else {
		resp.QuickActions = []dto.DashboardQuickAction{
			{Title: "学员管理", Desc: "快速查看和维护本校区学员档案", Path: "/students"},
			{Title: "排课管理", Desc: "查看本校区待执行课程与排课安排", Path: "/campus/schedules"},
			{Title: "用户管理", Desc: "维护本校区教师账号和现场执行账号", Path: "/campus/users"},
			{Title: "调补课处理", Desc: "处理请假、调课和补课申请", Path: "/campus/reschedules"},
		}
	}
	resp.Payments = mapPayments(payments)
	resp.Schedules = mapSchedules(schedules)
	resp.LessonRecords = mapLessonRecords(lessonRecords)
	_ = query
	return resp, nil
}

func mapPayments(items dto.PaymentRecordListResponse) []dto.DashboardPaymentItem {
	result := make([]dto.DashboardPaymentItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.DashboardPaymentItem{ID: item.ID, StudentName: item.StudentName, LessonPackageName: item.LessonPackageName, PaymentType: item.PaymentType, Amount: item.Amount})
	}
	return result
}

func mapSchedules(items dto.ScheduleListResponse) []dto.DashboardScheduleItem {
	result := make([]dto.DashboardScheduleItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.DashboardScheduleItem{ID: item.ID, CourseName: item.CourseName, ClassDate: item.ClassDate, StartTime: formatDateTime(item.StartTime), Status: item.ScheduleStatus})
	}
	return result
}

func mapLessonRecords(items dto.LessonRecordListResponse) []dto.DashboardLessonRecordItem {
	result := make([]dto.DashboardLessonRecordItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.DashboardLessonRecordItem{ID: item.ID, LessonContent: item.LessonContent, AttendanceStatus: item.AttendanceStatus, RecordedAt: formatDateTime(item.RecordedAt)})
	}
	return result
}

func (s *service) loadOperationalAlerts(ctx context.Context, actor proposal.SessionUserInfo) (int, int, int, error) {
	readDB := s.db.GetDbR().WithContext(ctx)
	base, err := serviceutil.ApplyTenantGormScope(actor, readDB.Model(&model.LessonPackage{}), "organization_id", "campus_id")
	if err != nil {
		return 0, 0, 0, err
	}

	lowLessonCount := int64(0)
	if err := base.Where("remain_lessons <= low_lesson_threshold").Distinct("student_id").Count(&lowLessonCount).Error; err != nil {
		return 0, 0, 0, err
	}

	arrearsCount := int64(0)
	arrearsQuery, err := serviceutil.ApplyTenantGormScope(actor, readDB.Model(&model.LessonPackage{}), "organization_id", "campus_id")
	if err != nil {
		return 0, 0, 0, err
	}
	if err := arrearsQuery.Where("paid_amount < total_amount").Distinct("student_id").Count(&arrearsCount).Error; err != nil {
		return 0, 0, 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -30)
	studentQuery, err := serviceutil.ApplyTenantGormScope(actor, readDB.Model(&model.Student{}), "organization_id", "campus_id")
	if err != nil {
		return 0, 0, 0, err
	}
	var students []model.Student
	if err := studentQuery.Where("status = ?", "active").Find(&students).Error; err != nil {
		return 0, 0, 0, err
	}
	if len(students) == 0 {
		return int(lowLessonCount), int(arrearsCount), 0, nil
	}
	studentIDs := make([]int32, 0, len(students))
	for _, item := range students {
		studentIDs = append(studentIDs, item.ID)
	}
	var rows []struct {
		StudentID int32
		Latest    time.Time
	}
	lessonQuery := readDB.Model(&model.LessonRecord{}).Select("student_id, MAX(recorded_at) AS latest").Where("student_id IN ?", studentIDs).Group("student_id")
	if err := lessonQuery.Scan(&rows).Error; err != nil {
		return 0, 0, 0, err
	}
	lastMap := make(map[int32]time.Time, len(rows))
	for _, row := range rows {
		lastMap[row.StudentID] = row.Latest
	}
	inactiveCount := 0
	for _, student := range students {
		latest, ok := lastMap[student.ID]
		if !ok || latest.Before(cutoff) {
			inactiveCount++
		}
	}
	return int(lowLessonCount), int(arrearsCount), inactiveCount, nil
}

func platformQuickActions(actor proposal.SessionUserInfo) []dto.DashboardQuickAction {
	switch actor.RoleCode {
	case proposal.RolePlatformAdmin:
		return []dto.DashboardQuickAction{
			{Title: "平台运营", Desc: "查看平台运营与增长概况", Path: "/platform/ops"},
			{Title: "续费跟进", Desc: "查看即将到期与高风险机构", Path: "/platform/subscriptions"},
			{Title: "机构管理", Desc: "维护机构基础信息", Path: "/platform/tenants"},
			{Title: "平台用户", Desc: "维护平台侧账号", Path: "/platform/users"},
		}
	case proposal.RolePlatformOps:
		return []dto.DashboardQuickAction{
			{Title: "平台运营", Desc: "查看平台运营与机构状态", Path: "/platform/ops"},
			{Title: "续费跟进", Desc: "跟进试用、续费与高风险机构", Path: "/platform/subscriptions"},
			{Title: "机构管理", Desc: "查看机构与校区基础信息", Path: "/platform/tenants"},
			{Title: "校区管理", Desc: "查看校区开通与启停状态", Path: "/platform/campuses"},
		}
	case proposal.RolePlatformFinance:
		return []dto.DashboardQuickAction{
			{Title: "续费跟进", Desc: "处理套餐、续费和配额相关事项", Path: "/platform/subscriptions"},
			{Title: "机构管理", Desc: "查看机构订阅与合同关联信息", Path: "/platform/tenants"},
		}
	case proposal.RolePlatformAuditor:
		return []dto.DashboardQuickAction{
			{Title: "审计日志", Desc: "查看关键操作审计记录", Path: "/platform/audit-logs"},
			{Title: "恢复失败告警", Desc: "查看账号恢复异常与安全告警", Path: "/platform/recovery-alerts"},
		}
	case proposal.RolePlatformSupport:
		return []dto.DashboardQuickAction{
			{Title: "机构管理", Desc: "查看机构基础信息辅助支持", Path: "/platform/tenants"},
			{Title: "校区管理", Desc: "查看校区信息辅助定位问题", Path: "/platform/campuses"},
			{Title: "恢复失败告警", Desc: "处理账号恢复失败告警", Path: "/platform/recovery-alerts"},
		}
	default:
		return nil
	}
}

func platformRoleName(roleCode string) string {
	switch roleCode {
	case proposal.RolePlatformAdmin:
		return "平台超管"
	case proposal.RolePlatformOps:
		return "平台运营"
	case proposal.RolePlatformFinance:
		return "平台财务"
	case proposal.RolePlatformAuditor:
		return "平台审计"
	case proposal.RolePlatformSupport:
		return "平台支持"
	default:
		return "平台角色"
	}
}

func sumUsedCampuses(items []tenantservice.OrgCampusSummary) int64 {
	total := int64(0)
	for _, item := range items {
		total += item.UsedCampuses
	}
	return total
}

func (s *service) loadOrgCampusRisks(ctx context.Context, actor proposal.SessionUserInfo) []dto.DashboardCampusRiskItem {
	campuses, err := s.tenantManagement.ListCampuses(ctx, actor)
	if err != nil || len(campuses) == 0 {
		return nil
	}
	result := make([]dto.DashboardCampusRiskItem, 0, len(campuses))
	for _, campus := range campuses {
		var activeUsers int64
		_ = s.db.GetDbR().WithContext(ctx).Model(&model.SysUser{}).Where("organization_id = ? AND campus_id = ? AND status = ?", actor.OrganizationID, campus.ID, proposal.UserStatusEnabled).Count(&activeUsers).Error
		var lowLessonCount int64
		_ = s.db.GetDbR().WithContext(ctx).Model(&model.LessonPackage{}).Where("organization_id = ? AND campus_id = ? AND remain_lessons <= low_lesson_threshold", actor.OrganizationID, campus.ID).Distinct("student_id").Count(&lowLessonCount).Error
		var arrearsCount int64
		_ = s.db.GetDbR().WithContext(ctx).Model(&model.LessonPackage{}).Where("organization_id = ? AND campus_id = ? AND paid_amount < total_amount", actor.OrganizationID, campus.ID).Distinct("student_id").Count(&arrearsCount).Error
		var rescheduleCount int64
		_ = s.db.GetDbR().WithContext(ctx).Model(&model.RescheduleRecord{}).Where("organization_id = ? AND campus_id = ?", actor.OrganizationID, campus.ID).Count(&rescheduleCount).Error
		result = append(result, dto.DashboardCampusRiskItem{
			CampusID:         campus.ID,
			CampusName:       campus.Name,
			OrganizationID:   campus.OrganizationID,
			OrganizationName: actor.OrganizationName,
			ActiveUsers:      activeUsers,
			RiskCount:        int32(lowLessonCount + arrearsCount + rescheduleCount),
			LowLessonCount:   int32(lowLessonCount),
			ArrearsCount:     int32(arrearsCount),
			InactiveCount:    0,
			RescheduleCount:  int32(rescheduleCount),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].RiskCount == result[j].RiskCount {
			return result[i].CampusName < result[j].CampusName
		}
		return result[i].RiskCount > result[j].RiskCount
	})
	if len(result) > 5 {
		result = result[:5]
	}
	return result
}

func chooseByRole[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func chooseLevel(count int) string {
	if count >= 5 {
		return "warning"
	}
	if count > 0 {
		return "info"
	}
	return "success"
}

func formatCurrency(value float64) string { return fmt.Sprintf("¥ %.0f", value) }
func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
func sortPlatformRisks(items []dto.DashboardPlatformRiskItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].RemainingDays == items[j].RemainingDays {
			return items[i].OrganizationName < items[j].OrganizationName
		}
		return items[i].RemainingDays < items[j].RemainingDays
	})
}
