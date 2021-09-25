.SILENT: test build

GOTEST=go test
GORELEASERCMD=goreleaser
FLAGS=--skip-publish --snapshot --rm-dist

all: test build
test:
	${GOTEST} ./... -timeout 30s -v -cover
build:
	${GORELEASERCMD} ${FLAGS}