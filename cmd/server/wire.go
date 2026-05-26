//go:build wireinject
// +build wireinject

package server

import (
	"edu-schedule-system/internal/api/admin"
	"edu-schedule-system/internal/api/course"
	"edu-schedule-system/internal/api/lesson_package"
	"edu-schedule-system/internal/api/lesson_record"
	"edu-schedule-system/internal/api/payment_record"
	"edu-schedule-system/internal/api/reschedule_record"
	"edu-schedule-system/internal/api/schedule"
	"edu-schedule-system/internal/api/student"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/redis"
	"edu-schedule-system/internal/router"
	adminService "edu-schedule-system/internal/service/admin"
	courseService "edu-schedule-system/internal/service/course"
	lessonPackageService "edu-schedule-system/internal/service/lesson_package"
	lessonRecordService "edu-schedule-system/internal/service/lesson_record"
	paymentRecordService "edu-schedule-system/internal/service/payment_record"
	rescheduleRecordService "edu-schedule-system/internal/service/reschedule_record"
	scheduleService "edu-schedule-system/internal/service/schedule"
	studentService "edu-schedule-system/internal/service/student"

	"github.com/google/wire"
)

// Injectors
func InitializeApp() (*App, func(), error) {
	wire.Build(
		// Infrastructure
		NewLogger,
		mysql.New,
		redis.NewCache, // Add Redis provider
		NewHTTPServer,
		router.NewHTTPMux,

		// Services & Handlers
		admin.New,
		adminService.New,
		student.New,
		studentService.New,
		course.New,
		courseService.New,
		lesson_package.New,
		lessonPackageService.New,
		payment_record.New,
		paymentRecordService.New,
		schedule.New,
		scheduleService.New,
		lesson_record.New,
		lessonRecordService.New,
		reschedule_record.New,
		rescheduleRecordService.New,

		// App
		NewApp,
	)
	return &App{}, nil, nil
}
