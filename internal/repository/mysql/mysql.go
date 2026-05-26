package mysql

import (
	"context"
	"fmt"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _ DBProvider = (*dbProvider)(nil)

// DBProvider owns the application's read/write MySQL handles.
//
// Services should wrap these handles with generated dao.Use(...); this type is
// infrastructure plumbing, not a domain repository implementation.
type DBProvider interface {
	GetDbR() *gorm.DB
	GetDbW() *gorm.DB
	DbRClose() error
	DbWClose() error
	Ping(ctx context.Context) error
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Repo is kept as a compatibility alias for existing service and Wire code.
type Repo = DBProvider

type dbProvider struct {
	DbR *gorm.DB
	DbW *gorm.DB
}

func New() (DBProvider, error) {
	cfg := configs.Get().MySQL
	dbr, err := dbConnect(cfg.Read.User, cfg.Read.Pass, cfg.Read.Addr, cfg.Read.Name)
	if err != nil {
		return nil, err
	}

	dbw, err := dbConnect(cfg.Write.User, cfg.Write.Pass, cfg.Write.Addr, cfg.Write.Name)
	if err != nil {
		if sqlDB, closeErr := dbr.DB(); closeErr == nil {
			_ = sqlDB.Close()
		}
		return nil, err
	}

	return &dbProvider{
		DbR: dbr,
		DbW: dbw,
	}, nil
}

func (d *dbProvider) GetDbR() *gorm.DB {
	return d.DbR
}

func (d *dbProvider) GetDbW() *gorm.DB {
	return d.DbW
}

func (d *dbProvider) DbRClose() error {
	sqlDB, err := d.DbR.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *dbProvider) DbWClose() error {
	sqlDB, err := d.DbW.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *dbProvider) Ping(ctx context.Context) error {
	if err := pingDB(ctx, d.DbR); err != nil {
		return fmt.Errorf("read db: %w", err)
	}
	if err := pingDB(ctx, d.DbW); err != nil {
		return fmt.Errorf("write db: %w", err)
	}
	return nil
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func dbConnect(user, pass, addr, dbName string) (*gorm.DB, error) {
	cfg := configs.Get()
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		user,
		pass,
		addr,
		dbName,
		true,
		"Local")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		//Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db connection failed] Database name: %s", dbName))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MySQL.Pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MySQL.Pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MySQL.Pool.ConnMaxLifetimeSeconds) * time.Second)

	err = db.Use(&TracePlugin{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
