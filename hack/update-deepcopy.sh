#!/usr/bin/env bash

deepcopy-gen \
--bounding-dirs github.com/jw-s/redis-operator/pkg/apis/redis/v1 \
--input-dirs github.com/jw-s/redis-operator/pkg/apis/redis/v1 \
-h /dev/null \
-O zz_generated.deepcopy
