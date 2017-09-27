REDIS_OPERATOR_IMAGE_NAME := redis-operator

.PHONY: all clean build

all: build

build: bin/redis
	docker build -t $(REDIS_OPERATOR_IMAGE_NAME) -f ./Dockerfile_prod .

bin/redis:
	GOOS=linux GOARCH=amd64 go build -o bin/redis-operator ./cmd/operator

deploy: undeploy
	kubectl create -f contrib/kube-redis/redis-cr.yml
	kubectl apply -f contrib/kube-redis/deployment.yml
	kubectl create -f contrib/kube-redis/redis-server.yml

undeploy:
	kubectl delete -f contrib/kube-redis/redis-server.yml || true
	kubectl delete -f contrib/kube-redis/deployment.yml || true
	kubectl delete -f contrib/kube-redis/redis-cr.yml || true

clean: 
	rm -rf bin/*
