.PHONY: all help examples build run local-run env-check prod-readiness-check frontend-build migration-lint release-backup-gate mysql-up mysql-down mysql-logs migrate bootstrap migration-smoke migration-check server-smoke smoke-e2e test vet clean lint fmt mfmt wire swagger generated-check new-migration gen-handler gen-dao gen-module ci ci-full vulncheck docker-build docker-run

ifneq (,$(wildcard .env))
include .env
export
endif

PROJECT_NAME := edu-schedule-system
MAIN_FILE := main.go
DOCKER_COMPOSE ?= docker compose
LOCAL_JWT_SECRET ?= dev-secret-local-only-change-me
JWT_SECRET ?= $(LOCAL_JWT_SECRET)
MYSQL_READ_PASS ?= app
MYSQL_WRITE_PASS ?= app
REDIS_ENABLED ?= false
CORS_ALLOWED_ORIGINS ?= http://127.0.0.1:5173,http://localhost:5173,http://127.0.0.1:5174,http://localhost:5174
LOCAL_DSN ?= app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local

all: build

help:
	@printf "edu-schedule-system development commands\n\n"
	@printf "  make build                 Build ./edu-schedule-system\n"
	@printf "  make run                   Run with -env dev and local-safe defaults\n"
	@printf "  make local-run             Run dev server with template-safe local env\n"
	@printf "  make env-check             Validate effective env/config\n"
	@printf "  make prod-readiness-check  Validate production security/readiness gates\n"
	@printf "  make frontend-build        Install/build frontend web app\n"
	@printf "  make migration-lint        Lint SQL migration filenames and risky statements\n"
	@printf "  make release-backup-gate   Verify recent successful backup before release\n"
	@printf "  make mysql-up              Start local MySQL with Docker Compose\n"
	@printf "  make mysql-down            Stop local MySQL container\n"
	@printf "  make migrate               Apply SQL migrations with -env dev\n"
	@printf "  make bootstrap BOOTSTRAP_PLATFORM_ADMIN_PASSWORD='...' BOOTSTRAP_ORG_ADMIN_PASSWORD='...' BOOTSTRAP_CAMPUS_ADMIN_PASSWORD='...' BOOTSTRAP_TEACHER_PASSWORD='...'\n"
	@printf "  make migration-smoke       Apply migrations/bootstrap twice for idempotency\n"
	@printf "  make migration-check       Alias for migration-smoke\n"
	@printf "  make server-smoke          Start built API and check health/ready\n"
	@printf "  make alert-test             Send a manual alert through ALERT_CHANNEL\n"
	@printf "  make backup                 Create MySQL backup with scripts/db_backup.sh\n"
	@printf "  make backup-healthcheck     Check backup freshness and disk space\n"
	@printf "  make restore BACKUP_FILE=... Restore backup to MYSQL_DB target\n"
	@printf "  make test                  Run go test ./... -cover\n"
	@printf "  make vet                   Run go vet ./...\n"
	@printf "  make lint                  Run golangci-lint run\n"
	@printf "  make ci                    Run local CI checks: vet, test, build\n"
	@printf "  make vulncheck             Run govulncheck ./...\n"
	@printf "  make fmt                   Run custom formatter\n"
	@printf "  make wire                  Regenerate cmd/server/wire_gen.go\n"
	@printf "  make swagger               Regenerate Swagger docs\n"
	@printf "  make generated-check       Regenerate deterministic files and fail on diff\n"
	@printf "  make new-migration TABLE=admin     # admin is demo/reference\n"
	@printf "  make gen-handler TABLE=admin       # admin is demo/reference\n"
	@printf "  make gen-dao DSN='app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local' TABLES='admin'  # demo/reference\n"
	@printf "  make gen-module TABLE=admin DSN='...' TABLES='admin' APPLY_MIGRATIONS=true  # admin is demo/reference\n"
	@printf "  make docker-build          Build Docker image\n"
	@printf "  make docker-run            Start docker-compose stack\n"
	@printf "\nRun 'make examples' for copy/paste examples.\n"

