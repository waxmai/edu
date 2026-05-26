package student

import (
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/dto"
)

func toStudentResponses(items []*model.Student) dto.StudentListResponse {
	list := make(dto.StudentListResponse, 0, len(items))
	for _, item := range items {
		if response := toStudentResponse(item); response != nil {
			list = append(list, *response)
		}
	}
	return list
}

func toStudentResponse(item *model.Student) *dto.StudentResponse {
	if item == nil {
		return nil
	}
	return &dto.StudentResponse{
		ID:           item.ID,
		StudentName:  item.StudentName,
		Gender:       item.Gender,
		Grade:        item.Grade,
		Phone:        item.Phone,
		ParentName:   item.ParentName,
		ParentPhone:  item.ParentPhone,
		Subject:      item.Subject,
		TeachingType: item.TeachingType,
		Status:       item.Status,
		Remark:       item.Remark,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
