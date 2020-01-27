.PHONY: help her
SHELL := /bin/bash

export PROJECT = her

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

her: ## Build docker image to run her
	docker build \
		-t $(PROJECT)/her-amd64:1.0 \
		--build-arg PACKAGE_NAME=her \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

build: ## Build binary in dist/
	cd cmd/her && go build -mod=readonly -ldflags "-X main.build=`git rev-parse HEAD`" -o ../../dist/her .

build-pi: ## Build binary for RaspberryPI in dist/
	cd cmd/her && env GOOS=linux GOARCH=arm GOARM=5 go build -mod=readonly -ldflags "-X main.build=`git rev-parse HEAD`" -o ../../dist/her-pi .

test: ## Run tests
	go test ./...

clean: ## Clean docker
	docker system prune -f

deps-reset: ## Reset dependencies
	git checkout -- go.mod
	go mod tidy

# deps-upgrade:
# 	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

deps-cleancache: ## Clean deps cache
	go clean -modcache
