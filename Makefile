admin-api = admin-api
user-api = user-api
jobs-scheduler = jobs-scheduler
invest-worker = invest-worker
domain = domain
service = service

apps_names = $(admin-api) $(user-api) $(jobs-scheduler) $(domain) $(service)

admin_api_path = ./apps/$(admin-api)
user_api_path = ./apps/$(user-api)
jobs_scheduler_path = ./apps/$(jobs-scheduler)
invest-worker_path = ./apps/$(invest-worker)
domain_path = ./$(domain)
service_path = ./$(service)
test_path = ./test
pkg_path = ./pkg

apps_path = $(admin_api_path) $(user_api_path) $(jobs_scheduler_path)
packages = $(domain_path) $(service_path) $(admin_api_path) $(user_api_path) $(jobs_scheduler_path) $(test_path)

.PHONY: start
start:
	echo "Starting the apps"

.PHONY: clean
clean:
	echo "Cleaning the apps"
	go clean -modcache
	go clean -testcache
	go clean -cache

.PHONY: user-api
user-api: ${user_api_path}
	$(MAKE) -C ${user_api_path} ${CMD} ${PARAMS}

.PHONY: admin-api
admin-api: ${admin_api_path}
	$(MAKE) -C ${admin_api_path} ${CMD}

.PHONY: jobs-scheduler
jobs-scheduler: ${jobs_scheduler_path}
	$(MAKE) -C ${jobs_scheduler_path} ${CMD}

# 코드 생성
.PHONY: codegen
codegen:
	@echo "Generating code from yaml definitions..."
	@go run cmd/*.go codegen


.PHONY: test
test:
	echo "Testing the apps"
	for app in $(packages); do \
  		app_name=$$(basename $$app); \
		echo "Testing $$app_name"; \
		go test -v $$app/... -coverprofile bin/"$$app_name".out; \
        if [ -f bin/"$$app_name".out ]; then \
          echo "Appending coverage.out"; \
          cat bin/"$$app_name".out >> bin/coverage.out; \
          rm bin/"$$app_name".out; \
		fi \
	done
	echo "Testing finished"

.PHONY: workspace
work:
	@echo "Creating workspace"
	@go work init && \
	go work use . && \
	go work use $(admin_api_path) && \
	go work use $(user_api_path) && \
	go work use $(jobs_scheduler_path) && \
	go work use $(invest-worker_path) && \
	go work sync && \
	echo "Workspace created"

.PHONY: sync
work-sync:
	@echo "Synchronizing go work dependencies"
	@go work sync
	@echo "Dependencies synchronized"

.PHONY: app-test
app-test:
	echo "Testing the app"
	go test -v $(admin_api_path)/...; \
	echo "Testing finished"

.PHONY: go-gen
go-gen:
	echo "Generating the mocks"
	go generate ./domain/...
	echo "Mocks generated"

# generate swagger
COMMIT_HASH := $(shell git rev-parse --short=6 HEAD)
COMMIT_DATE := $(shell git log -1 --format=%cd --date=short)
VERSION_INFO := $(COMMIT_DATE).$(COMMIT_HASH)
GIT_ROOT := $(shell git rev-parse --show-toplevel)
export VERSION_INFO
export GIT_ROOT

.PHONY: gen-swagger-admin-api
gen-swagger-admin-api:
	@echo "Generating ./swagger/swagger-admin-api.json..."
	@SWAGGER_MODE=GEN_SWAGGER_MODE go run ./apps/admin-api/main.go swagger
	@echo "--"

.PHONY: gen-swagger-user-api
gen-swagger-user-api:
	@echo "Generating ./swagger/swagger-user-api.json..."
	@SWAGGER_MODE=GEN_SWAGGER_MODE go run ./apps/user-api/main.go swagger
	@echo "--"

.PHONY: gen-swagger-job-scheduler
gen-swagger-job-scheduler:
	@echo "Generating ./swagger/swagger-job-scheduler-api.json..."
	@SWAGGER_MODE=GEN_SWAGGER_MODE go run ./apps/jobs-scheduler/main.go swagger
	@echo "--"

.PHONY: gen-swagger
gen-swagger: gen-swagger-admin-api gen-swagger-user-api gen-swagger-job-scheduler

## 로컬에서 MySQL 컨테이너를 실행하기 위한 Targets
# MySQL 컨테이너 이름 정의
CONTAINER_NAME=mysql8.0.32
SAMPLE_DATA_DIR=./db/sample_data

# 데이터베이스 덤프를 생성하는 타겟
mysql_dump:
	@echo "데이터베이스 덤프 생성 중..."
	@./mysql_dump.sh

# 로컬에서 MySQL 컨테이너를 실행하는 타겟
mysql_run:
	@echo "MySQL 컨테이너 실행 작업 시작..."
	@./mysql_run.sh

# 실행 중인 MySQL 컨테이너를 중지하고 삭제하는 타겟
mysql_remove:
	@echo "MySQL 컨테이너 중지 및 삭제 중..."
	@docker stop $(CONTAINER_NAME) || true
	@docker rm $(CONTAINER_NAME) || true

mock:
	@mockery # common
	@cd $(admin_api_path) && mockery
	@cd $(user_api_path) && mockery

.PHONY: db-setup
db-setup:
	@echo "Seeding energysharesdb schema..."
	@docker compose exec -T energysharesdb mysql -uroot -ppassword energysharesdb < db/sql/version/energyshares.base.schema.sql
	@echo "Seeding mynumberdb schema..."
	@docker compose exec -T mynumberdb sh -c 'mysql -u root -ppassword mynumberdb' < db/sql/version/mynumberdb_dump.sql
	@echo "Running admin-api migrations..."
	@cd $(admin_api_path) && go run main.go migrate --env local

.PHONY: start-admin-api
start-admin-api:
	@cd $(admin_api_path) && go run main.go run --env local

.PHONY: start-user-api
start-user-api:
	@cd $(user_api_path) && go run main.go run --env local

.PHONY: start-jobs-scheduler
start-jobs-scheduler:
	@cd $(jobs_scheduler_path) && go run main.go run --env local

.PHONY: start-invest-worker
start-invest-worker:
	@cd $(invest-worker_path) && go run main.go

.PHONY: seed-sample-data
seed-sample-data:
	@echo "Applying sample data (*.sql) under $(SAMPLE_DATA_DIR) to energysharesdb..."
	@if [ -z "$$(docker compose ps -q energysharesdb)" ]; then \
		echo "energysharesdb is not running. Start it via docker compose or make mysql_run first."; \
		exit 1; \
	fi
	@files=$$(ls $(SAMPLE_DATA_DIR)/*.sql 2>/dev/null); \
	if [ -z "$$files" ]; then \
		echo "No sample data files found in $(SAMPLE_DATA_DIR)."; \
		exit 0; \
	fi; \
	for file in $$files; do \
		echo "Applying $$file..."; \
		docker compose exec -T energysharesdb mysql -uroot -ppassword energysharesdb < $$file; \
	done

.PHONY: mysql-client
mysql-client:
	@echo "Connecting to energysharesdb MySQL container..."
	@docker compose exec energysharesdb mysql -uroot -ppassword energysharesdb;

.PHONY: docker-clean
docker-clean:
	@echo "Stopping docker compose services and removing volumes..."
	@docker compose down -v

.PHONY: db-migrate
db-migrate:
	@echo "Running admin-api migrations..."
	@cd $(admin_api_path) && go run main.go migrate --env local
