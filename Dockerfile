###### Builder image ######
FROM golang:1.19.2-alpine as builder

# Devtools
RUN apk add --no-cache make bash vim jq

# Install dependencies
RUN go install golang.org/x/lint/golint@latest

# Build from source
RUN mkdir /home/nodewatcher
WORKDIR /home/nodewatcher
COPY . .
RUN make build

###### Final image optimized for size ######
FROM alpine as final

RUN apk add --no-cache bash vim

ARG SMS_ENABLE
ARG LN_NODE_URL
ARG MACAROON_HEADER
ARG LND_TLS_CERT_PATH
ARG TWILIO_ACCOUNT_SID
ARG TWILIO_AUTH_TOKEN
ARG TWILIO_PHONE_NUMBER
ARG TO_PHONE_NUMBER
ARG POSTGRES_HOST
ARG POSTGRES_DB
ARG POSTGRES_PORT
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD

ENV SMS_ENABLE=${SMS_ENABLE}
ENV LN_NODE_URL=${LN_NODE_URL}
ENV MACAROON_HEADER=${MACAROON_HEADER}
ENV LND_TLS_CERT_PATH=${LND_TLS_CERT_PATH}
ENV TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID}
ENV TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN}
ENV TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER}
ENV TO_PHONE_NUMBER=${TO_PHONE_NUMBER}
ENV POSTGRES_HOST=${POSTGRES_HOST}
ENV POSTGRES_DB=${POSTGRES_DB}
ENV POSTGRES_PORT=${POSTGRES_PORT}
ENV POSTGRES_USER=${POSTGRES_USER}
ENV POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

RUN mkdir /home/nodewatcher
WORKDIR /home/nodewatcher

# get database migrations
COPY --from=builder /home/nodewatcher/db ./db

# get tls cert for the lightning node
COPY --from=builder /home/nodewatcher/creds ./creds

COPY --from=builder /home/nodewatcher/nw /bin/
CMD ["/bin/nw"]