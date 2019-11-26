FROM golang:1.10.4-alpine

RUN apk add --no-cache git mercurial \
    && go get "github.com/google/uuid" \
    && go get "github.com/gorilla/websocket"\
    && go get "github.com/streadway/amqp" \
    && apk del git mercurial
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
EXPOSE 8000
CMD ["/app/main"]
