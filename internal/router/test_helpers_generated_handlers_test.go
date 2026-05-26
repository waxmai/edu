package router

import (
	"context"

	apiCourse "edu-schedule-system/internal/api/course"
	apiLessonPackage "edu-schedule-system/internal/api/lesson_package"
	apiLessonRecord "edu-schedule-system/internal/api/lesson_record"
	apiPaymentRecord "edu-schedule-system/internal/api/payment_record"
	apiRescheduleRecord "edu-schedule-system/internal/api/reschedule_record"
	apiSchedule "edu-schedule-system/internal/api/schedule"
	apiStudent "edu-schedule-system/internal/api/student"
	courseSvc "edu-schedule-system/internal/service/course"
	"edu-schedule-system/internal/service/dto"
	lessonPackageSvc "edu-schedule-system/internal/service/lesson_package"
	lessonRecordSvc "edu-schedule-system/internal/service/lesson_record"
	paymentRecordSvc "edu-schedule-system/internal/service/payment_record"
	rescheduleRecordSvc "edu-schedule-system/internal/service/reschedule_record"
	scheduleSvc "edu-schedule-system/internal/service/schedule"
	studentSvc "edu-schedule-system/internal/service/student"
	"go.uber.org/zap"
)

func newFakeStudentHandler() *apiStudent.Handler {
	return apiStudent.New(zap.NewNop(), fakeStudentService{})
}
func newFakeCourseHandler() *apiCourse.Handler {
	return apiCourse.New(zap.NewNop(), fakeCourseService{})
}
func newFakeLessonPackageHandler() *apiLessonPackage.Handler {
	return apiLessonPackage.New(zap.NewNop(), fakeLessonPackageService{})
}
func newFakePaymentRecordHandler() *apiPaymentRecord.Handler {
	return apiPaymentRecord.New(zap.NewNop(), fakePaymentRecordService{})
}
func newFakeScheduleHandler() *apiSchedule.Handler {
	return apiSchedule.New(zap.NewNop(), fakeScheduleService{})
}
func newFakeLessonRecordHandler() *apiLessonRecord.Handler {
	return apiLessonRecord.New(zap.NewNop(), fakeLessonRecordService{})
}
func newFakeRescheduleRecordHandler() *apiRescheduleRecord.Handler {
	return apiRescheduleRecord.New(zap.NewNop(), fakeRescheduleRecordService{})
}

type fakeCourseService struct{}

func (fakeCourseService) Create(context.Context, *dto.CourseCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeCourseService) List(context.Context) (dto.CourseListResponse, error) {
	return dto.CourseListResponse{}, nil
}
func (fakeCourseService) GetByID(context.Context, int32) (*dto.CourseResponse, error) {
	return &dto.CourseResponse{ID: 1}, nil
}
func (fakeCourseService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeCourseService) UpdateByID(context.Context, int32, dto.CourseUpdateRequest) (int64, error) {
	return 1, nil
}

var _ courseSvc.Service = fakeCourseService{}

type fakeStudentService struct{}

func (fakeStudentService) Create(context.Context, *dto.StudentCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeStudentService) List(context.Context, ...dto.StudentListQuery) (dto.StudentListResponse, error) {
	return dto.StudentListResponse{}, nil
}
func (fakeStudentService) GetByID(context.Context, int32) (*dto.StudentResponse, error) {
	return &dto.StudentResponse{ID: 1}, nil
}
func (fakeStudentService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeStudentService) UpdateByID(context.Context, int32, dto.StudentUpdateRequest) (int64, error) {
	return 1, nil
}

var _ studentSvc.Service = fakeStudentService{}

type fakeLessonPackageService struct{}

func (fakeLessonPackageService) Create(context.Context, *dto.LessonPackageCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeLessonPackageService) List(context.Context, ...dto.LessonPackageListQuery) (dto.LessonPackageListResponse, error) {
	return dto.LessonPackageListResponse{}, nil
}
func (fakeLessonPackageService) GetByID(context.Context, int32) (*dto.LessonPackageResponse, error) {
	return &dto.LessonPackageResponse{ID: 1}, nil
}
func (fakeLessonPackageService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeLessonPackageService) UpdateByID(context.Context, int32, dto.LessonPackageUpdateRequest) (int64, error) {
	return 1, nil
}