examples:
	@printf "make bootstrap BOOTSTRAP_PLATFORM_ADMIN_PASSWORD='PlatformInit123!' BOOTSTRAP_ORG_ADMIN_PASSWORD='OrgInit123!' BOOTSTRAP_CAMPUS_ADMIN_PASSWORD='CampusInit123!' BOOTSTRAP_TEACHER_PASSWORD='TeacherInit123!'\n"
	@printf "make new-migration TABLE=admin  # admin is demo/reference; replace with your table\n"
	@printf "make gen-handler TABLE=admin  # admin is demo/reference; replace with your table\n"
	@printf "make gen-dao DSN='app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local' TABLES='admin'  # admin is demo/reference\n"
	@printf "make gen-dao DSN='app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local' TABLES=''\n"
	@printf "make gen-module TABLE=admin DSN='app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local' APPLY_MIGRATIONS=true  # admin is demo/reference\n"
	@printf "make wire\n"
	@printf "make swagger\n"
	@printf "make fmt\n"

build:
	go build -o $(PROJECT_NAME) $(MAIN_FILE)

run:
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) CORS_ALLOWED_ORIGINS=$(CORS_ALLOWED_ORIGINS) go run $(MAIN_FILE) -env dev

local-run:
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) CORS_ALLOWED_ORIGINS=$(CORS_ALLOWED_ORIGINS) go run $(MAIN_FILE) -env dev

env-check:
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) go run cmd/envcheck/main.go -env dev

prod-readiness-check:
	bash ./scripts/prod_readiness_check.sh configs/pro_configs.toml

frontend-build:
	bash ./scripts/frontend_build_check.sh

migration-lint:
	bash ./scripts/migration_lint.sh

release-backup-gate:
	bash ./scripts/release_backup_gate.sh

mysql-up:
	$(DOCKER_COMPOSE) up -d mysql

mysql-down:
	$(DOCKER_COMPOSE) stop mysql

mysql-logs:
	$(DOCKER_COMPOSE) logs -f mysql

migrate:
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) go run ./cmd/migrate -env dev

bootstrap:
	@test -n "$(BOOTSTRAP_PLATFORM_ADMIN_PASSWORD)$(BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH)" || (printf "Missing BOOTSTRAP_PLATFORM_ADMIN_PASSWORD or BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH\n"; exit 1)
	@test -n "$(BOOTSTRAP_ORG_ADMIN_PASSWORD)$(BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH)" || (printf "Missing BOOTSTRAP_ORG_ADMIN_PASSWORD or BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH\n"; exit 1)
	@test -n "$(BOOTSTRAP_CAMPUS_ADMIN_PASSWORD)$(BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH)" || (printf "Missing BOOTSTRAP_CAMPUS_ADMIN_PASSWORD or BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH\n"; exit 1)
	@test -n "$(BOOTSTRAP_TEACHER_PASSWORD)$(BOOTSTRAP_TEACHER_PASSWORD_HASH)" || (printf "Missing BOOTSTRAP_TEACHER_PASSWORD or BOOTSTRAP_TEACHER_PASSWORD_HASH\n"; exit 1)
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) BOOTSTRAP_PLATFORM_ADMIN_PASSWORD='$(BOOTSTRAP_PLATFORM_ADMIN_PASSWORD)' BOOTSTRAP_ORG_ADMIN_PASSWORD='$(BOOTSTRAP_ORG_ADMIN_PASSWORD)' BOOTSTRAP_CAMPUS_ADMIN_PASSWORD='$(BOOTSTRAP_CAMPUS_ADMIN_PASSWORD)' BOOTSTRAP_TEACHER_PASSWORD='$(BOOTSTRAP_TEACHER_PASSWORD)' BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH='$(BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH)' BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH='$(BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH)' BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH='$(BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH)' BOOTSTRAP_TEACHER_PASSWORD_HASH='$(BOOTSTRAP_TEACHER_PASSWORD_HASH)' go run ./cmd/migrate -env dev -bootstrap

