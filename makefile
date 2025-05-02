include data/makefile.tag

FLAGS=-trimpath -buildvcs=false -tags='netgo,osusergo,static_build'
LDFLAGS=-ldflags='-s -w -extldflags "-static"'

default: build

build:
	@go mod tidy
	CGO_ENABLED=0 go build ${FLAGS} ${LDFLAGS} -o ./bin/ .

release: pre release-prod help
release-check:
	goreleaser check
	goreleaser healthcheck
release-prod: release-check
	goreleaser release --verbose --clean --skip=validate
release-dev:
	goreleaser release --verbose --snapshot --clean

pre: 
	@go mod tidy
	# @go fmt ./... && go vet ./...

update:
	go get -u .

stop:
	@pkill -9 dev.sh ||:
	@pkill -9 inotify ||:
	@pkill -9 ssm ||:

help:
	build/ssm_linux_amd64_v1/ssm --help >help.md

clean:
	rm -rf build/*

backup: 
	rm -rf build/*
	tar -czvf ../ssm-$(shell date +%Y%m%d).tgz --exclude='.git' .

.PHONY: build
