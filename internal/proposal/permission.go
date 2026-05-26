package proposal

const (
	DataScopeAll    = "all"
	DataScopeOrg    = "org"
	DataScopeCampus = "campus"
	DataScopeSelf   = "self"
)

const (
	FeatureAuth               = "auth"
	FeatureUserManagement     = "user_management"
	FeatureStudent            = "student"
	FeatureCourse             = "course"
	FeatureLessonPackage      = "lesson_package"
	FeaturePayment            = "payment"
	FeatureSchedule           = "schedule"
	FeatureLessonRecord       = "lesson_record"
	FeatureReschedule         = "reschedule"
	FeaturePlatformManagement = "platform_management"
	FeatureSubscriptionCenter = "subscription_center"
	FeatureRecoveryOps        = "recovery_ops"
	FeatureAuditExport        = "audit_export"
	FeatureCustomerSuccess    = "customer_success"
	FeaturePlatformReport     = "platform_report"
)

const (
	MenuDashboard          = "menu:dashboard:view"
	MenuStudent            = "menu:student:view"
	MenuLessonPackage      = "menu:lesson-package:view"
	MenuSchedule           = "menu:schedule:view"
	MenuLessonRecord       = "menu:lesson-record:view"
	MenuPayment            = "menu:payment:view"
	MenuReschedule         = "menu:reschedule:view"
	MenuCourse             = "menu:course:view"
	MenuSystemUser         = "menu:system:user:view"
	MenuPlatform           = "menu:platform:view"
	MenuPlatformUsers      = "menu:platform:users:view"
	MenuPlatformOrgs       = "menu:platform:orgs:view"
	MenuCampusManage       = "menu:campus:manage:view"
	MenuSubscriptionCenter = "menu:subscription:center:view"

	PermAuthMe                       = "auth:me:view"
	PermStudentList                  = "student:list"
	PermStudentGet                   = "student:get"
	PermStudentCreate                = "student:create"
	PermStudentUpdate                = "student:update"
	PermStudentDelete                = "student:delete"
	PermCourseList                   = "course:list"
	PermCourseGet                    = "course:get"
	PermCourseCreate                 = "course:create"
	PermCourseUpdate                 = "course:update"
	PermCourseDelete                 = "course:delete"
	PermLessonPackageList            = "lesson-package:list"
	PermLessonPackageGet             = "lesson-package:get"
	PermLessonPackageCreate          = "lesson-package:create"
	PermLessonPackageUpdate          = "lesson-package:update"
	PermLessonPackageDelete          = "lesson-package:delete"
	PermScheduleList                 = "schedule:list"
	PermScheduleGet                  = "schedule:get"
	PermScheduleCreate               = "schedule:create"
	PermScheduleUpdate               = "schedule:update"
	PermScheduleDelete               = "schedule:delete"
	PermScheduleCancel               = "schedule:cancel"
	PermScheduleLeave                = "schedule:leave"
	PermScheduleReschedule           = "schedule:reschedule"
	PermScheduleMakeup               = "schedule:makeup"
	PermLessonRecordList             = "lesson-record:list"
	PermLessonRecordGet              = "lesson-record:get"
	PermLessonRecordCreate           = "lesson-record:create"
	PermLessonRecordUpdate           = "lesson-record:update"
	PermLessonRecordDelete           = "lesson-record:delete"
	PermPaymentList                  = "payment:list"
	PermPaymentGet                   = "payment:get"
	PermPaymentCreate                = "payment:create"
	PermPaymentUpdate                = "payment:update"
	PermPaymentDelete                = "payment:delete"
	PermPaymentStats                 = "payment:stats"
	PermRescheduleList               = "reschedule:list"
	PermRescheduleGet                = "reschedule:get"
	PermRescheduleCreate             = "reschedule:create"
	PermRescheduleUpdate             = "reschedule:update"
	PermRescheduleDelete             = "reschedule:delete"
	PermSystemUserList               = "system:user:list"
	PermSystemUserCreate             = "system:user:create"
	PermSystemUserUpdate             = "system:user:update"
	PermSystemUserStatus             = "system:user:status"
	PermSystemUserUnlock             = "system:user:unlock"
	PermSystemUserResetPass          = "system:user:reset-password"
	PermPlatformTenantList           = "platform:tenant:list"
	PermPlatformTenantGet            = "platform:tenant:get"
	PermPlatformTenantCreate         = "platform:tenant:create"
	PermPlatformTenantUpdate         = "platform:tenant:update"
	PermPlatformTenantDisable        = "platform:tenant:disable"
	PermPlatformTenantRestore        = "platform:tenant:restore"
	PermPlatformOrgList              = PermPlatformTenantList
	PermPlatformOrgCreate            = PermPlatformTenantCreate
	PermPlatformOrgUpdate            = PermPlatformTenantUpdate
	PermPlatformCampusList           = "platform:campus:list"
	PermPlatformCampusCreate         = "platform:campus:create"
	PermPlatformCampusUpdate         = "platform:campus:update"
	PermPlatformCampusDisable        = "platform:campus:disable"
	PermPlatformSubscriptionList     = "platform:subscription:list"
	PermPlatformSubscriptionGet      = "platform:subscription:get"
	PermPlatformSubscriptionUpdate   = "platform:subscription:update"
	PermPlatformSubscriptionQuota    = "platform:subscription:adjust-quota"
	PermPlatformSubscriptionTrial    = "platform:subscription:extend-trial"
	PermPlatformSubscriptionSuspend  = "platform:subscription:suspend"
	PermPlatformSubscriptionResume   = "platform:subscription:resume"
	PermPlatformCSList               = "platform:cs:list"
	PermPlatformCSUpdateOwner        = "platform:cs:update-owner"
	PermPlatformCSAddFollowup        = "platform:cs:add-followup"
	PermPlatformCSUpdateStatus       = "platform:cs:update-status"
	PermPlatformOpsView              = "platform:ops:view"
	PermPlatformAuditList            = "platform:audit:list"
	PermPlatformAuditView            = PermPlatformAuditList
	PermPlatformAuditExport          = "platform:audit:export"
	PermPlatformRecoveryAlertList    = "platform:recovery-alert:list"
	PermPlatformRecoveryAlertResolve = "platform:recovery-alert:resolve"
	PermPlatformRecoveryView         = PermPlatformRecoveryAlertList
	PermPlatformRecoveryManage       = PermPlatformRecoveryAlertResolve
	PermPlatformNotificationList     = "platform:notification:list"
	PermPlatformReportView           = "platform:report:view"
	PermPlatformReportExport         = "platform:report:export"
	PermTenantSettingsView           = "tenant:settings:view"
	PermTenantSettingsUpdate         = "tenant:settings:update"
)

