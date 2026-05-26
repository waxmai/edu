package notification

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/serviceutil"
)

type Query struct {
	OrganizationID int32  `form:"organizationId"`
	Channel        string `form:"channel"`
	Scene          string `form:"scene"`
	Status         string `form:"status"`
	ErrorCode      string `form:"errorCode"`
	RequestID      string `form:"requestId"`
	FailuresOnly   bool   `form:"failuresOnly"`
	Offset         int    `form:"offset"`
	Limit          int    `form:"limit"`
}

type Record struct {
	ID             uint64     `json:"id"`
	OrganizationID int32      `json:"organizationId"`
	Channel        string     `json:"channel"`
	Scene          string     `json:"scene"`
	TemplateCode   string     `json:"templateCode"`
	TargetMasked   string     `json:"targetMasked"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	ErrorCode      string     `json:"errorCode"`
	ErrorMessage   string     `json:"errorMessage"`
	RequestID      string     `json:"requestId"`
	Attempts       int32      `json:"attempts"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type ListResponse struct {
	Items  []Record `json:"items"`
	Total  int64    `json:"total"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
}

type QueryService interface {
	List(ctx context.Context, query Query) (*ListResponse, error)
}

type queryService struct {
	db mysql.Repo
}

func NewQueryService(db mysql.Repo) QueryService {
	if db == nil || db.GetDbR() == nil {
		return NoopQueryService{}
	}
	return &queryService{db: db}
}

type NoopQueryService struct{}

func (NoopQueryService) List(ctx context.Context, query Query) (*ListResponse, error) {
	limit := normalizeLimit(query.Limit)
	offset := normalizeOffset(query.Offset)
	return &ListResponse{Items: []Record{}, Total: 0, Offset: offset, Limit: limit}, nil
}

func (s *queryService) List(ctx context.Context, query Query) (*ListResponse, error) {
	limit := normalizeLimit(query.Limit)
	offset := normalizeOffset(query.Offset)
	actor := servicectx.ActorFromContext(ctx)
	dbq := s.db.GetDbR().WithContext(ctx).Model(&model.NotificationRecord{})
	if !actor.IsPlatformAdmin() {
		if actor.OrganizationID <= 0 {
			return nil, apperr.Forbidden("actor organization scope is missing")
		}
		dbq = dbq.Where("organization_id = ?", actor.OrganizationID)
	}
	if query.OrganizationID > 0 {
		if err := serviceutil.EnsureSameOrganization(actor, query.OrganizationID); err != nil {
			return nil, err
		}
		dbq = dbq.Where("organization_id = ?", query.OrganizationID)
	}
	if strings.TrimSpace(query.Channel) != "" {
		dbq = dbq.Where("channel = ?", strings.TrimSpace(query.Channel))
	}
	if strings.TrimSpace(query.Scene) != "" {
		dbq = dbq.Where("scene = ?", strings.TrimSpace(query.Scene))
	}
	if query.FailuresOnly {
		dbq = dbq.Where("status = ?", "error")
	} else if strings.TrimSpace(query.Status) != "" {
		dbq = dbq.Where("status = ?", strings.TrimSpace(query.Status))
	}
	if strings.TrimSpace(query.ErrorCode) != "" {
		dbq = dbq.Where("error_code = ?", strings.TrimSpace(query.ErrorCode))
	}
	if strings.TrimSpace(query.RequestID) != "" {
		dbq = dbq.Where("request_id = ?", strings.TrimSpace(query.RequestID))
	}
	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []model.NotificationRecord
	if err := dbq.Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]Record, 0, len(rows))
	for _, row := range rows {
		items = append(items, Record{ID: row.ID, OrganizationID: row.OrganizationID, Channel: row.Channel, Scene: row.Scene, TemplateCode: row.TemplateCode, TargetMasked: row.TargetMasked, Provider: row.Provider, Status: row.Status, ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage, RequestID: row.RequestID, Attempts: row.Attempts, SentAt: row.SentAt, CreatedAt: row.CreatedAt})
	}
	return &ListResponse{Items: items, Total: total, Offset: offset, Limit: limit}, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
