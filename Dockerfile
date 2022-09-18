# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN mkdir build/
RUN go build -o ./build ./...

RUN mkdir -p /usr/bin && mv ./build/hoser /usr/bin/hoser

CMD [ "hoser", "run", "-v", "examples/hello.hos" ]