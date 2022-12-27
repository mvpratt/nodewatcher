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
	docker build . -t nodewatcher \
	--build-arg SMS_ENABLE=${SMS_ENABLE} \
	--build-arg LN_NODE_URL=${LN_NODE_URL} \
	--build-arg MACAROON_HEADER=${MACAROON_HEADER} \
	--build-arg TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID} \
	--build-arg TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN} \
	--build-arg TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER} \
	--build-arg TO_PHONE_NUMBER=${TO_PHONE_NUMBER}

.PHONY:docker-run
docker-run:
	docker run nodewatcher

.PHONY:deploy
deploy:
	docker tag nodewatcher:latest 314833960266.dkr.ecr.us-east-1.amazonaws.com/nodewatcher:latest
	docker push 314833960266.dkr.ecr.us-east-1.amazonaws.com/nodewatcher:latest