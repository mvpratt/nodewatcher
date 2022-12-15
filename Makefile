.DEFAULT_GOAL := build

.PHONY:env
env:
	source env.sh

.PHONY:lint
lint:
	golint ./...

.PHONY:build
build: lint
	go build nw.go

.PHONY:run
run:
	./nw

.PHONY:docker-build
docker-build:
	docker build --no-cache . -t nodewatcher

.PHONY:docker-run
docker-run:
	docker run \
	-e SMS_ENABLE \
	-e LN_NODE_URL \
	-e MACAROON_HEADER \
	-e TWILIO_ACCOUNT_SID \
	-e TWILIO_AUTH_TOKEN \
	-e TWILIO_PHONE_NUMBER \
	-e TO_PHONE_NUMBER \
	nodewatcher