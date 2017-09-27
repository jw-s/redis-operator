## Install Redis operator

Register CRD in Kubernetes

```bash
$ kubectl create -f contrib/kube-redis/redis-cr.yml
```

Create a deployment for Redis operator:

```bash
$ kubectl create -f contrib/kube-redis/deployment.yml
```

## Uninstall Redis operator

Note that the Redis servers managed by the Redis operator will **NOT** be deleted even if the operator is uninstalled.
In order to delete all Redis servers, delete all CR objects before you uninstall the operator.