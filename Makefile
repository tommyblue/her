SHELL := /bin/bash

export PROJECT = her

all: her

her:
	docker build \
		-t $(PROJECT)/her-amd64:1.0 \
		--build-arg PACKAGE_NAME=her \
		--build-arg TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN} \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.
# up:
# 	docker-compose up

# down:
# 	docker-compose down

# test:
# 	go test ./...

# clean:
# 	docker system prune -f

# deps-reset:
# 	git checkout -- go.mod
# 	go mod tidy

# deps-upgrade:
# 	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

# deps-cleancache:
# 	go clean -modcache