type AccessProfile struct {
	Permissions     []string `json:"permissions"`
	MenuPermissions []string `json:"menuPermissions"`
	DataScope       string   `json:"dataScope"`
}

func PlatformAdminPermissions() []string {
	return []string{
		PermAuthMe,
		PermStudentList, PermStudentGet,
		PermCourseList, PermCourseGet,
		PermLessonPackageList, PermLessonPackageGet,
		PermScheduleList, PermScheduleGet,
		PermLessonRecordList, PermLessonRecordGet,
		PermPaymentList, PermPaymentGet,
		PermRescheduleList, PermRescheduleGet,
		PermSystemUserList,
		PermPlatformTenantList, PermPlatformTenantGet, PermPlatformTenantCreate, PermPlatformTenantUpdate, PermPlatformTenantDisable, PermPlatformTenantRestore,
		PermPlatformCampusList, PermPlatformCampusCreate, PermPlatformCampusUpdate, PermPlatformCampusDisable,
		PermPlatformSubscriptionList, PermPlatformSubscriptionGet, PermPlatformSubscriptionUpdate, PermPlatformSubscriptionQuota, PermPlatformSubscriptionTrial, PermPlatformSubscriptionSuspend, PermPlatformSubscriptionResume,
		PermPlatformCSList, PermPlatformCSUpdateOwner, PermPlatformCSAddFollowup, PermPlatformCSUpdateStatus,
		PermPlatformOpsView,
		PermPlatformAuditList, PermPlatformAuditExport,
		PermPlatformRecoveryAlertList, PermPlatformRecoveryAlertResolve,
		PermPlatformNotificationList,
		PermPlatformReportView, PermPlatformReportExport,
	}
}

