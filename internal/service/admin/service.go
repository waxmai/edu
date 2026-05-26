package admin

import (
	"context"
	"strings"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
)

// Service 定义 Admin 服务接口
type Service interface {
	Create(ctx context.Context, req *dto.AdminCreateRequest) (int32, error)
	List(ctx context.Context) (dto.AdminListResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.AdminResponse, error)
	DeleteByID(ctx context.Context, id int32) (int64, error)
	UpdateByID(ctx context.Context, id int32, req dto.AdminUpdateRequest) (int64, error)
}

type service struct {
	db mysql.Repo
}

func New(db mysql.Repo) Service {
	return &service{
		db: db,
	}
}

func (s *service) Create(ctx context.Context, req *dto.AdminCreateRequest) (int32, error) {
	if req == nil {
		return 0, apperr.InvalidArgument("admin create request is required")
	}
	username := strings.TrimSpace(req.Username)
	mobile := strings.TrimSpace(req.Mobile)
	if username == "" {
		return 0, apperr.InvalidArgument("username is required")
	}
	if mobile == "" {
		return 0, apperr.InvalidArgument("mobile is required")
	}

	admin := &model.Admin{
		Username: username,
		Mobile:   mobile,
	}

	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.WithContext(ctx).Admin.Create(admin)
	if err != nil {
		return 0, err
	}
	return admin.ID, nil
}

func (s *service) List(ctx context.Context) (dto.AdminListResponse, error) {
	readDB := dao.Use(s.db.GetDbR())
	items, err := readDB.WithContext(ctx).Admin.Find()
	if err != nil {
		return nil, err
	}
	return toAdminResponses(items), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.AdminResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("admin id must be positive")
	}

	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).Admin.Where(readDB.Admin.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return toAdminResponse(item), nil
}

func toAdminResponses(items []*model.Admin) dto.AdminListResponse {
	list := make(dto.AdminListResponse, 0, len(items))
	for _, item := range items {
		if response := toAdminResponse(item); response != nil {
			list = append(list, *response)
		}
	}
	return list
}

func toAdminResponse(item *model.Admin) *dto.AdminResponse {
	if item == nil {
		return nil
	}
	return &dto.AdminResponse{
		ID:        item.ID,
		Username:  item.Username,
		Mobile:    item.Mobile,
		CreatedAt: item.CreatedAt,
	}
}

func (s *service) DeleteByID(ctx context.Context, id int32) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("admin id must be positive")
	}

	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err := writeDB.Transaction(func(tx *dao.Query) error {
		if _, err := tx.WithContext(ctx).Admin.Where(tx.Admin.ID.Eq(id)).First(); err != nil {
			return err
		}

		info, err := tx.WithContext(ctx).Admin.Where(tx.Admin.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}
		rowsAffected = info.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s *service) UpdateByID(ctx context.Context, id int32, req dto.AdminUpdateRequest) (int64, error) {
	if id <= 0 {
		return 0, apperr.InvalidArgument("admin id must be positive")
	}

	updates, err := sanitizeAdminUpdates(req)
	if err != nil {
		return 0, err
	}

	var rowsAffected int64
	writeDB := dao.Use(s.db.GetDbW())
	err = writeDB.Transaction(func(tx *dao.Query) error {
		if _, err := tx.WithContext(ctx).Admin.Where(tx.Admin.ID.Eq(id)).First(); err != nil {
			return err
		}

		info, err := tx.WithContext(ctx).Admin.Where(tx.Admin.ID.Eq(id)).Updates(updates)
		if err != nil {
			return err
		}
		rowsAffected = info.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func sanitizeAdminUpdates(req dto.AdminUpdateRequest) (map[string]interface{}, error) {
	if len(req) == 0 {
		return nil, apperr.InvalidArgument("admin update fields are required")
	}

	updates := make(map[string]interface{}, len(req))
	for field, value := range req {
		switch field {
		case "username", "mobile":
			text, ok := value.(string)
			if !ok {
				return nil, apperr.InvalidArgument(field + " must be a string")
			}
			text = strings.TrimSpace(text)
			if text == "" {
				return nil, apperr.InvalidArgument(field + " is required")
			}
			updates[field] = text
		default:
			return nil, apperr.InvalidArgument("admin update field " + field + " is not allowed")
		}
	}

	return updates, nil
}
