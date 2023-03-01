.DEFAULT_GOAL := build

.PHONY:lint
lint:
	golint ./...
	go vet ./...

.PHONY:build
build: lint
	cd cmd/nw && go build
	cd cmd/graphql && go build -o graphql
	cd cmd/rest-api && go build

.PHONY:run
run:
	cd cmd/nw && ./nw

.PHONY:test
test:
	go test -v ./...

.PHONY: gql
gql:
	go generate ./...

common-build-args = --build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
    --build-arg POSTGRES_HOST=${POSTGRES_HOST} \
	--build-arg POSTGRES_DB=${POSTGRES_DB} \
	--build-arg POSTGRES_PORT=${POSTGRES_PORT} \
	--build-arg POSTGRES_USER=${POSTGRES_USER} \
	--build-arg POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

.PHONY:docker-build
docker-build: lint
	docker build . -t nodewatcher $(common-build-args) --build-arg CMD=/bin/nw

.PHONY:docker-build-graphql
docker-build-graphql: lint
	docker build . -t nodewatcher-graphql $(common-build-args) --build-arg CMD=/bin/graphql

.PHONY:docker-run
docker-run:
	docker run --rm -it nodewatcher bash

.PHONE:docker-exec
docker-exec:
	docker exec -it nodewatcher bash