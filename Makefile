.DEFAULT_GOAL := build

.PHONY:lint
lint:
	golint ./...
	go vet ./...

.PHONY:build
build: lint
	cd cmd/nw && go build -o nw
	cd cmd/nwapi && go build -o server

.PHONY:run
run:
	cd cmd/nw && ./nw

.PHONY:test
test:
	go test ./...

.PHONY: gql
gql:
	go generate ./...

.PHONY:docker-build
docker-build: lint
	docker build . -t nodewatcher \
	--build-arg SMS_ENABLE=${SMS_ENABLE} \
	--build-arg LN_NODE_URL=${LN_NODE_URL} \
	--build-arg MACAROON_HEADER=${MACAROON_HEADER} \
	--build-arg LND_TLS_CERT_PATH=${LND_TLS_CERT_PATH} \
	--build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
	--build-arg TO_PHONE_NUMBER=${TO_PHONE_NUMBER} \
    --build-arg POSTGRES_HOST=${POSTGRES_HOST} \
	--build-arg POSTGRES_DB=${POSTGRES_DB} \
	--build-arg POSTGRES_PORT=${POSTGRES_PORT} \
	--build-arg POSTGRES_USER=${POSTGRES_USER} \
	--build-arg POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \

.PHONY:docker-build-aws
docker-build-aws: lint
	docker build . -t nodewatcher \
	--build-arg SMS_ENABLE=${SMS_ENABLE} \
	--build-arg LN_NODE_URL=${LN_NODE_URL} \
	--build-arg MACAROON_HEADER=${MACAROON_HEADER} \
	--build-arg LND_TLS_CERT_PATH=${LND_TLS_CERT_PATH} \
	--build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
	--build-arg TO_PHONE_NUMBER=${TO_PHONE_NUMBER} \
	--build-arg POSTGRES_HOST=${POSTGRES_HOST} \
	--build-arg POSTGRES_DB=${POSTGRES_DB} \
	--build-arg POSTGRES_PORT=${POSTGRES_PORT} \
	--build-arg POSTGRES_USER=${POSTGRES_USER} \
	--build-arg POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	--platform=linux/amd64

.PHONY:docker-run
docker-run:
	docker run --rm nodewatcher

.PHONE:docker-exec
docker-exec:
	docker exec -it nodewatcher bash

.PHONY:deploy
deploy:
	docker tag nodewatcher:latest ${DOCKER_REPO}:latest
	docker push ${DOCKER_REPO}:latest