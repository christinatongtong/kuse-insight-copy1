VS ?= 0.1
APP_NAME=user-insight-job
ENV=staging
DOCKER_IMAGE=$(shell echo $(APP_NAME) |awk -F '@' '{print "us-east4-docker.pkg.dev/kuse-ai-staging/cloud-run-source-deploy/kuse-insight-go/" $$0 ":v${VS}"}')

ifeq ($(ENV),prd)
	DOCKER_IMAGE=$(shell echo $(APP_NAME) |awk -F '@' '{print "us-east4-docker.pkg.dev/kuse-ai/cloud-run-source-deploy/kuse-insight-go/" $$0 ":v${VS}"}')
endif

.PHONY: all
all: image push

.PHONY: clean
clean:
	@go clean
	@rm -rf ./bin

.PHONY: mod
tidy:
	@go mod tidy

.PHONY: build
build:
	@GOOS=linux GOARCH=amd64 go build -o ./bin/$(BIN_NAME) 

.PHONY: tgz
tgz:
	tar -czf ./bin/$(BIN_NAME).tgz ./bin/$(BIN_NAME)

.PHONY: gcloud
auth:
	gcloud auth login
bigquery:
	gcloud auth application-default login

.PHONY: docker
image:
	docker build . -t $(DOCKER_IMAGE) --platform=linux/amd64 --progress=plain
push:
	docker push $(DOCKER_IMAGE)


