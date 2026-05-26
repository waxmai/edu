package router

import (
	"edu-schedule-system/internal/api/admin"
	auditapi "edu-schedule-system/internal/api/audit"
	apiAuth "edu-schedule-system/internal/api/auth"
	"edu-schedule-system/internal/api/course"
	"edu-schedule-system/internal/api/lesson_package"
	"edu-schedule-system/internal/api/lesson_record"
	notificationapi "edu-schedule-system/internal/api/notification"
	"edu-schedule-system/internal/api/payment_record"
	recoveryalertapi "edu-schedule-system/internal/api/recoveryalert"
	"edu-schedule-system/internal/api/reschedule_record"
	"edu-schedule-system/internal/api/schedule"
	"edu-schedule-system/internal/api/student"
	tenantapi "edu-schedule-system/internal/api/tenant"
	apiUser "edu-schedule-system/internal/api/user"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/router/interceptor"
)

type routeRegistrarDeps struct {
	apiV1Group              core.RouterGroup
	securedGroup            core.RouterGroup
	teacherScopedGroup      core.RouterGroup
	adminOnlyGroup          core.RouterGroup
	authHandler             *apiAuth.Handler
	auditHandler            *auditapi.Handler
	recoveryAlertHandler    *recoveryalertapi.Handler
	notificationHandler     *notificationapi.Handler
	auditPlatformGroup      core.RouterGroup
	userHandler             *apiUser.Handler
	tenantHandler           *tenantapi.Handler
	adminHandler            *admin.Handler
	studentHandler          *student.Handler
	courseHandler           *course.Handler
	lessonPackageHandler    *lesson_package.Handler
	paymentRecordHandler    *payment_record.Handler
	scheduleHandler         *schedule.Handler
	lessonRecordHandler     *lesson_record.Handler
	rescheduleRecordHandler *reschedule_record.Handler
	authInterceptor         interceptor.Interceptor
}

func registerAuthAndUserRoutes(deps routeRegistrarDeps) {
	apiAuth.RegisterRoutes(deps.authHandler, deps.apiV1Group, deps.securedGroup)
	auditListGroup := deps.auditPlatformGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPlatformAuditList, proposal.PermPlatformAuditExport))
	auditExportGroup := deps.auditPlatformGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPlatformAuditExport))
	auditapi.RegisterRoutes(deps.auditHandler, auditListGroup, auditExportGroup)
	recoveryalertapi.RegisterRoutes(deps.recoveryAlertHandler, deps.auditPlatformGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPlatformRecoveryAlertList, proposal.PermPlatformRecoveryAlertResolve)))
	notificationapi.RegisterRoutes(deps.notificationHandler, deps.auditPlatformGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPlatformNotificationList)))
	userGroup := deps.adminOnlyGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermSystemUserList, proposal.PermSystemUserCreate, proposal.PermSystemUserUpdate, proposal.PermSystemUserStatus, proposal.PermSystemUserUnlock, proposal.PermSystemUserResetPass))
	apiUser.RegisterRoutes(deps.userHandler, userGroup)
	tenantapi.RegisterRoutes(deps.tenantHandler, deps.adminOnlyGroup, deps.auditPlatformGroup, deps.adminOnlyGroup, deps.authInterceptor)
	admin.RegisterGeneratedAdminRoutes(deps.adminHandler, deps.adminOnlyGroup)
}

func registerStudentRoutes(deps routeRegistrarDeps) {
	studentReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermStudentList, proposal.PermStudentGet))
	studentWriteGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermStudentCreate, proposal.PermStudentUpdate, proposal.PermStudentDelete))
	student.RegisterGeneratedStudentRoutes(deps.studentHandler, studentReadGroup)
	student.RegisterBusinessRoutes(deps.studentHandler, studentReadGroup, studentWriteGroup)
}

func registerCourseRoutes(deps routeRegistrarDeps) {
	courseReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermCourseList, proposal.PermCourseGet))
	courseWriteGroup := deps.adminOnlyGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermCourseCreate, proposal.PermCourseUpdate, proposal.PermCourseDelete))
	course.RegisterGeneratedCourseRoutes(deps.courseHandler, courseReadGroup)
	course.RegisterBusinessRoutes(deps.courseHandler, courseReadGroup, courseWriteGroup)
}

func registerLessonPackageRoutes(deps routeRegistrarDeps) {
	lessonPackageReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermLessonPackageList, proposal.PermLessonPackageGet))
	lessonPackageWriteGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermLessonPackageCreate, proposal.PermLessonPackageUpdate, proposal.PermLessonPackageDelete))
	lesson_package.RegisterGeneratedLessonPackageRoutes(deps.lessonPackageHandler, lessonPackageReadGroup)
	lesson_package.RegisterBusinessRoutes(deps.lessonPackageHandler, lessonPackageReadGroup, lessonPackageWriteGroup)
}

func registerPaymentRoutes(deps routeRegistrarDeps) {
	paymentReaderGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPaymentList, proposal.PermPaymentGet, proposal.PermPaymentStats))
	paymentWriterGroup := deps.adminOnlyGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermPaymentCreate, proposal.PermPaymentUpdate, proposal.PermPaymentDelete))
	payment_record.RegisterGeneratedPaymentRecordRoutes(deps.paymentRecordHandler, paymentWriterGroup)
	payment_record.RegisterBusinessRoutes(deps.paymentRecordHandler, paymentReaderGroup, paymentWriterGroup)
}

func registerScheduleRoutes(deps routeRegistrarDeps) {
	scheduleReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermScheduleList, proposal.PermScheduleGet))
	scheduleWriteGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermScheduleCreate, proposal.PermScheduleUpdate, proposal.PermScheduleDelete, proposal.PermScheduleCancel, proposal.PermScheduleLeave, proposal.PermScheduleReschedule, proposal.PermScheduleMakeup))
	schedule.RegisterGeneratedScheduleRoutes(deps.scheduleHandler, scheduleReadGroup)
	schedule.RegisterBusinessRoutes(deps.scheduleHandler, scheduleReadGroup, scheduleWriteGroup)
}

func registerLessonRecordRoutes(deps routeRegistrarDeps) {
	lessonRecordReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermLessonRecordList, proposal.PermLessonRecordGet))
	lessonRecordWriteGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermLessonRecordCreate, proposal.PermLessonRecordUpdate, proposal.PermLessonRecordDelete))
	lesson_record.RegisterGeneratedLessonRecordRoutes(deps.lessonRecordHandler, lessonRecordReadGroup)
	lesson_record.RegisterBusinessRoutes(deps.lessonRecordHandler, lessonRecordReadGroup, lessonRecordWriteGroup)
}

func registerRescheduleRoutes(deps routeRegistrarDeps) {
	rescheduleReadGroup := deps.teacherScopedGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermRescheduleList, proposal.PermRescheduleGet))
	rescheduleWriteGroup := deps.adminOnlyGroup.Group("", deps.authInterceptor.RequirePermissions(proposal.PermRescheduleCreate, proposal.PermRescheduleUpdate, proposal.PermRescheduleDelete))
	reschedule_record.RegisterGeneratedRescheduleRecordRoutes(deps.rescheduleRecordHandler, rescheduleReadGroup)
	reschedule_record.RegisterBusinessRoutes(deps.rescheduleRecordHandler, rescheduleReadGroup, rescheduleWriteGroup)
}
