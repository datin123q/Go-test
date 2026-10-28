TARGET = admin-api
SRCS = main.go
BIN = bin
VERSION := $(shell git describe --tags)
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X github.com/illuminarean/energy-share-core-api/server.Version=$(VERSION) -X github.com/illuminarean/energy-share-core-api/server.Commit=$(COMMIT)"
ASSETS := $(CURDIR)/../../assets
CONFIG := $(CURDIR)/../../config

DEPLOY_ENV ?= dev
PROJECT_NAME ?= energy-shares-jp
AWS_DEPLOY_REGION ?= ap-northeast-1
AWS_ACCOUNT_ID ?= $(shell echo "540464586484")
AWS_ECR_BASE_URL ?= ${AWS_ACCOUNT_ID}.dkr.ecr.$(AWS_DEPLOY_REGION).amazonaws.com
AWS_ECR_REGISTRY ?= illuminarean/${PROJECT_NAME}-$(DEPLOY_ENV)-$(TARGET)
AWS_ECR_IMAGE_TAG ?= latest
AWS_ECS_CLUSTER_NAME ?= ${PROJECT_NAME}-$(DEPLOY_ENV)-cluster-v2
AWS_ECS_SERVICE_NAME ?= ${PROJECT_NAME}-$(DEPLOY_ENV)-$(TARGET)
AWS_ECS_TASK_FAMILY ?= ${PROJECT_NAME}-$(DEPLOY_ENV)-$(TARGET)
TASK_REVISION = $(shell aws ecs describe-task-definition --task-definition $(AWS_ECS_TASK_FAMILY) --region $(AWS_DEPLOY_REGION) | jq -r '.taskDefinition | .revision' )
TASK_DESIRED_COUNT ?= 0

DOCKERFILE_NAME ?= Dockerfile

.PHONY: start
start:
	echo "Starting the apps"

.PHONY: clean
clean:
	echo "Cleaning the apps"
	go clean -modcache
	go clean -testcache
	go clean -cache
	rm -rf ./$(BIN)

.PHONY: test
test: deps
	echo "Testing the apps"
	go test -v ./...;
	echo "Testing finished"

.PHONY: deps
deps :
	go mod download

.PHONY: build
build: $(SRCS) deps
	echo "Building the app $(TARGET)"
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o ./$(BIN)/$(TARGET)  && \
	cp -R $(CONFIG)/config.base.yml ./$(BIN)/config.base.yml && \
    sed -i "s/ENV_VALUE/$(DEPLOY_ENV)/g" ./$(BIN)/config.base.yml && \
    sed -i "s/REGION_VALUE/$(AWS_DEPLOY_REGION)/g" ./$(BIN)/config.base.yml && \
	mkdir -p ./$(BIN)/assets && \
	cp -R $(ASSETS)/key ./$(BIN)/assets && \
	cp -R $(ASSETS)/template ./$(BIN)/assets && \
	cp -R $(ASSETS)/i18n ./$(BIN)/assets && \
	cp -R $(ASSETS)/font ./$(BIN)/assets && \
	cp -R $(ASSETS)/img ./$(BIN)/assets && \
	cp -R $(ASSETS)/tabledata ./$(BIN)/assets

# macOS build 시에는 sed 명령어에 "" 옵션을 추가해야 함
.PHONY: build-mac
build-mac: $(SRCS) deps
	echo "Building the app $(TARGET) on macOS"
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o ./$(BIN)/$(TARGET)  && \
	cp -R $(CONFIG)/config.base.yml ./$(BIN)/config.base.yml && \
    sed -i "" "s/ENV_VALUE/$(DEPLOY_ENV)/g" ./$(BIN)/config.base.yml && \
    sed -i "" "s/REGION_VALUE/$(AWS_DEPLOY_REGION)/g" ./$(BIN)/config.base.yml && \
	mkdir -p ./$(BIN)/assets && \
	cp -R $(ASSETS)/key ./$(BIN)/assets && \
	cp -R $(ASSETS)/template ./$(BIN)/assets && \
	cp -R $(ASSETS)/i18n ./$(BIN)/assets && \
	cp -R $(ASSETS)/font ./$(BIN)/assets && \
	cp -R $(ASSETS)/img ./$(BIN)/assets && \
	cp -R $(ASSETS)/tabledata ./$(BIN)/assets

