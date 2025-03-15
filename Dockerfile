# stage 1
FROM golang:1.24.0-alpine3.20 AS builder

RUN apk update && apk add git

# setting up the go environments
ENV GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0
ENV GOPATH=/app
ENV GOBIN=/go/bin

# create a directory for main app
RUN mkdir -p /app

# set main app run on this directory
WORKDIR /app

# copy all files into main directory above
ADD . /app

# download dependencies and cached
RUN go mod download

# build into binary
RUN go build -o server ./cmd/main.go

# stage 2
FROM alpine:3.20.1

# add tzdata
RUN apk add --no-cache tzdata bash

# set main app run on this directory
WORKDIR /app

# copy all file from the stage 1 into new directory on stage 2
COPY --from=builder /app/server /app/server

RUN chmod +x /app/server
EXPOSE 8000

# execute the service binary
CMD [ "./server" ]