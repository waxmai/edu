package reschedule_record

import (
	"context"
	"errors"
	"testing"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"github.com/stretchr/testify/assert"
)

func TestBuildRescheduleRecordModelValidatesFields(t *testing.T) {
	_, _, err := buildRescheduleRecordModel(&dto.RescheduleRecordCreateRequest{OldScheduleID: 1, OperationType: "reschedule"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "newScheduleId")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
}

func TestSanitizeRescheduleRecordUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.RescheduleRecordUpdateRequest{{}, {"operationType": "oops"}, {"operatorId": -1.0}, {"unknown": "x"}}
	for _, tc := range cases {
		_, err := sanitizeRescheduleRecordUpdates(tc)
		assert.Error(t, err)
	}
}

func TestValidateRescheduleRecordListQueryRejectsInvalidValues(t *testing.T) {
	cases := []dto.RescheduleRecordListQuery{
		{StudentID: -1},
		{OperationType: "oops"},
		{StartDate: "2026/05/01"},
		{StartDate: "2026-05-02", EndDate: "2026-05-01"},
	}
	for _, tc := range cases {
		err := validateRescheduleRecordListQuery(tc)
		assert.Error(t, err)
	}
}

func TestDeleteByIDRequiresPositiveID(t *testing.T) {
	svc := New(&mockRepo{})
	_, err := svc.DeleteByID(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}
