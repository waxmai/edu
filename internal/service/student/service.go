package student

import (
	"context"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/dto"
)

type Service interface {
	Create(ctx context.Context, req *dto.StudentCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.StudentListQuery) (dto.StudentListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.StudentResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.StudentUpdateRequest) (int64, error)
}

type service struct {
	db mysql.Repo
}

func New(db mysql.Repo) Service {
	return &service{db: db}
}

func actorFromContext(ctx context.Context) proposal.SessionUserInfo {
	return servicectx.ActorFromContext(ctx)
}
