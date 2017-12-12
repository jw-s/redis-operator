package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	redisclient "github.com/jw-s/redis-operator/pkg/generated/clientset/typed/redis/v1"
	"github.com/sirupsen/logrus"
	"github.com/jw-s/redis-operator/pkg/operator/spec"
)

type Config struct {
	RedisCRClient redisclient.RedisesGetter
	RedisClient   *redis.Client
}

type Redis struct {
	logger                    *logrus.Entry
	Redis                     *v1.Redis
	Config                    Config
	SeedMasterProcessComplete bool
	SeedMasterDeleted         bool
}

func New(config Config, redis *v1.Redis) *Redis {
	newRedis := &Redis{
		logger: logrus.WithField("pkg", "redis"),
		Config: config,
	}

	redis.Spec.ApplyDefaults()

	redis.Spec.Sentinels.ApplyDefaults(spec.GetSentinelConfigMapName(redis.Name))

	newRedis.Redis = redis

	return newRedis
}

func (r *Redis) UpdateRedisStatus() error {

	r.logger.Debugf("Updating redis %s", r.Redis.Name)
	defer r.logger.Debugf("Finished updating redis %s", r.Redis.Name)

	newRedis, err := r.Config.RedisCRClient.Redises(r.Redis.Namespace).Update(r.Redis)

	if err != nil {
		return fmt.Errorf("failed to update CR status: %v", err)
	}

	r.Redis = newRedis

	return nil
}

func (r *Redis) ReportFailed() error {
	r.Redis.Status.SetPhase(v1.ServerFailedPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportRunning() error {
	r.Redis.Status.SetPhase(v1.ServerRunningPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportCreating() error {
	r.Redis.Status.SetPhase(v1.ServerCreatingPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportStopping() error {
	r.Redis.Status.SetPhase(v1.ServerStoppingPhase)
	return r.UpdateRedisStatus()
}
