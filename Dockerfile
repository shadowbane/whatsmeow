FROM golang:1.19-alpine

MAINTAINER Adli I. Ifkar <adly.shadowbane@gmail.com>
ENV LANG=en_US.UTF-8
ENV DEBIAN_FRONTEND noninteractive

USER root

# GCC
RUN apk add build-base

RUN mkdir -p /opt/application

WORKDIR /opt/application

COPY . .

RUN go build -o /opt/gomeow cmd/api/main.go

WORKDIR /opt
RUN mkdir -p /opt/log && chmod -R 777 /opt/log
RUN rm -rf /opt/applications

CMD ["/opt/gomeow"]