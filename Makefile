.DEFAULT_GOAL := build

env:
	source env.sh
lint:
	golint ./...
.PHONY:lint

build: lint
	go build nw.go
.PHONY:build