REDIS_OPERATOR_IMAGE_NAME := joelws/redis-operator
VERSION=$(shell git rev-parse --short HEAD)
.PHONY: all clean build

all: build

build: bin/redis
	docker build -t $(REDIS_OPERATOR_IMAGE_NAME):$(VERSION) -f ./Dockerfile_prod .

bin/redis: install-deps
	GOOS=linux GOARCH=amd64 go build -o bin/redis-operator ./cmd/operator

install-deps:
	glide up -v

push:
	docker push $(REDIS_OPERATOR_IMAGE_NAME):$(VERSION)

clean: 
	rm -rf bin/*
