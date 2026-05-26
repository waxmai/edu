package course

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gen"
)

func (s *service) List(ctx context.Context) (dto.CourseListResponse, error) {
	readDB := dao.Use(s.db.GetDbR())
	actor := servicectx.ActorFromContext(ctx)
	items, err := readDB.WithContext(ctx).Course.Where(servicectx.ScopeCourses(actor, readDB, nil)...).Order(readDB.Course.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toCourseResponses(items), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.CourseResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("course id must be positive")
	}
	readDB := dao.Use(s.db.GetDbR())
	actor := servicectx.ActorFromContext(ctx)
	item, err := readDB.WithContext(ctx).Course.Where(servicectx.ScopeCourses(actor, readDB, []gen.Condition{readDB.Course.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	return toCourseResponse(item), nil
}