var _ lessonPackageSvc.Service = fakeLessonPackageService{}

type fakePaymentRecordService struct{}

func (fakePaymentRecordService) Create(context.Context, *dto.PaymentRecordCreateRequest) (int32, error) {
	return 1, nil
}
func (fakePaymentRecordService) List(context.Context, ...dto.PaymentRecordListQuery) (dto.PaymentRecordListResponse, error) {
	return dto.PaymentRecordListResponse{}, nil
}
func (fakePaymentRecordService) GetByID(context.Context, int32) (*dto.PaymentRecordResponse, error) {
	return &dto.PaymentRecordResponse{ID: 1}, nil
}
func (fakePaymentRecordService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakePaymentRecordService) UpdateByID(context.Context, int32, dto.PaymentRecordUpdateRequest) (int64, error) {
	return 1, nil
}
func (fakePaymentRecordService) IncomeStatistics(context.Context, dto.PaymentIncomeStatsQuery) (*dto.PaymentIncomeStatsResponse, error) {
	return &dto.PaymentIncomeStatsResponse{}, nil
}

var _ paymentRecordSvc.Service = fakePaymentRecordService{}

type fakeScheduleService struct{}

func (fakeScheduleService) Create(context.Context, *dto.ScheduleCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeScheduleService) List(context.Context, ...dto.ScheduleListQuery) (dto.ScheduleListResponse, error) {
	return dto.ScheduleListResponse{}, nil
}
func (fakeScheduleService) GetByID(context.Context, int32) (*dto.ScheduleResponse, error) {
	return &dto.ScheduleResponse{ID: 1}, nil
}
func (fakeScheduleService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeScheduleService) UpdateByID(context.Context, int32, dto.ScheduleUpdateRequest) (int64, error) {
	return 1, nil
}
func (fakeScheduleService) Cancel(context.Context, int32, *dto.ScheduleCancelRequest) (int64, error) {
	return 1, nil
}
func (fakeScheduleService) Leave(context.Context, int32, *dto.ScheduleLeaveRequest) (int32, error) {
	return 1, nil
}
func (fakeScheduleService) Reschedule(context.Context, int32, *dto.ScheduleRescheduleRequest) (int32, error) {
	return 1, nil
}
func (fakeScheduleService) CreateMakeup(context.Context, *dto.MakeupScheduleCreateRequest) (int32, error) {
	return 1, nil
}

var _ scheduleSvc.Service = fakeScheduleService{}

type fakeLessonRecordService struct{}

func (fakeLessonRecordService) Create(context.Context, *dto.LessonRecordCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeLessonRecordService) List(context.Context, ...dto.LessonRecordListQuery) (dto.LessonRecordListResponse, error) {
	return dto.LessonRecordListResponse{}, nil
}
func (fakeLessonRecordService) GetByID(context.Context, int32) (*dto.LessonRecordResponse, error) {
	return &dto.LessonRecordResponse{ID: 1}, nil
}
func (fakeLessonRecordService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeLessonRecordService) UpdateByID(context.Context, int32, dto.LessonRecordUpdateRequest) (int64, error) {
	return 1, nil
}

var _ lessonRecordSvc.Service = fakeLessonRecordService{}

type fakeRescheduleRecordService struct{}

func (fakeRescheduleRecordService) Create(context.Context, *dto.RescheduleRecordCreateRequest) (int32, error) {
	return 1, nil
}
func (fakeRescheduleRecordService) List(context.Context, ...dto.RescheduleRecordListQuery) (dto.RescheduleRecordListResponse, error) {
	return dto.RescheduleRecordListResponse{}, nil
}
func (fakeRescheduleRecordService) GetByID(context.Context, int32) (*dto.RescheduleRecordResponse, error) {
	return &dto.RescheduleRecordResponse{ID: 1}, nil
}
func (fakeRescheduleRecordService) DeleteByID(context.Context, int32) (int64, error) { return 1, nil }
func (fakeRescheduleRecordService) UpdateByID(context.Context, int32, dto.RescheduleRecordUpdateRequest) (int64, error) {
	return 1, nil
}

var _ rescheduleRecordSvc.Service = fakeRescheduleRecordService{}
