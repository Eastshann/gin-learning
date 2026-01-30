# 变量定义
APP_NAME = webook
VERSION = v0.0.1
DOCKER_USER = gufan

# 编译参数
GO_BIN = go
GOOS = linux
GOARCH = arm
TAGS = k8s

# 镜像名称
IMAGE_NAME = $(DOCKER_USER)/$(APP_NAME):$(VERSION)

.PHONY: all build docker clean help

# 默认执行的目标
all: clean build docker

# 1. 删除二进制文件和镜像
clean:
	@echo [log] Clean webook and docker image ...
	del $(APP_NAME)
	docker rmi -f $(IMAGE_NAME)

# 1. 编译二进制文件
build:
	@echo [log] Compile $(APP_NAME) to $(GOOS)/$(GOARCH) ...
	set GOOS=$(GOOS)&& set GOARCH=$(GOARCH)&& $(GO_BIN) build -tags=$(TAGS) -o $(APP_NAME) main.go

# 2. 构建 Docker 镜像
docker:
	@echo [log] Build $(IMAGE_NAME) ...
	docker build -t $(IMAGE_NAME) .


# 帮助信息
help:
	@echo [log] How to use:
	@echo       make         - Build binary and docker image
	@echo       make clean   - Clean binary and docker image
	@echo       make build   - Compile the go binary
	@echo       make docker  - Build Docker Image