.SILENT: test build

GORELEASERCMD=goreleaser
FLAGS=--skip-publish --snapshot --rm-dist

all: test build
test:
	${GOTEST} -v -cover
build:
	${GORELEASERCMD} ${FLAGS}