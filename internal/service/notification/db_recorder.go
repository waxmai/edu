package notification

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
)

type DBRecorder struct {
	db mysql.Repo
}

func NewDBRecorder(db mysql.Repo) Recorder {
	if db == nil {
		return NewFailureAlertingRecorder(nil, nil)
	}
	return NewFailureAlertingRecorder(&DBRecorder{db: db}, nil)
}

func (r *DBRecorder) Record(ctx context.Context, record DeliveryRecord) error {
	if r == nil || r.db == nil {
		return nil
	}
	createdAt := record.RecordedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	sentAt := createdAt
	item := &model.NotificationRecord{
		OrganizationID: record.OrganizationID,
		Channel:        truncate(record.Channel, 20),
		Scene:          truncate(record.Scene, 64),
		TemplateCode:   truncate(record.TemplateCode, 100),
		TargetMasked:   truncate(record.TargetMasked, 120),
		Provider:       truncate(record.Provider, 64),
		Status:         truncate(record.Status, 20),
		ErrorCode:      truncate(record.ErrorKind, 64),
		ErrorMessage:   truncate(sanitizeErrorMessage(record.ErrorMessage), 512),
		RequestID:      truncate(firstNonEmpty(record.ExternalID, record.IdempotencyKey), 128),
		Attempts:       1,
		SentAt:         &sentAt,
		CreatedAt:      createdAt,
	}
	return r.db.GetDbW().WithContext(ctx).Create(item).Error
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func sanitizeErrorMessage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.ReplaceAll(value, "\n", " ")
}
