package admin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	driverMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// mockRepo 模拟 mysql.Repo 接口
type mockRepo struct {
	db *gorm.DB
}

func (m *mockRepo) GetDbR() *gorm.DB {
	return m.db
}

func (m *mockRepo) GetDbW() *gorm.DB {
	return m.db
}

func (m *mockRepo) DbRClose() error {
	return nil
}

func (m *mockRepo) DbWClose() error {
	return nil
}

func (m *mockRepo) Ping(context.Context) error {
	return nil
}

func (m *mockRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestService_Create(t *testing.T) {
	// 1. 初始化 sqlmock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// 2. 初始化 GORM
	gormDB, err := gorm.Open(driverMysql.New(driverMysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// 3. 初始化 Service
	repo := &mockRepo{db: gormDB}
	svc := New(repo)

	// 4. 定义测试用例
	req := &dto.AdminCreateRequest{
		Username: "test_user",
		Mobile:   "10000000000",
	}

	// 5. 设置 Mock 期望
	// INSERT INTO `admin` ...
	// 根据 model 定义，CreatedAt 有 default:CURRENT_TIMESTAMP，所以 Create 时如果为空可能不会出现在 SQL 中
	// 或者 GORM 可能会忽略零值。
	// 根据报错信息：arguments do not match: expected 3, but got 2 arguments
	// 说明实际执行的 SQL 确实只传递了 2 个参数 (username, mobile)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `admin`").
		WithArgs(req.Username, req.Mobile).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 6. 执行测试
	id, err := svc.Create(context.Background(), req)

	// 7. 验证结果
	assert.NoError(t, err)
	assert.Equal(t, int32(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateValidatesBusinessFields(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.Create(context.Background(), &dto.AdminCreateRequest{Username: " ", Mobile: "10000000000"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username is required")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperr.KindInvalidArgument, appErr.Kind)
}

func TestService_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(driverMysql.New(driverMysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	repo := &mockRepo{db: gormDB}
	svc := New(repo)

	// 设置 Mock 期望
	rows := sqlmock.NewRows([]string{"id", "username", "mobile", "created_at"}).
		AddRow(1, "user1", "138001", nil).
		AddRow(2, "user2", "138002", nil)

	// SELECT * FROM `admin`
	// 修改期望匹配 SELECT *
	mock.ExpectQuery("SELECT \\* FROM `admin`").
		WillReturnRows(rows)

	// 执行测试
	list, err := svc.List(context.Background())

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "user1", list[0].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByIDReturnsResponseDTO(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(driverMysql.New(driverMysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	svc := New(&mockRepo{db: gormDB})
	rows := sqlmock.NewRows([]string{"id", "username", "mobile", "created_at"}).
		AddRow(1, "user1", "138001", nil)
	mock.ExpectQuery("SELECT \\* FROM `admin` WHERE `admin`\\.`id` = \\? ORDER BY `admin`\\.`id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(rows)

	item, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, int32(1), item.ID)
	assert.Equal(t, "user1", item.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateByIDWhitelistsFields(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.UpdateByID(context.Background(), 1, dto.AdminUpdateRequest{"password": "secret"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
	var appErr *apperr.Error
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperr.KindInvalidArgument, appErr.Kind)
}

func TestService_UpdateByIDRequiresPositiveID(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.UpdateByID(context.Background(), 0, dto.AdminUpdateRequest{"username": "admin"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestService_UpdateByIDUsesTransactionAndSanitizedUpdates(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(driverMysql.New(driverMysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	repo := &mockRepo{db: gormDB}
	svc := New(repo)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `admin` WHERE `admin`\\.`id` = \\? ORDER BY `admin`\\.`id` LIMIT \\?").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "mobile", "created_at"}).AddRow(1, "old", "10000000000", nil))
	mock.ExpectExec("UPDATE `admin` SET `mobile`=\\?,`username`=\\? WHERE `admin`\\.`id` = \\?").
		WithArgs("10000000000", "new_admin", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rowsAffected, err := svc.UpdateByID(context.Background(), 1, dto.AdminUpdateRequest{
		"username": " new_admin ",
		"mobile":   "10000000000",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSanitizeAdminUpdatesRejectsInvalidValues(t *testing.T) {
	cases := []dto.AdminUpdateRequest{
		{},
		{"username": ""},
		{"mobile": 123},
	}

	for _, tc := range cases {
		_, err := sanitizeAdminUpdates(tc)
		if err == nil || strings.TrimSpace(err.Error()) == "" {
			t.Fatalf("sanitizeAdminUpdates(%v) error = %v, want validation error", tc, err)
		}
	}
}
