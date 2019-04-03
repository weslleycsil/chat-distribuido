FROM golang:1.12.1-alpine3.9

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
EXPOSE 8081
CMD ["/app/main"]
