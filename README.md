# Redis Operator

## Status

*Development*

### Overview

The *Redis* operator manages redis servers deployed to *Kubernetes* and automates tasks related to operating a Redis server setup.

## Requirements

- Kubernetes 1.7+

## Demo

## Getting started

### Deploy Redis operator

See [instructions on how to install Redis operator](doc/INSTALL.md)

### Create and destroy an Redis Server

See [operator flow](doc/design/flow.md)

### Create

```bash
$ kubectl create -f contrib/kube-redis/redis-server.yml
```

See [Redis CRD schema](pkg/apis/redis/v1/redis.go)

A 3 member redis service will be created. (deployment resource of 3 replicas)

```bash
$ kubectl get pods
NAME                                        READY     STATUS    RESTARTS   AGE
redis-slave-my-redis-2728646000-2hctj       1/1       Running   0          1m
redis-slave-my-redis-2728646000-70n86       1/1       Running   0          1m
redis-slave-my-redis-2728646000-npvgb       1/1       Running   0          1m
```

### Delete
```bash
$ kubectl delete -f contrib/kube-redis/redis-server.yml
```