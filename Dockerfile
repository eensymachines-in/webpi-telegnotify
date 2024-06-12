FROM golang:1.21.8-alpine3.19

RUN apk add git
RUN mkdir -p /usr/eensy/telegnotify /var/log/eensy/telegnotify /usr/bin/eensy/telegnotify
WORKDIR /usr/eensy/telegnotify
RUN chmod -R +x /usr/bin/eensy/telegnotify

COPY go.mod .
COPY go.sum .
RUN go mod download 
COPY . .

RUN go build -o /usr/bin/eensy/telegnotify/telegnotify .
ENTRYPOINT /usr/bin/eensy/telegnotify/telegnotify