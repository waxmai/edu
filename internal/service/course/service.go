package course

import (
	"context"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/service/dto"
)

var courseStatuses = map[string]struct{}{
	"enabled":  {},
	"disabled": {},
}

type Service interface {
	Create(ctx context.Context, req *dto.CourseCreateRequest) (int32, error)
	List(ctx context.Context) (dto.CourseListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.CourseResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.CourseUpdateRequest) (int64, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }
