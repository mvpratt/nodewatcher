###### Builder image #####
FROM golang:1.19.2-alpine as builder

# Install build dependencies
RUN apk add --no-cache git make bash vim
RUN go install golang.org/x/lint/golint@latest

# Nodewatchter
WORKDIR /home
RUN git clone https://github.com/mvpratt/nodewatcher.git nodewatcher
RUN cd nodewatcher && make


###### Final image optimized for size
FROM alpine as final

ARG SMS_ENABLE
ARG LN_NODE_URL
ARG MACAROON_HEADER
ARG TWILIO_ACCOUNT_SID
ARG TWILIO_AUTH_TOKEN
ARG TWILIO_PHONE_NUMBER
ARG TO_PHONE_NUMBER

ENV SMS_ENABLE=${SMS_ENABLE}
ENV LN_NODE_URL=${LN_NODE_URL}
ENV MACAROON_HEADER=${MACAROON_HEADER}
ENV TWILIO_ACCOUNT_SID=${TWILIO_ACCOUNT_SID}
ENV TWILIO_AUTH_TOKEN=${TWILIO_AUTH_TOKEN}
ENV TWILIO_PHONE_NUMBER=${TWILIO_PHONE_NUMBER}
ENV TO_PHONE_NUMBER=${TO_PHONE_NUMBER}

RUN apk add --no-cache bash vim
COPY --from=builder /home/nodewatcher/nw /bin/

CMD ["/bin/nw"]