GIT_COMMIT ?= $(shell git rev-parse --short HEAD || echo "GitNotFound")

.PHONY: build
build:
	go build -o "bin/api" cmd/api/main.go

.PHONY: build-docker
build-docker:
	docker build -t scorum/account-svc:${GIT_COMMIT} .