migration-smoke:
	@python3 -c "import bcrypt" >/dev/null 2>&1 || (printf "migration-smoke requires python3 bcrypt package for bootstrap hash generation\n"; exit 1)
	@PLATFORM_HASH=$$(python3 scripts/hash_password.py 'PlatformInit123!'); \
	ORG_HASH=$$(python3 scripts/hash_password.py 'OrgInit123!'); \
	CAMPUS_HASH=$$(python3 scripts/hash_password.py 'CampusInit123!'); \
	TEACHER_HASH=$$(python3 scripts/hash_password.py 'TeacherInit123!'); \
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=false \
	BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH="$$PLATFORM_HASH" \
	BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH="$$ORG_HASH" \
	BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH="$$CAMPUS_HASH" \
	BOOTSTRAP_TEACHER_PASSWORD_HASH="$$TEACHER_HASH" \
	go run ./cmd/migrate -env dev -bootstrap
	@PLATFORM_HASH=$$(python3 scripts/hash_password.py 'PlatformInit123!'); \
	ORG_HASH=$$(python3 scripts/hash_password.py 'OrgInit123!'); \
	CAMPUS_HASH=$$(python3 scripts/hash_password.py 'CampusInit123!'); \
	TEACHER_HASH=$$(python3 scripts/hash_password.py 'TeacherInit123!'); \
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=false \
	BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH="$$PLATFORM_HASH" \
	BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH="$$ORG_HASH" \
	BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH="$$CAMPUS_HASH" \
	BOOTSTRAP_TEACHER_PASSWORD_HASH="$$TEACHER_HASH" \
	go run ./cmd/migrate -env dev -bootstrap

migration-check: migration-smoke

server-smoke: build
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=false bash ./scripts/server_smoke.sh

smoke-e2e:
	SMOKE_PREPARE_USERS=$${SMOKE_PREPARE_USERS:-none} bash ./scripts/smoke_e2e.sh

alert-test:
	bash ./scripts/alert_test_trigger.sh $${ALERT_TEST_KIND:-manual}

backup:
	bash ./scripts/db_backup.sh

backup-healthcheck:
	bash ./scripts/db_backup_healthcheck.sh

restore:
	@test -n "$(BACKUP_FILE)" || (printf "Usage: make restore BACKUP_FILE=./output/db-backups/<file>.sql.gz[.enc]\n"; exit 1)
	bash ./scripts/db_restore.sh "$(BACKUP_FILE)"

test:
	go test ./... -cover

vet:
	go vet ./...

ci:
	go vet ./...
	go test ./... -cover
	go build ./...
	make generated-check
	make frontend-build
	make migration-lint
	make prod-readiness-check

ci-full: ci vulncheck

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

clean:
	rm -f $(PROJECT_NAME)
	rm -f coverage.out

lint:
	golangci-lint run

fmt: mfmt

mfmt:
	go run cmd/mfmt/main.go

wire:
	go run github.com/google/wire/cmd/wire ./cmd/server

swagger:
	./scripts/swagger.sh

generated-check:
	go run github.com/google/wire/cmd/wire ./cmd/server
	./scripts/swagger.sh
	git add -N .
	git diff --exit-code -- cmd/server/wire_gen.go docs/docs.go docs/swagger.json docs/swagger.yaml

new-migration:
	@test -n "$(TABLE)" || (printf "Usage: make new-migration TABLE=admin [MIGRATION_NAME=create_admin]\n"; exit 1)
	./scripts/new_migration.sh "$(TABLE)" "$(MIGRATION_NAME)"

gen-handler:
	@test -n "$(TABLE)" || (printf "Usage: make gen-handler TABLE=admin\n"; exit 1)
	go run cmd/handlergen/main.go -table "$(TABLE)"

gen-dao:
	@test -n "$(DSN)" || (printf "Usage: make gen-dao DSN='app:app@tcp(127.0.0.1:3306)/edu-schedule-system?charset=utf8mb4&parseTime=True&loc=Local' TABLES='admin'\n"; exit 1)
	go run cmd/gormgen/main.go -dsn "$(DSN)" -tables "$(TABLES)"

gen-module:
	@test -n "$(TABLE)" || (printf "Usage: make gen-module TABLE=admin DSN='optional-live-db-dsn' TABLES='admin' APPLY_MIGRATIONS=true\n"; exit 1)
	JWT_SECRET=$(JWT_SECRET) MYSQL_READ_PASS=$(MYSQL_READ_PASS) MYSQL_WRITE_PASS=$(MYSQL_WRITE_PASS) REDIS_ENABLED=$(REDIS_ENABLED) ./scripts/gen_module.sh "$(TABLE)" "$${DSN:-$(LOCAL_DSN)}" "$${TABLES:-$(TABLE)}" "$${APPLY_MIGRATIONS:-true}"

docker-build:
	docker build -t $(PROJECT_NAME) .

docker-run:
	$(DOCKER_COMPOSE) up -d
