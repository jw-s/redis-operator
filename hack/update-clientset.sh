#!/usr/bin/env bash

PROJECT=github.com/jw-s/redis-operator
OUTPUT_BASE="${GOPATH}/src"

client-gen \
--go-header-file /dev/null \
--output-base ${OUTPUT_BASE} \
--input-base ${PROJECT}/pkg/apis \
--clientset-path ${PROJECT}/pkg/generated \
--input redis/v1 \
--clientset-name clientset