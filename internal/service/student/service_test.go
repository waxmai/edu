package student

import (
	"context"
	"strings"
	"testing"

	"edu-schedule-system/internal/service/dto"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockRepo struct{}

func (m *mockRepo) GetDbR() *gorm.DB           { return nil }
func (m *mockRepo) GetDbW() *gorm.DB           { return nil }
func (m *mockRepo) DbRClose() error            { return nil }
func (m *mockRepo) DbWClose() error            { return nil }
func (m *mockRepo) Ping(context.Context) error { return nil }
func (m *mockRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestValidateStudentCreateRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *dto.StudentCreateRequest
		want string
	}{
		{name: "nil request", req: nil, want: "student create request is required"},
		{name: "missing studentName", req: &dto.StudentCreateRequest{Subject: "数学", TeachingType: "one_to_one"}, want: "studentName is required"},
		{name: "missing subject", req: &dto.StudentCreateRequest{StudentName: "张三", TeachingType: "one_to_one"}, want: "subject is required"},
		{name: "missing teachingType", req: &dto.StudentCreateRequest{StudentName: "张三", Subject: "数学"}, want: "teachingType is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStudentCreateRequest(tt.req)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestSanitizeStudentUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.StudentUpdateRequest{
		{},
		{"studentName": ""},
		{"subject": 123},
		{"unknown": "x"},
	}

	for _, tc := range cases {
		_, err := sanitizeStudentUpdates(tc)
		if err == nil || strings.TrimSpace(err.Error()) == "" {
			t.Fatalf("sanitizeStudentUpdates(%v) error = %v, want validation error", tc, err)
		}
	}
}
