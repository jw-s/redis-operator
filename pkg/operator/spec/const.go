package spec

const (
	RedisPort              = 6379
	RedisSentinelPort      = 26379
	MasterPodName          = "redis-master-%s"
	MasterServiceName      = "redis-master-%s"
	SentinelServiceName    = "redis-sentinel-%s"
	SentinelDeploymentName = "redis-sentinel-%s"
	SlaveDeploymentName    = "redis-slave-%s"
	OperatorLabel          = "redis_operator"
	SentinelConfigMapName  = "sentinel-config-%s"
	ConfigVolumeName       = "config"
	DataVolumeName         = "data"
	ConfigMapConfKeyName   = "redis.conf"
	ConfMountPath          = "/usr/local/etc/redis"
	DataMountPath          = "/data"
)
