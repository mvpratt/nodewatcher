#!/bin/sh
export LND_TLS_CERT_PATH=/path/to/tls.cert

# sms notifications
export TWILIO_ACCOUNT_SID=ABCD
export TWILIO_AUTH_TOKEN=BEEF42
export TWILIO_PHONE_NUMBER=+15556667777

# database credentials
export POSTGRES_HOST=nodewatcher-db
export POSTGRES_DB=postgres
export POSTGRES_PORT=5432
export POSTGRES_USER=nodewatcher
export POSTGRES_PASSWORD=password

# aws
export DOCKER_REPO=11111111.dkr.ecr.us-east-1.amazonaws.com/nodewatcher