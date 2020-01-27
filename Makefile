#!/usr/bin/make -f 
src_dir      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	GOOS=$(OS) CGO_ENABLED=0 go build -a -ldflags "-X gitlab.seznam.net/sklik-devops/slo-exporter/version.buildRevision=${CI_COMMIT_SHA} -X gitlab.seznam.net/sklik-devops/slo-exporter/version.buildRef=${CI_COMMIT_REF_NAME} -X gitlab.seznam.net/sklik-devops/slo-exporter/version.buildAuthor=${GITLAB_USER_LOGIN} -extldflags '-static'" -o slo_exporter $(src_dir)/cmd/slo_exporter.go

lint:
	go get github.com/mgechev/revive
	revive -formatter friendly $(shell find $(src_dir) -name "*.go" | grep -v "^$(src_dir)/vendor/")

test:
	go test --race -coverprofile=coverage.out $(shell go list ./... | grep -v /vendor/)

compose: build
	docker-compose up --force-recreate --renew-anon-volumes --abort-on-container-exit --remove-orphans --exit-code-from slo-exporter
	docker-compose rm --force --stop -v

.PHONY: build lint test compose