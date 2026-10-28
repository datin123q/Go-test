TARGET = jobs-scheduler
SRCS = main.go
BIN = bin
VERSION := $(shell git describe --tags)
COMMIT := $(shell git rev-parse --short HEAD)
GIT_HASH := $(shell git rev-parse HEAD)
LDFLAGS :=
ASSETS := $(CURDIR)/../../assets
CONFIG := $(CURDIR)/../../config

DEP_FILE_PATH = src/main/rersources/index.html

jobs-scheduler = jobs-scheduler
domain = domain

apps_names = $(jobs-scheduler) $(domain)

jobs_scheduler_path = ./apps/$(TARGET)
domain_path = ../../$(domain)

packages = $(jobs_scheduler_path) $(domain_path)

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
	cp -R $(ASSETS)/i18n ./$(BIN)/assets && \
	cp -R $(ASSETS)/font ./$(BIN)/assets

ifeq ($(DEPLOY_ENV),stg)
	cp ./conf.$(DEPLOY_ENV).json ./$(BIN)/conf.json
else ifeq ($(DEPLOY_ENV),prod)
	cp ./conf.$(DEPLOY_ENV).json ./$(BIN)/conf.json
else ifeq ($(DEPLOY_ENV),dev)
	cp ./conf.$(DEPLOY_ENV).json ./$(BIN)/conf.json
endif

.PHONY: config_copy
config_copy:
	echo "Config Copy"
	git rev-parse --short HEAD > version
	find .deployscripts/promtail ! -name "promtail-config_$(DEPLOY_ENV).yaml" -type f -print -exec rm {} \;
	find .deployscripts/service ! -name "energyshares-jobs-scheduler-service-$(DEPLOY_ENV)" -type f -print -exec rm {} \;
	mv .deployscripts/promtail/promtail-config_$(DEPLOY_ENV).yaml .deployscripts/promtail/promtail-config.yaml
	mv .deployscripts/service/energyshares-jobs-scheduler-service-$(DEPLOY_ENV) .deployscripts/service/energyshares-jobs-scheduler-service

	mkdir -p ./build/.deployscripts
	cp -R ./.deployscripts/syslog ./build/.deployscripts
	cp -R ./.deployscripts/logrotate ./build/.deployscripts
	cp -R ./.deployscripts/systemctl ./build/.deployscripts
	cp -R ./.deployscripts/codedeploy ./build/.deployscripts
	cp -R ./.deployscripts/promtail ./build/.deployscripts
	cp -R ./.deployscripts/service ./build/.deployscripts

# .PHONY: make_zip_file
# make_zip_file:
# 	echo "Make Zip File!! "
# 	zip -r ./$(GIT_HASH)_JENKINS.zip ./build ./appspec.yml
#
# .PHONY: artifacts_to_s3
# artifacts_to_s3:
# 	echo "artifacts to s3"
# 	aws s3 cp --region $(AWS_DEPLOY_REGION) ./$(GIT_HASH)_JENKINS.zip s3://$(BUCKET_NAME)/$(BUCKET_TYPE)/$(PROJECT_NAME)/$(GIT_HASH)_JENKINS.zip
#
# .PHONY: code_deploy
# code_deploy:
# 	echo "Code Deploy"
# 	aws deploy create-deployment --application-name $(APP_NAME) --deployment-config-name CodeDeployDefault.OneAtATime --deployment-group-name $(APP_NAME)-group --s3-location bucket=$(BUCKET_NAME),bundleType=zip,key=$(BUCKET_TYPE)/$(PROJECT_NAME)/$(GIT_HASH)_JENKINS.zip > code_deploy
# 	cat code_deploy | jq '.deploymentId' > deploymentId
#
# .PHONY: wait_deploy
# wait_deploy:
# 	cat deploymentId
# 	$(eval DEPLOYMENT_ID := $(call read_file, deploymentId))
# 	$(eval deployMsg = $(shell aws deploy wait deployment-successful --region $(AWS_DEPLOY_REGION)  --deployment-id $(DEPLOYMENT_ID)))
# 	$(info $(deployMsg))
#
# .PHONY: deploy_clean
# deploy_clean:
# 	echo "Deploy Clean"
# 	rm ./*.zip
#
# .PHONY: deploy
# deploy: config_copy make_zip_file artifacts_to_s3 code_deploy deploy_clean

.PHONY: run
run:
	docker run -it $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_DEPLOY_REGION).amazonaws.com/$(AWS_ECR_REGISTRY):$(AWS_ECR_IMAGE_TAG)

.PHONY: make_docker
make_docker:
	docker build \
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

define read_file
$(shell cat $(1))
endef
