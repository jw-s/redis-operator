package spec

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	RedisNameSingular = "redisserver"
	RedisNamePlural   = "redisservers"
	RedisGroupName    = "operator.joelws.com"
	RedisCRDName      = RedisNamePlural + "." + RedisGroupName
)

var RedisSchemeGroupVersion = schema.GroupVersion{Group: RedisGroupName, Version: "v1"}