.PHONY: run
run:
	docker run -it $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_DEPLOY_REGION).amazonaws.com/$(AWS_ECR_REGISTRY):$(AWS_ECR_IMAGE_TAG)

.PHONY: make_docker
make_docker:
	DOCKER_BUILDKIT=1 docker build \
		-f $(DOCKERFILE_NAME) \
		--build-arg BIN=$(BIN) \
		--build-arg TARGET=$(TARGET) \
		--build-arg DEPLOY_ENV=$(DEPLOY_ENV) \
		-t $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_DEPLOY_REGION).amazonaws.com/$(AWS_ECR_REGISTRY):$(AWS_ECR_IMAGE_TAG) .

.PHONY: push_ecr
push_ecr:
	echo "ECR Docker Image Push! "
	aws ecr get-login-password --region $(AWS_DEPLOY_REGION) | docker login --username AWS --password-stdin $(AWS_ECR_BASE_URL)/$(AWS_ECR_REGISTRY)
	docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_DEPLOY_REGION).amazonaws.com/$(AWS_ECR_REGISTRY):$(AWS_ECR_IMAGE_TAG)
	aws ecs register-task-definition --cli-input-json file://task-definition-$(DEPLOY_ENV).json --region $(AWS_DEPLOY_REGION) > task_definition
	@cat task_definition 2> /dev/null

.PHONY: get_task_definition
get_task_definition:
	echo "ECS Service Update!"
	@echo "AWS_ECS_TASK_FAMILY : $(AWS_ECS_TASK_FAMILY)"
	$(eval TASK_REVISION = $(shell aws ecs describe-task-definition --task-definition $(AWS_ECS_TASK_FAMILY) --region $(AWS_DEPLOY_REGION) | jq '.taskDefinition | .revision' ))
	@echo "TASK_REVISION : $(TASK_REVISION)"
	$(eval TASK_DESIRED_COUNT = $(shell aws ecs describe-services --cluster $(AWS_ECS_CLUSTER_NAME) --services $(AWS_ECS_SERVICE_NAME) --region $(AWS_DEPLOY_REGION) | jq '.services[0] | .desiredCount' ))
	$(eval TASK_DESIRED_COUNT = $(strip $(TASK_DESIRED_COUNT)))
	@echo "TASK_DESIRED_COUNT : $(TASK_DESIRED_COUNT)"
ifeq ($(TASK_DESIRED_COUNT), 0)
	$(eval TASK_DESIRED_COUNT = 1)
endif
	@echo "TASK_DESIRED_COUNT : $(TASK_DESIRED_COUNT)"

.PHONY: update_service
update_service:
	@echo "TASK_DESIRED_COUNT : $(TASK_DESIRED_COUNT)"
	aws ecs update-service --region $(AWS_DEPLOY_REGION) --cluster $(AWS_ECS_CLUSTER_NAME) --service $(AWS_ECS_SERVICE_NAME) --task-definition $(AWS_ECS_TASK_FAMILY):$(TASK_REVISION) --desired-count $(TASK_DESIRED_COUNT) > ecs_deploy
	@cat ecs_deploy 2> /dev/null

.PHONY: wait_deploy
wait_deploy:
	@./.deployscripts/ecs-deploy-wait.sh --cluster $(AWS_ECS_CLUSTER_NAME) --service $(AWS_ECS_SERVICE_NAME) --task-definition $(AWS_ECS_TASK_FAMILY):$(TASK_REVISION) --region $(AWS_DEPLOY_REGION)

.PHONY: deploy
deploy: get_task_definition update_service

.PHONY: all
all: build make_docker push_ecr deploy

# local 은 macOS 에서 사용한다고 가정한다.
.PHONY: local
local: build-mac make_docker run
