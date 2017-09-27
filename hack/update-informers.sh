#!/usr/bin/env bash

PROJECT=github.com/jw-s/redis-operator
OUTPUT_BASE="${GOPATH}/src"

informer-gen \
  --logtostderr \
  --go-header-file /dev/null \
  --output-base ${OUTPUT_BASE} \
  --input-dirs ${PROJECT}/pkg/apis/redis/v1 \
  --output-package ${PROJECT}/pkg/generated/informers \
  --listers-package ${PROJECT}/pkg/generated/listers \
  --internal-clientset-package ${PROJECT}/pkg/generated/clientset \
  --versioned-clientset-package ${PROJECT}/pkg/generated/clientset \