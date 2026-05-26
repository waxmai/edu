package payment_record

import (
	"context"
	"errors"
	"testing"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"github.com/stretchr/testify/assert"
)

func TestBuildPaymentRecordModelValidatesFields(t *testing.T) {
	_, err := buildPaymentRecordModel(&dto.PaymentRecordCreateRequest{StudentID: 1, LessonPackageID: 1, PaymentType: "signup", Amount: 10, PaymentMethod: "wechat", PaymentTime: "bad"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "paymentTime")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
}

func TestSanitizePaymentRecordUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.PaymentRecordUpdateRequest{{}, {"amount": 0.0}, {"paymentStatus": "oops"}, {"unknown": "x"}}
	for _, tc := range cases {
		_, err := sanitizePaymentRecordUpdates(tc)
		assert.Error(t, err)
	}
}

func TestValidatePaymentRecordListQueryRejectsInvalidValues(t *testing.T) {
	cases := []dto.PaymentRecordListQuery{
		{StudentID: -1},
		{LessonPackageID: -1},
		{PaymentType: "oops"},
		{PaymentStatus: "oops"},
		{StartDate: "2026/05/01"},
		{StartDate: "2026-05-02", EndDate: "2026-05-01"},
	}
	for _, tc := range cases {
		err := validatePaymentRecordListQuery(tc)
		assert.Error(t, err)
	}
}

func TestNormalizePaymentStatsQueryRejectsInvalidGroupBy(t *testing.T) {
	_, _, _, err := normalizePaymentStatsQuery(dto.PaymentIncomeStatsQuery{GroupBy: "year"})
	assert.Error(t, err)
}

func TestDeleteByIDRequiresPositiveID(t *testing.T) {
	svc := New(&mockRepo{})
	_, err := svc.DeleteByID(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}
