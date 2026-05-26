package lesson_package

import (
	"context"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/service/dto"
)

var lessonPackageStatuses = map[string]struct{}{
	"pending":   {},
	"active":    {},
	"exhausted": {},
	"expired":   {},
	"closed":    {},
}

type Service interface {
	Create(ctx context.Context, req *dto.LessonPackageCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.LessonPackageListQuery) (dto.LessonPackageListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.LessonPackageResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.LessonPackageUpdateRequest) (int64, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }
