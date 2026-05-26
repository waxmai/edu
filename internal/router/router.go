package router

import (
	"net/http"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/alert"
	"edu-schedule-system/internal/api/admin"
	auditapi "edu-schedule-system/internal/api/audit"
	apiAuth "edu-schedule-system/internal/api/auth"
	courseAPI "edu-schedule-system/internal/api/course"
	"edu-schedule-system/internal/api/dashboard"
	lessonPackageAPI "edu-schedule-system/internal/api/lesson_package"
	lessonRecordAPI "edu-schedule-system/internal/api/lesson_record"
	notificationapi "edu-schedule-system/internal/api/notification"
	paymentRecordAPI "edu-schedule-system/internal/api/payment_record"
	recoveryalertapi "edu-schedule-system/internal/api/recoveryalert"
	rescheduleRecordAPI "edu-schedule-system/internal/api/reschedule_record"
	scheduleAPI "edu-schedule-system/internal/api/schedule"
	studentAPI "edu-schedule-system/internal/api/student"
	tenantapi "edu-schedule-system/internal/api/tenant"
	apiUser "edu-schedule-system/internal/api/user"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/metrics"
	pkgAudit "edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/jwtoken"
	pkgRecoveryAlert "edu-schedule-system/internal/pkg/recoveryalert"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/redis"
	"edu-schedule-system/internal/router/interceptor"
	serviceAuth "edu-schedule-system/internal/service/auth"
	dashboardservice "edu-schedule-system/internal/service/dashboard"
	lessonrecordservice "edu-schedule-system/internal/service/lesson_record"
	"edu-schedule-system/internal/service/notification"
	paymentrecordservice "edu-schedule-system/internal/service/payment_record"
	reschedulerecordservice "edu-schedule-system/internal/service/reschedule_record"
	scheduleservice "edu-schedule-system/internal/service/schedule"
	studentservice "edu-schedule-system/internal/service/student"
	tenantservice "edu-schedule-system/internal/service/tenant"
	serviceUser "edu-schedule-system/internal/service/user"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type templateTokenRequest struct {
	ID            int32  `json:"id"`
	UserName      string `json:"username"`
	NickName      string `json:"nickname"`
	ExpireSeconds int64  `json:"expire_seconds"`
}

const (
	defaultTemplateTokenTTL = 2 * time.Hour
	maxTemplateTokenTTL     = 2 * time.Hour
)

func NewHTTPMux(logger *zap.Logger, db mysql.Repo, cache redis.Repo, adminHandler *admin.Handler, studentHandler *studentAPI.Handler, courseHandler *courseAPI.Handler, lessonPackageHandler *lessonPackageAPI.Handler, paymentRecordHandler *paymentRecordAPI.Handler, scheduleHandler *scheduleAPI.Handler, lessonRecordHandler *lessonRecordAPI.Handler, rescheduleRecordHandler *rescheduleRecordAPI.Handler) (core.Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	if db == nil {
		return nil, errors.New("db required")
	}

	alertHandler := alert.NotifyHandler(logger)
	alertRules := alert.NewRuleEngine(alert.NewNotifierFromEnv(), configs.ProjectName, env.Active().Value())
	metricsMonitor := alert.NewMetricsRuleMonitor(alertRules)
	loginFailureMonitor := alert.NewLoginFailureMonitor(alertRules)
	mux, err := core.New(logger,
		core.WithEnableCors(),
		core.WithEnableSwagger(),
		core.WithEnablePProf(),
		core.WithEnablePrometheus(alert.MetricsRecordHandlerWithAlerts(metrics.RecordHandler(), metricsMonitor)),
		core.WithAlertNotify(alertHandler),
		core.WithAlertRules(alertRules),
		core.WithReadinessChecks(map[string]core.ReadinessCheck{
			"mysql": db.Ping,
			"redis": cache.Ping,
		}),
	)

	if err != nil {
		panic(err)
	}

	tokenRevoker := serviceAuth.NewTokenRevoker(cache)
	authInterceptor := interceptor.New(logger, db, interceptor.WithTokenRevoker(tokenRevoker))
	authHandler := apiAuth.New(logger, serviceAuth.New(db, serviceAuth.WithTokenRevoker(tokenRevoker), serviceAuth.WithLoginFailureMonitor(loginFailureMonitor)))
	auditHandler := auditapi.New(logger, pkgAudit.NewQueryService(), pkgAudit.NewTargetSearchService(db))
	recoveryAlertHandler := recoveryalertapi.New(logger, pkgRecoveryAlert.NewQueryService(""))
	notificationHandler := notificationapi.New(logger, notification.NewQueryService(db))
	userHandler := apiUser.New(logger, serviceUser.New(db))
	tenantManagement := tenantservice.NewManagement(db)
	tenantSubscriptionQuery := tenantservice.NewSubscriptionQuery(db)
	tenantHandler := tenantapi.New(logger, tenantManagement, tenantSubscriptionQuery)

	dashboardHandler := dashboard.New(dashboardservice.New(
		db,
		studentservice.New(db),
		paymentrecordservice.New(db),
		scheduleservice.New(db),
		lessonrecordservice.New(db),
		reschedulerecordservice.New(db),
		tenantSubscriptionQuery,
		tenantManagement,
	))

	apiV1Group := mux.Group("/api/v1")
	registerTemplateAuthRoutes(apiV1Group)

	securedGroup := apiV1Group
	businessGroup := apiV1Group
	requireAuth := authRequired(configs.Get())
	if requireAuth {
		securedGroup = apiV1Group.Group("", authInterceptor.Authenticate())
		businessGroup = securedGroup.Group("", authInterceptor.RequirePasswordChanged(), authInterceptor.RequireTenantActive())
	} else {
		businessGroup = securedGroup
	}

	teacherScopedGroup := businessGroup
	adminOnlyGroup := businessGroup
	auditPlatformGroup := businessGroup
	if requireAuth {
		teacherScopedGroup = businessGroup.Group("", authInterceptor.RequireRoles(proposal.RolePlatformAdmin, proposal.RolePlatformSupport, proposal.RoleOrgAdmin, proposal.RoleCampusAdmin, proposal.RoleTeacher), authInterceptor.RequireDataScopeSelfOrAll())
		adminOnlyGroup = businessGroup.Group("", authInterceptor.RequireAdmin())
		auditPlatformGroup = businessGroup.Group("", authInterceptor.RequirePlatformRoles())
	}

	deps := routeRegistrarDeps{
		apiV1Group:              apiV1Group,
		securedGroup:            securedGroup,
		teacherScopedGroup:      teacherScopedGroup,
		adminOnlyGroup:          adminOnlyGroup,
		auditPlatformGroup:      auditPlatformGroup,
		authHandler:             authHandler,
		auditHandler:            auditHandler,
		recoveryAlertHandler:    recoveryAlertHandler,
		notificationHandler:     notificationHandler,
		userHandler:             userHandler,
		tenantHandler:           tenantHandler,
		adminHandler:            adminHandler,
		studentHandler:          studentHandler,
		courseHandler:           courseHandler,
		lessonPackageHandler:    lessonPackageHandler,
		paymentRecordHandler:    paymentRecordHandler,
		scheduleHandler:         scheduleHandler,
		lessonRecordHandler:     lessonRecordHandler,
		rescheduleRecordHandler: rescheduleRecordHandler,
		authInterceptor:         authInterceptor,
	}

	registerAuthAndUserRoutes(deps)
	dashboard.RegisterRoutes(dashboardHandler, teacherScopedGroup)
	registerStudentRoutes(deps)
	registerCourseRoutes(deps)
	registerLessonPackageRoutes(deps)
	registerPaymentRoutes(deps)
	registerScheduleRoutes(deps)
	registerLessonRecordRoutes(deps)
	registerRescheduleRoutes(deps)

	return mux, nil
}

func registerTemplateAuthRoutes(group core.RouterGroup) {
	if !templateTokenEnabled() {
		return
	}

	group.POST("/auth/token", func(ctx core.Context) {
		var req templateTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		if req.ID <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}

		if strings.TrimSpace(req.UserName) == "" {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "username is required"))
			return
		}

		expireSeconds := req.ExpireSeconds
		if expireSeconds <= 0 {
			expireSeconds = int64(defaultTemplateTokenTTL.Seconds())
		}
		if expireSeconds > int64(maxTemplateTokenTTL.Seconds()) {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "expire_seconds must be less than or equal to 7200"))
			return
		}

		jwtCfg := configs.Get().JWT
		token, err := jwtoken.New(
			jwtCfg.Secret,
			jwtoken.WithIssuer(jwtCfg.Issuer),
			jwtoken.WithAudience(jwtCfg.Audience),
			jwtoken.WithLeeway(time.Duration(jwtCfg.LeewaySeconds)*time.Second),
		).Sign(proposal.SessionUserInfo{
			Id:       req.ID,
			UserName: req.UserName,
			NickName: req.NickName,
		}, time.Duration(expireSeconds)*time.Second)
		if err != nil {
			ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
			return
		}

		ctx.Payload(map[string]interface{}{
			"token_type": "Bearer",
			"expires_in": expireSeconds,
			"token":      token,
		})
	})
}

func templateTokenEnabled() bool {
	cfg := configs.Get().Auth
	return cfg.TemplateTokenConfigured && cfg.TemplateTokenEnabled
}

func authRequired(cfg configs.Config) bool {
	switch strings.ToLower(strings.TrimSpace(cfg.Auth.Mode)) {
	case "required":
		return true
	case "disabled":
		return false
	default:
		return env.Active().IsPro()
	}
}
