package schedule

import (
	"context"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/service/dto"
)

var scheduleStatuses = map[string]struct{}{
	"scheduled": {}, "completed": {}, "leave": {}, "rescheduled": {}, "cancelled": {}, "makeup_pending": {},
}

type Service interface {
	Create(ctx context.Context, req *dto.ScheduleCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.ScheduleListQuery) (dto.ScheduleListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.ScheduleResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.ScheduleUpdateRequest) (int64, error)
	Cancel(ctx context.Context, id int32, req *dto.ScheduleCancelRequest) (int64, error)
	Leave(ctx context.Context, id int32, req *dto.ScheduleLeaveRequest) (int32, error)
	Reschedule(ctx context.Context, id int32, req *dto.ScheduleRescheduleRequest) (int32, error)
	CreateMakeup(ctx context.Context, req *dto.MakeupScheduleCreateRequest) (int32, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }
