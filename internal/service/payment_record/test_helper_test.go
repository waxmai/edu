package payment_record

import (
	"context"

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
