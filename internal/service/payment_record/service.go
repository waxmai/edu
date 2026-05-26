package payment_record

import (
	"context"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/service/dto"
)

var paymentTypes = map[string]struct{}{"signup": {}, "renewal": {}}
var paymentStatuses = map[string]struct{}{"paid": {}, "pending": {}, "partial": {}, "refunded": {}}

const paymentTimeLayout = "2006-01-02 15:04:05"
const paymentDateLayout = "2006-01-02"

type Service interface {
	Create(ctx context.Context, req *dto.PaymentRecordCreateRequest) (int32, error)
	List(ctx context.Context, query ...dto.PaymentRecordListQuery) (dto.PaymentRecordListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.PaymentRecordResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.PaymentRecordUpdateRequest) (int64, error)
	IncomeStatistics(ctx context.Context, query dto.PaymentIncomeStatsQuery) (*dto.PaymentIncomeStatsResponse, error)
}

type service struct{ db mysql.Repo }

func New(db mysql.Repo) Service { return &service{db: db} }
