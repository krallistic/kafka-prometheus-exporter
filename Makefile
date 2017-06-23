# Makefile for the Docker image
# MAINTAINER: github.com/krallistic

.PHONY: all build container push clean test

TAG ?= v0.1.0
PREFIX ?= krallistic

all: container

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kafka_exporter main/cmd.go

container: build
	docker build -t $(PREFIX)/kafka_exporter:$(TAG) .

push: container
	docker push $(PREFIX)/kafka_exporter:$(TAG)

clean:
	rm -f kafka_exporter

