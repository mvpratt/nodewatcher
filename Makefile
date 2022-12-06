.DEFAULT_GOAL := build

lint:
	golint ./...
.PHONY:lint

build: lint
	go build nw.go
.PHONY:build