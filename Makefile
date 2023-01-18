.DEFAULT_GOAL := build

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
docker-build: lint
	docker build . -t nodewatcher \
	--build-arg SMS_ENABLE=${SMS_ENABLE} \
	--build-arg LN_NODE_URL=${LN_NODE_URL} \
	--build-arg MACAROON_HEADER=${MACAROON_HEADER} \
	--build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
	--build-arg TO_PHONE_NUMBER=${TO_PHONE_NUMBER}

.PHONY:docker-build-aws
docker-build-aws: lint
	docker build . -t nodewatcher \
	--build-arg SMS_ENABLE=${SMS_ENABLE} \
	--build-arg LN_NODE_URL=${LN_NODE_URL} \
	--build-arg MACAROON_HEADER=${MACAROON_HEADER} \
	--build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
	--build-arg TO_PHONE_NUMBER=${TO_PHONE_NUMBER} \
	--platform=linux/amd64	

.PHONY:docker-run
docker-run:
	docker run --rm nodewatcher nodewatcher

.PHONE:docker-exec
docker-exec:
	docker exec -it nodewatcher bash

.PHONY:deploy
deploy: docker-build-aws
	docker tag nodewatcher:latest ${DOCKER_REPO}:latest
	docker push ${DOCKER_REPO}:latest