func BuildAccessProfile(roleCode string) AccessProfile {
	switch roleCode {
	case RolePlatformAdmin:
		return AccessProfile{
			Permissions:     PlatformAdminPermissions(),
			MenuPermissions: []string{MenuDashboard, MenuSystemUser, MenuPlatform, MenuPlatformUsers, MenuPlatformOrgs, MenuCampusManage},
			DataScope:       DataScopeAll,
		}
	case RolePlatformOps:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermPlatformTenantList, PermPlatformTenantGet,
				PermPlatformCampusList,
				PermPlatformSubscriptionList, PermPlatformSubscriptionGet,
				PermPlatformCSList, PermPlatformCSUpdateOwner, PermPlatformCSAddFollowup, PermPlatformCSUpdateStatus,
				PermPlatformOpsView, PermPlatformReportView, PermPlatformNotificationList,
			},
			MenuPermissions: []string{MenuDashboard, MenuPlatform, MenuPlatformOrgs, MenuCampusManage},
			DataScope:       DataScopeAll,
		}
	case RolePlatformFinance:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermPlatformTenantList, PermPlatformTenantGet,
				PermPlatformSubscriptionList, PermPlatformSubscriptionGet, PermPlatformSubscriptionUpdate, PermPlatformSubscriptionQuota, PermPlatformSubscriptionTrial, PermPlatformSubscriptionSuspend, PermPlatformSubscriptionResume,
				PermPlatformReportView, PermPlatformReportExport,
			},
			MenuPermissions: []string{MenuDashboard, MenuPlatform, MenuPlatformOrgs},
			DataScope:       DataScopeAll,
		}
	case RolePlatformAuditor:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermPlatformAuditList, PermPlatformAuditExport,
				PermPlatformRecoveryAlertList, PermPlatformNotificationList,
			},
			MenuPermissions: []string{MenuDashboard, MenuPlatform},
			DataScope:       DataScopeAll,
		}
	case RolePlatformSupport:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermStudentList, PermStudentGet,
				PermCourseList, PermCourseGet,
				PermLessonPackageList, PermLessonPackageGet,
				PermScheduleList, PermScheduleGet,
				PermLessonRecordList, PermLessonRecordGet,
				PermPaymentList, PermPaymentGet,
				PermRescheduleList, PermRescheduleGet,
				PermSystemUserList,
				PermPlatformTenantList, PermPlatformTenantGet,
				PermPlatformCampusList,
				PermPlatformRecoveryAlertList, PermPlatformRecoveryAlertResolve, PermPlatformNotificationList,
			},
			MenuPermissions: []string{MenuDashboard, MenuStudent, MenuLessonPackage, MenuSchedule, MenuLessonRecord, MenuPayment, MenuReschedule, MenuCourse, MenuSystemUser, MenuPlatform, MenuCampusManage},
			DataScope:       DataScopeAll,
		}
	case RoleOrgAdmin:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermStudentList, PermStudentGet, PermStudentCreate, PermStudentUpdate, PermStudentDelete,
				PermCourseList, PermCourseGet, PermCourseCreate, PermCourseUpdate, PermCourseDelete,
				PermLessonPackageList, PermLessonPackageGet, PermLessonPackageCreate, PermLessonPackageUpdate, PermLessonPackageDelete,
				PermScheduleList, PermScheduleGet, PermScheduleCreate, PermScheduleUpdate, PermScheduleDelete, PermScheduleCancel, PermScheduleLeave, PermScheduleReschedule, PermScheduleMakeup,
				PermLessonRecordList, PermLessonRecordGet, PermLessonRecordCreate, PermLessonRecordUpdate, PermLessonRecordDelete,
				PermPaymentList, PermPaymentGet, PermPaymentCreate, PermPaymentUpdate, PermPaymentDelete, PermPaymentStats,
				PermRescheduleList, PermRescheduleGet, PermRescheduleCreate, PermRescheduleUpdate, PermRescheduleDelete,
				PermSystemUserList, PermSystemUserCreate, PermSystemUserUpdate, PermSystemUserStatus, PermSystemUserUnlock, PermSystemUserResetPass,
				PermPlatformCampusList, PermPlatformCampusCreate, PermPlatformCampusUpdate,
				PermTenantSettingsView, PermTenantSettingsUpdate,
			},
			MenuPermissions: []string{MenuDashboard, MenuStudent, MenuLessonPackage, MenuSchedule, MenuLessonRecord, MenuPayment, MenuReschedule, MenuCourse, MenuSystemUser, MenuCampusManage, MenuSubscriptionCenter},
			DataScope:       DataScopeOrg,
		}
	case RoleCampusAdmin:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermStudentList, PermStudentGet, PermStudentCreate, PermStudentUpdate, PermStudentDelete,
				PermCourseList, PermCourseGet, PermCourseCreate, PermCourseUpdate, PermCourseDelete,
				PermLessonPackageList, PermLessonPackageGet, PermLessonPackageCreate, PermLessonPackageUpdate, PermLessonPackageDelete,
				PermScheduleList, PermScheduleGet, PermScheduleCreate, PermScheduleUpdate, PermScheduleDelete, PermScheduleCancel, PermScheduleLeave, PermScheduleReschedule, PermScheduleMakeup,
				PermLessonRecordList, PermLessonRecordGet, PermLessonRecordCreate, PermLessonRecordUpdate, PermLessonRecordDelete,
				PermPaymentList, PermPaymentGet, PermPaymentCreate, PermPaymentUpdate, PermPaymentDelete, PermPaymentStats,
				PermRescheduleList, PermRescheduleGet, PermRescheduleCreate, PermRescheduleUpdate, PermRescheduleDelete,
				PermSystemUserList, PermSystemUserCreate, PermSystemUserUpdate, PermSystemUserStatus, PermSystemUserUnlock, PermSystemUserResetPass,
				PermPlatformCampusList,
			},
			MenuPermissions: []string{MenuDashboard, MenuStudent, MenuLessonPackage, MenuSchedule, MenuLessonRecord, MenuPayment, MenuReschedule, MenuCourse, MenuSystemUser},
			DataScope:       DataScopeCampus,
		}
	case RoleTeacher:
		return AccessProfile{
			Permissions: []string{
				PermAuthMe,
				PermStudentList, PermStudentGet,
				PermCourseList, PermCourseGet,
				PermLessonPackageList, PermLessonPackageGet,
				PermScheduleList, PermScheduleGet, PermScheduleCreate, PermScheduleLeave, PermScheduleReschedule,
				PermLessonRecordList, PermLessonRecordGet, PermLessonRecordCreate, PermLessonRecordUpdate,
				PermRescheduleList, PermRescheduleGet,
			},
			MenuPermissions: []string{MenuDashboard, MenuStudent, MenuSchedule, MenuLessonRecord, MenuReschedule},
			DataScope:       DataScopeSelf,
		}
	default:
		return AccessProfile{Permissions: []string{}, MenuPermissions: []string{}, DataScope: DataScopeSelf}
	}
}
