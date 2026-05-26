package course

import (
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/dto"
)

func toCourseResponses(items []*model.Course) dto.CourseListResponse {
	list := make(dto.CourseListResponse, 0, len(items))
	for _, item := range items {
		if resp := toCourseResponse(item); resp != nil {
			list = append(list, *resp)
		}
	}
	return list
}

func toCourseResponse(item *model.Course) *dto.CourseResponse {
	if item == nil {
		return nil
	}
	return &dto.CourseResponse{
		ID:              item.ID,
		CourseName:      item.CourseName,
		Subject:         item.Subject,
		CourseType:      item.CourseType,
		DurationMinutes: item.DurationMinutes,
		FeeStandard:     item.FeeStandard,
		Remark:          item.Remark,
		Status:          item.Status,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}
