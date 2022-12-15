###### Builder image #####
FROM golang:1.19.2-alpine

# Install build dependencies
RUN apk add --no-cache git make bash vim
RUN go install golang.org/x/lint/golint@latest

# Nodewatchter
WORKDIR /home
RUN git clone https://github.com/mvpratt/nodewatcher.git nodewatcher
RUN cd nodewatcher && make

CMD ["/home/nodewatcher/nw"]