package payment_record

import (
	"context"

	"edu-schedule-system/internal/repository/mysql/dao"
	"edu-schedule-system/internal/repository/mysql/model"
	servicectx "edu-schedule-system/internal/service"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gen"
)

func (s *service) List(ctx context.Context, query ...dto.PaymentRecordListQuery) (dto.PaymentRecordListResponse, error) {
	listQuery := dto.PaymentRecordListQuery{}
	if len(query) > 0 {
		listQuery = query[0]
	}
	if err := validatePaymentRecordListQuery(listQuery); err != nil {
		return nil, err
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	do := readDB.WithContext(ctx).PaymentRecord.Where(servicectx.ScopePaymentRecords(actor, readDB, buildPaymentRecordConditions(readDB, listQuery))...).Order(readDB.PaymentRecord.PaymentTime.Desc())
	if listQuery.PageNum > 0 {
		do = do.Offset((listQuery.PageNum - 1) * listQuery.PageSize).Limit(listQuery.PageSize)
	}
	items, err := do.Find()
	if err != nil {
		return nil, err
	}
	studentMap, lessonPackageMap, err := s.loadPaymentRecordRelations(ctx, readDB, items)
	if err != nil {
		return nil, err
	}
	return toPaymentRecordResponses(items, studentMap, lessonPackageMap), nil
}

func (s *service) GetByID(ctx context.Context, id int32) (*dto.PaymentRecordResponse, error) {
	if id <= 0 {
		return nil, apperr.InvalidArgument("payment record id must be positive")
	}
	actor := servicectx.ActorFromContext(ctx)
	readDB := dao.Use(s.db.GetDbR())
	item, err := readDB.WithContext(ctx).PaymentRecord.Where(servicectx.ScopePaymentRecords(actor, readDB, []gen.Condition{readDB.PaymentRecord.ID.Eq(id)})...).First()
	if err != nil {
		return nil, err
	}
	studentMap, lessonPackageMap, err := s.loadPaymentRecordRelations(ctx, readDB, []*model.PaymentRecord{item})
	if err != nil {
		return nil, err
	}
	return toPaymentRecordResponse(item, studentMap, lessonPackageMap), nil
}

func (s *service) loadPaymentRecordRelations(ctx context.Context, readDB *dao.Query, items []*model.PaymentRecord) (map[int32]*model.Student, map[int32]*model.LessonPackage, error) {
	studentIDs := make([]int32, 0, len(items))
	lessonPackageIDs := make([]int32, 0, len(items))
	seenStudents := map[int32]struct{}{}
	seenLessonPackages := map[int32]struct{}{}
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := seenStudents[item.StudentID]; !ok {
			seenStudents[item.StudentID] = struct{}{}
			studentIDs = append(studentIDs, item.StudentID)
		}
		if _, ok := seenLessonPackages[item.LessonPackageID]; !ok {
			seenLessonPackages[item.LessonPackageID] = struct{}{}
			lessonPackageIDs = append(lessonPackageIDs, item.LessonPackageID)
		}
	}
	studentMap := map[int32]*model.Student{}
	lessonPackageMap := map[int32]*model.LessonPackage{}
	if len(studentIDs) > 0 {
		students, err := readDB.WithContext(ctx).Student.Where(readDB.Student.ID.In(studentIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range students {
			studentMap[item.ID] = item
		}
	}
	if len(lessonPackageIDs) > 0 {
		lessonPackages, err := readDB.WithContext(ctx).LessonPackage.Where(readDB.LessonPackage.ID.In(lessonPackageIDs...)).Find()
		if err != nil {
			return nil, nil, err
		}
		for _, item := range lessonPackages {
			lessonPackageMap[item.ID] = item
		}
	}
	return studentMap, lessonPackageMap, nil
}
