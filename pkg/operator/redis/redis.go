package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	redisclient "github.com/jw-s/redis-operator/pkg/generated/clientset/typed/redis/v1"
	"github.com/jw-s/redis-operator/pkg/operator/spec"
	"github.com/sirupsen/logrus"
)

const (
	ReportingPhaseMessage = "Updating phase"
	PhaseKey              = "Phase"
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
		logger: logrus.WithFields(logrus.Fields{"pkg": "redis", "queue_item": redis.Name}),
		Config: config,
	}

	redis.Spec.ApplyDefaults(spec.GetSentinelConfigMapName(redis.Name))

	newRedis.Redis = redis

	return newRedis
}

func (r *Redis) UpdateRedisStatus() error {

	r.logger.Debug("Updating redis status")
	defer r.logger.Debug("Finished updating redis status")

	newRedis, err := r.Config.RedisCRClient.Redises(r.Redis.Namespace).Update(r.Redis)

	if err != nil {
		return fmt.Errorf("failed to update CR status: %v", err)
	}

	r.Redis = newRedis

	return nil
}

func (r *Redis) ReportFailed() error {
	r.logger.
		WithField(PhaseKey, v1.ServerFailedPhase).
		Debug(ReportingPhaseMessage)
	r.Redis.Status.SetPhase(v1.ServerFailedPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportRunning() error {
	r.logger.
		WithField(PhaseKey, v1.ServerRunningPhase).
		Debug(ReportingPhaseMessage)
	r.Redis.Status.SetPhase(v1.ServerRunningPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportCreating() error {
	r.logger.
		WithField(PhaseKey, v1.ServerCreatingPhase).
		Debug(ReportingPhaseMessage)
	r.Redis.Status.SetPhase(v1.ServerCreatingPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) ReportStopping() error {
	r.logger.
		WithField(PhaseKey, v1.ServerStoppingPhase).
		Debug(ReportingPhaseMessage)
	r.Redis.Status.SetPhase(v1.ServerStoppingPhase)
	return r.UpdateRedisStatus()
}

func (r *Redis) MarkReadyCondition() error {
	r.Redis.Status.MarkReadyCondition()
	return r.UpdateRedisStatus()
}

func (r *Redis) MarkAddSeedMasterCondition() error {
	r.Redis.Status.MarkAddSeedMasterCondition()
	return r.UpdateRedisStatus()
}

func (r *Redis) MarkRemoveSeedMasterCondition() error {
	r.Redis.Status.MarkRemoveSeedMasterCondition()
	return r.UpdateRedisStatus()
}
