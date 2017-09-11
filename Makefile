REDIS_OPERATOR_IMAGE_NAME := redis-operator

.PHONY: all clean build

all: build

build: bin/redis
	docker build -t $(REDIS_OPERATOR_IMAGE_NAME) -f ./Dockerfile_prod .

bin/redis: 
	GOOS=linux GOARCH=amd64 go build -o bin/redis-operator ./cmd/operator

clean: 
	rm -rf bin/*
