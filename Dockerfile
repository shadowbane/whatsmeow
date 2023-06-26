## Build Stage
FROM golang:1.20-alpine AS builder

MAINTAINER Adli I. Ifkar <adly.shadowbane@gmail.com>
ENV LANG=en_US.UTF-8
ENV DEBIAN_FRONTEND noninteractive

USER root

# GCC
RUN apk add build-base

RUN mkdir -p /opt/application

WORKDIR /opt/application

COPY . .

RUN go build -ldflags="-s -w" -o /opt/gomeow cmd/api/main.go

## Deploy Stage
FROM golang:1.20-alpine

MAINTAINER Adli I. Ifkar <adly.shadowbane@gmail.com>
ENV LANG=en_US.UTF-8
ENV DEBIAN_FRONTEND noninteractive

WORKDIR /opt
COPY --from=builder /opt/gomeow /opt/gomeow
RUN mkdir -p /opt/log && chmod -R 777 /opt/log

CMD ["/opt/gomeow"]
