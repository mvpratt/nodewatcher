#!/bin/sh
export LN_NODE_URL=abcxyz.io:10009
export MACAROON_HEADER=ABCD01234567
export LND_TLS_CERT_PATH=/path/to/tls.cert
export SMS_ENABLE=TRUE
export TWILIO_ACCOUNT_SID=ABCD
export TWILIO_AUTH_TOKEN=BEEF42
export TWILIO_PHONE_NUMBER=+15556667777
export TO_PHONE_NUMBER=5554443333
export DOCKER_REPO=11111111.dkr.ecr.us-east-1.amazonaws.com/nodewatcher
export POSTGRES_HOST=localhost
export POSTGRES_DB=depot
export POSTGRES_PORT=5432
export POSTGRES_USER=user
export POSTGRES_PASSWORD=password