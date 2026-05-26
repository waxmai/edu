package lesson_package

import (
	"context"
	"errors"
	"testing"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"github.com/stretchr/testify/assert"
)

func TestBuildLessonPackageModelValidatesBusinessFields(t *testing.T) {
	_, err := buildLessonPackageModel(&dto.LessonPackageCreateRequest{StudentID: 1, CourseID: 1, TotalLessons: 10, UsedLessons: 11, RemainLessons: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usedLessons cannot exceed totalLessons")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
}

func TestSanitizeLessonPackageUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.LessonPackageUpdateRequest{{}, {"status": "oops"}, {"totalLessons": -1.0}, {"unknown": "x"}}
	for _, tc := range cases {
		_, err := sanitizeLessonPackageUpdates(tc)
		assert.Error(t, err)
	}
}

func TestDeleteByIDRequiresPositiveID(t *testing.T) {
	svc := New(&mockRepo{})
	_, err := svc.DeleteByID(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}
