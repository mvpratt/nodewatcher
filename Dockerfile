###### Builder image ######
FROM golang:1.19.2-alpine as builder

# Devtools
RUN apk add --no-cache make bash vim jq

# Install dependencies
RUN go install golang.org/x/lint/golint@latest

# Disable CGO to fix this error when building nodewatcher:
#     go: downloading golang.org/x/text v0.3.8
#     # runtime/cgo
#     _cgo_export.c:3:10: fatal error: stdlib.h: No such file or directory
#         3 | #include <stdlib.h>
#           |          ^~~~~~~~~~
#     compilation terminated.
ENV CGO_ENABLED=0

# Build from source
RUN mkdir /home/nodewatcher
WORKDIR /home/nodewatcher
COPY . .
RUN make build

###### Final image optimized for size ######
FROM alpine as final

RUN apk add --no-cache bash vim

ARG TWILIO_ACCOUNT_SID
ARG TWILIO_AUTH_TOKEN
ARG TWILIO_PHONE_NUMBER
ARG POSTGRES_HOST
ARG POSTGRES_DB
ARG POSTGRES_PORT
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD

ENV TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID}
ENV TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN}
ENV TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER}
ENV POSTGRES_HOST=${POSTGRES_HOST}
ENV POSTGRES_DB=${POSTGRES_DB}
ENV POSTGRES_PORT=${POSTGRES_PORT}
ENV POSTGRES_USER=${POSTGRES_USER}
ENV POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

RUN mkdir /home/nodewatcher
WORKDIR /home/nodewatcher

COPY --from=builder /home/nodewatcher/cmd/nw/nw /bin/
COPY --from=builder /home/nodewatcher/cmd/graphql/graphql /bin/
COPY --from=builder /home/nodewatcher/cmd/rest-api/rest-api /bin/

ARG CMD
ENV CMD=${CMD}
CMD ${CMD}