package schedule

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gen"
)

func (s *service) List(ctx context.Context, query ...dto.ScheduleListQuery) (dto.ScheduleListResponse, error) {
	listQuery := dto.ScheduleListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validateScheduleListQuery(listQuery); err != nil {
		return nil, err
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	do := readDB.WithContext(ctx).Schedule.Where(servicectx.ScopeSchedules(actor, readDB, buildScheduleConditions(readDB, listQuery))...).Order(readDB.Schedule.StartTime.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	studentMap, courseMap, teacherMap, err := s.loadScheduleRelations(ctx, readDB, items)
	if err != nil {
		return nil, err
	}
	return toScheduleResponses(items, studentMap, courseMap, teacherMap), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.ScheduleResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("schedule id must be positive")
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).Schedule.Where(servicectx.ScopeSchedules(actor, readDB, []gen.Condition{readDB.Schedule.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	studentMap, courseMap, teacherMap, err := s.loadScheduleRelations(ctx, readDB, []*model.Schedule{item})
	if err != nil {
		return nil, err
	}
	return toScheduleResponse(item, studentMap, courseMap, teacherMap), nil
}

func (s *service) loadScheduleRelations(ctx context.Context, readDB *dao.Query, items []*model.Schedule) (map[int32]*model.Student, map[int32]*model.Course, map[int32]*model.SysUser, error) {
	studentIDs := make([]int32, 0, len(items))
	courseIDs := make([]int32, 0, len(items))
	teacherIDs := make([]int32, 0, len(items))
	seenStudents := map[int32]struct{}{}
	seenCourses := map[int32]struct{}{}
	seenTeachers := map[int32]struct{}{}
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
		if _, ok := seenTeachers[item.TeacherID]; !ok {
			seenTeachers[item.TeacherID] = struct{}{}
			teacherIDs = append(teacherIDs, item.TeacherID)
		}
	}
	studentMap := map[int32]*model.Student{}
	courseMap := map[int32]*model.Course{}
	teacherMap := map[int32]*model.SysUser{}
	if len(studentIDs) > 0 {
		students, err := readDB.WithContext(ctx).Student.Where(readDB.Student.ID.In(studentIDs...)).Find()
		if err != nil {
			return nil, nil, nil, err
		}
		for _, item := range students {
			studentMap[item.ID] = item
		}
	}
	if len(courseIDs) > 0 {
		courses, err := readDB.WithContext(ctx).Course.Where(readDB.Course.ID.In(courseIDs...)).Find()
		if err != nil {
			return nil, nil, nil, err
		}
		for _, item := range courses {
			courseMap[item.ID] = item
		}
	}
	if len(teacherIDs) > 0 {
		teachers, err := readDB.WithContext(ctx).SysUser.Where(readDB.SysUser.ID.In(teacherIDs...)).Find()
		if err != nil {
			return nil, nil, nil, err
		}
		for _, item := range teachers {
			teacherMap[item.ID] = item
		}
	}
	return studentMap, courseMap, teacherMap, nil
}
