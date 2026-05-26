package student

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gen"
)

func (s *service) List(ctx context.Context, query ...dto.StudentListQuery) (dto.StudentListResponse, error) {
	listQuery := dto.StudentListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validateStudentListQuery(listQuery); err != nil {
		return nil, err
	}

	readDB := dao.Use(s.db.GetDbR())
	actor := servicectx.ActorFromContext(ctx)
	baseConds := buildStudentConditions(readDB, listQuery)
	if listQuery.InactiveAlert {
		items, err := readDB.WithContext(ctx).Student.Where(servicectx.ScopeStudents(actor, readDB, baseConds)...).Order(readDB.Student.ID.Desc()).Find()
		if err != nil {
			return nil, err
		}
		filtered, err := s.filterInactiveStudents(ctx, readDB, items)
		if err != nil {
			return nil, err
		}
		if listQuery.PageNum > 0 && listQuery.PageSize > 0 {
			start := (listQuery.PageNum - 1) * listQuery.PageSize
			if start >= len(filtered) {
				return dto.StudentListResponse{}, nil
			}
			end := start + listQuery.PageSize
			if end > len(filtered) {
				end = len(filtered)
			}
			filtered = filtered[start:end]
		}
		return toStudentResponses(filtered), nil
	}

	do := readDB.WithContext(ctx).Student.Where(servicectx.ScopeStudents(actor, readDB, baseConds)...).Order(readDB.Student.ID.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	return toStudentResponses(items), nil
}

func (s *service) filterInactiveStudents(ctx context.Context, readDB *dao.Query, items []*model.Student) ([]*model.Student, error) {
	if len(items) == 0 {
		return items, nil
	}
	studentIDs := make([]int32, 0, len(items))
	for _, item := range items {
		studentIDs = append(studentIDs, item.ID)
	}
	cutoff := time.Now().AddDate(0, 0, -30)
	var rows []struct {
		StudentID int32     `gorm:"column:student_id"`
		Latest    time.Time `gorm:"column:latest"`
	}
	if err := readDB.WithContext(ctx).LessonRecord.
		Select(readDB.LessonRecord.StudentID, readDB.LessonRecord.RecordedAt.Max().As("latest")).
		Where(readDB.LessonRecord.StudentID.In(studentIDs...)).
		Group(readDB.LessonRecord.StudentID).
		Scan(&rows); err != nil {
		return nil, err
	}
	lastMap := make(map[int32]time.Time, len(rows))
	for _, row := range rows {
		lastMap[row.StudentID] = row.Latest
	}
	filtered := make([]*model.Student, 0, len(items))
	for _, item := range items {
		latest, ok := lastMap[item.ID]
		if !ok || latest.Before(cutoff) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func buildStudentConditions(readDB *dao.Query, query dto.StudentListQuery) []gen.Condition {
	conditions := make([]gen.Condition, 0, 4)
	if name := strings.TrimSpace(query.StudentName); name != "" {
		conditions = append(conditions, readDB.Student.StudentName.Like("%"+name+"%"))
	}
	if subject := strings.TrimSpace(query.Subject); subject != "" {
		conditions = append(conditions, readDB.Student.Subject.Like("%"+subject+"%"))
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		conditions = append(conditions, readDB.Student.Status.Eq(status))
	}
	if parentPhone := strings.TrimSpace(query.ParentPhone); parentPhone != "" {
		conditions = append(conditions, readDB.Student.ParentPhone.Like("%"+parentPhone+"%"))
	}
	return conditions
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.StudentResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("student id must be positive")
	}

	readDB := dao.Use(s.db.GetDbR())
	actor := servicectx.ActorFromContext(ctx)
	item, err := readDB.WithContext(ctx).Student.Where(servicectx.ScopeStudents(actor, readDB, []gen.Condition{readDB.Student.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	return toStudentResponse(item), nil
}
