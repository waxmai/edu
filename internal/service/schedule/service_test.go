package schedule

import (
	"context"
	"errors"
	"testing"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"github.com/stretchr/testify/assert"
)

func TestBuildScheduleModelValidatesFields(t *testing.T) {
	_, err := buildScheduleModel(&dto.ScheduleCreateRequest{StudentID: 1, CourseID: 1, TeacherID: 1, ClassDate: "2026-05-08", StartTime: "2026-05-08 16:00:00", EndTime: "2026-05-08 15:00:00"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startTime must be earlier than endTime")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
}

func TestSanitizeScheduleUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.ScheduleUpdateRequest{{}, {"scheduleStatus": "oops"}, {"isMakeup": "x"}, {"unknown": "x"}}
	for _, tc := range cases {
		_, err := sanitizeScheduleUpdates(tc)
		assert.Error(t, err)
	}
}

func TestAppendReasonAndReqReason(t *testing.T) {
	assert.Equal(t, "leave: test", appendReason("leave", reqReason(" test ")))
	assert.Equal(t, "cancel", appendReason("cancel", reqReason(" ")))
}

func TestDeleteByIDRequiresPositiveID(t *testing.T) {
	svc := New(&mockRepo{})
	_, err := svc.DeleteByID(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}
