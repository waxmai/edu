package lesson_package

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gen"
)

func (s *service) List(ctx context.Context, query ...dto.LessonPackageListQuery) (dto.LessonPackageListResponse, error) {
	listQuery := dto.LessonPackageListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validateLessonPackageListQuery(listQuery); err != nil {
		return nil, err
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	do := readDB.WithContext(ctx).LessonPackage.Where(servicectx.ScopeLessonPackages(actor, readDB, buildLessonPackageConditions(readDB, listQuery))...).Order(readDB.LessonPackage.ID.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	studentMap, courseMap, err := s.loadLessonPackageRelations(ctx, readDB, items)
	if err != nil {
		return nil, err
	}
	return toLessonPackageResponses(items, studentMap, courseMap), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.LessonPackageResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("lesson package id must be positive")
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).LessonPackage.Where(servicectx.ScopeLessonPackages(actor, readDB, []gen.Condition{readDB.LessonPackage.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	studentMap, courseMap, err := s.loadLessonPackageRelations(ctx, readDB, []*model.LessonPackage{item})
	if err != nil {
		return nil, err
	}
	return toLessonPackageResponse(item, studentMap, courseMap), nil
}

func (s *service) loadLessonPackageRelations(ctx context.Context, readDB *dao.Query, items []*model.LessonPackage) (map[int32]*model.Student, map[int32]*model.Course, error) {
	studentIDs := make([]int32, 0, len(items))
	courseIDs := make([]int32, 0, len(items))
	seenStudents := make(map[int32]struct{}, len(items))
	seenCourses := make(map[int32]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seenStudents[item.StudentID]; !ok {
			seenStudents[item.StudentID] = struct{}{}
			studentIDs = append(studentIDs, item.StudentID)
		}
		if _, ok := seenCourses[item.CourseID]; !ok {
			seenCourses[item.CourseID] = struct{}{}
			courseIDs = append(courseIDs, item.CourseID)
		}
	}

	studentMap := map[int32]*model.Student{}
	courseMap := map[int32]*model.Course{}
	if len(studentIDs) > 0 {
		students, err := readDB.WithContext(ctx).Student.Where(readDB.Student.ID.In(studentIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range students {
			studentMap[item.ID] = item
		}
	}
	if len(courseIDs) > 0 {
		courses, err := readDB.WithContext(ctx).Course.Where(readDB.Course.ID.In(courseIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range courses {
			courseMap[item.ID] = item
		}
	}
	return studentMap, courseMap, nil
}
