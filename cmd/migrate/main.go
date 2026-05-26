package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/env"

	_ "github.com/go-sql-driver/mysql"
)

var bootstrap = flag.Bool("bootstrap", false, "apply bootstrap sql after migrations")

type migrationFile struct {
	Version        string
	Path           string
	Name           string
	RollbackPolicy string
	Checksum       string
	SQL            string
}

func main() {
	env.Init()
	if err := configs.Init(); err != nil {
		log.Fatalf("config init err: %v", err)
	}

	cfg := configs.Get()
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		cfg.MySQL.Write.User,
		cfg.MySQL.Write.Pass,
		cfg.MySQL.Write.Addr,
		cfg.MySQL.Write.Name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql err: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql err: %v", err)
	}

	if err := ensureSchemaMigrations(ctx, db); err != nil {
		log.Fatalf("ensure schema_migrations err: %v", err)
	}

	files, err := loadMigrationFiles("migrations", false)
	if err != nil {
		log.Fatalf("load migrations err: %v", err)
	}
	if err := applyMigrations(ctx, db, files); err != nil {
		log.Fatalf("apply migrations err: %v", err)
	}

	if *bootstrap {
		if err := validateBootstrapSecrets(); err != nil {
			log.Fatalf("validate bootstrap secrets err: %v", err)
		}
		bootstrapFiles, err := loadMigrationFiles(filepath.Join("migrations", "bootstrap"), true)
		if err != nil {
			log.Fatalf("load bootstrap err: %v", err)
		}
		if err := applyRawFiles(ctx, db, bootstrapFiles); err != nil {
			log.Fatalf("apply bootstrap err: %v", err)
		}
	}

	log.Printf("migration completed successfully")
}

func ensureSchemaMigrations(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) NOT NULL PRIMARY KEY,
  checksum VARCHAR(64) NOT NULL,
  rollback_policy VARCHAR(64) NOT NULL,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='schema migrations';`)
	return err
}

func loadMigrationFiles(root string, optional bool) ([]migrationFile, error) {
	entries := make([]migrationFile, 0)
	info, err := os.Stat(root)
	if err != nil {
		if optional && os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if !optional && path != root && filepath.Base(path) == "bootstrap" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		version := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		rollbackPolicy := parseRollbackPolicy(string(content))
		if !optional && rollbackPolicy == "" {
			return fmt.Errorf("migration %s missing rollback-policy header", path)
		}
		sum := sha256.Sum256(content)
		entries = append(entries, migrationFile{
			Version:        version,
			Path:           path,
			Name:           filepath.Base(path),
			RollbackPolicy: rollbackPolicy,
			Checksum:       hex.EncodeToString(sum[:]),
			SQL:            string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version < entries[j].Version
	})
	return entries, nil
}

func parseRollbackPolicy(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- rollback-policy:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "-- rollback-policy:"))
		}
	}
	return ""
}

func applyMigrations(ctx context.Context, db *sql.DB, files []migrationFile) error {
	for _, file := range files {
		var checksum string
		var rollbackPolicy string
		err := db.QueryRowContext(ctx, "SELECT checksum, rollback_policy FROM schema_migrations WHERE version = ?", file.Version).Scan(&checksum, &rollbackPolicy)
		switch {
		case err == sql.ErrNoRows:
			if _, err := db.ExecContext(ctx, file.SQL); err != nil {
				return fmt.Errorf("execute migration %s: %w", file.Name, err)
			}
			if _, err := db.ExecContext(ctx, "INSERT INTO schema_migrations(version, checksum, rollback_policy) VALUES (?, ?, ?)", file.Version, file.Checksum, file.RollbackPolicy); err != nil {
				return fmt.Errorf("record migration %s: %w", file.Name, err)
			}
			log.Printf("applied migration %s", file.Name)
		case err != nil:
			return fmt.Errorf("check migration %s: %w", file.Name, err)
		default:
			if checksum != file.Checksum {
				return fmt.Errorf("migration checksum mismatch for %s", file.Name)
			}
			if rollbackPolicy != file.RollbackPolicy {
				return fmt.Errorf("migration rollback policy mismatch for %s", file.Name)
			}
			log.Printf("migration already applied %s", file.Name)
		}
	}
	return nil
}

func applyRawFiles(ctx context.Context, db *sql.DB, files []migrationFile) error {
	for _, file := range files {
		if strings.TrimSpace(file.SQL) == "" {
			continue
		}
		sqlText := file.SQL
		if strings.Contains(file.Path, string(filepath.Separator)+"bootstrap"+string(filepath.Separator)) {
			sqlText = applyBootstrapReplacements(sqlText)
		}
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("execute bootstrap %s: %w", file.Name, err)
		}
		log.Printf("applied bootstrap %s", file.Name)
	}
	return nil
}
