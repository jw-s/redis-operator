package controller

import (
	"encoding/json"
	"reflect"

	redisv1 "github.com/jw-s/redis-operator/pkg/apis/redis/v1"
	"github.com/jw-s/redis-operator/pkg/operator/redis"
	"github.com/jw-s/redis-operator/pkg/operator/spec"
	"github.com/jw-s/redis-operator/pkg/operator/util"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubernetes/plugin/pkg/scheduler/api"
)

func (c *RedisController) reconcile(redis *redis.Redis) error {

	var masterIP string

	if !redis.SeedMasterProcessComplete {
		if err := c.seedMasterPodProcess(redis.Redis); err != nil {
			return err
		}

		seedIP, err := util.GetSeedMasterIP(c.podLister, redis.Redis.Namespace, redis.Redis.Name)

		if err != nil {
			return err
		}

		masterIP = seedIP

		redis.SeedMasterProcessComplete = true

	} else {
		ip, err := util.GetMasterIPByName(redis.Config.RedisClient, spec.GetRedisMasterName(redis.Redis))

		if err != nil {
			// Something went wrong, mark to spin up seed pod on next run
			redis.SeedMasterProcessComplete = false
			return err
		}

		masterIP = ip

		deletePolicy := v1.DeletePropagationForeground

		err = c.kubernetesClient.CoreV1().Pods(redis.Redis.Namespace).Delete(spec.GetMasterPodName(redis.Redis.Name),
			&metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			})

		if err != nil && !util.ResourceNotFoundError(err) {
			redis.SeedMasterDeleted = false
			return err
		}

		redis.SeedMasterDeleted = true
	}

	if err := c.masterEndpointProcess(redis.Redis, masterIP); err != nil {
		return err
	}

	if err := c.masterServiceProcess(redis.Redis); err != nil {
		return err
	}

	if err := c.sentinelConfigProcess(redis.Redis); err != nil {
		return err
	}

	if err := c.sentinelServiceProcess(redis.Redis); err != nil {
		return err
	}

	if err := c.sentinelProcess(redis.Redis); err != nil {
		return err
	}

	if err := c.slaveProcess(redis.Redis); err != nil {
		return err
	}

	return nil

}
func (c *RedisController) seedMasterPodProcess(redis *redisv1.Redis) error {

	_, err := c.podLister.Pods(redis.Namespace).Get(spec.GetMasterPodName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {
			if _, err := c.kubernetesClient.CoreV1().Pods(redis.Namespace).Create(spec.MasterSeedPod(redis)); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil

}

func (c *RedisController) sentinelProcess(redis *redisv1.Redis) error {

	actual, err := c.deploymentLister.Deployments(redis.Namespace).Get(spec.GetSentinelDeploymentName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredDeployment(redis, spec.SentinelDeployment)
		}

		return err
	}

	obj, err := api.Scheme.DeepCopy(actual)

	if err != nil {
		return err
	}

	return c.updateDeploymentToDesired(redis, obj.(*appsv1beta1.Deployment), spec.SentinelDeployment)

}

func (c *RedisController) sentinelConfigProcess(redis *redisv1.Redis) error {

	configMapName := string(redis.Spec.Sentinels.ConfigMap)

	_, err := c.configMapLister.ConfigMaps(redis.Namespace).Get(configMapName)

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredConfigMap(redis, spec.DefaultSentinelConfig)
		}

		return err
	}

	return nil
}

func (c *RedisController) slaveProcess(redis *redisv1.Redis) error {

	actual, err := c.statefulSetLister.StatefulSets(redis.Namespace).Get(spec.GetSlaveStatefulSetName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredStatefulSet(redis, spec.SlaveStatefulSet)
		}

		return err
	}

	obj, err := api.Scheme.DeepCopy(actual)

	if err != nil {
		return err
	}

	return c.updateStateFulSetToDesired(redis, obj.(*appsv1beta1.StatefulSet), spec.SlaveStatefulSet)

}

func (c *RedisController) sentinelServiceProcess(redis *redisv1.Redis) error {

	actual, err := c.serviceLister.Services(redis.Namespace).Get(spec.GetSentinelServiceName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredService(redis, spec.SentinelService)

		}

		return err
	}

	obj, err := api.Scheme.DeepCopy(actual)

	if err != nil {
		return err
	}

	return c.updateServiceToDesired(redis, obj.(*apiv1.Service), spec.SentinelService)

}

func (c *RedisController) masterServiceProcess(redis *redisv1.Redis) error {

	actual, err := c.serviceLister.Services(redis.Namespace).Get(spec.GetMasterServiceName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredService(redis, spec.MasterService)
		}

		return err
	}

	obj, err := api.Scheme.DeepCopy(actual)

	if err != nil {
		return err
	}

	return c.updateServiceToDesired(redis, obj.(*apiv1.Service), spec.MasterService)
}

func (c *RedisController) masterEndpointProcess(redis *redisv1.Redis, ipAddress string) error {

	actual, err := c.endpointsLister.Endpoints(redis.Namespace).Get(spec.GetMasterServiceName(redis.Name))

	if err != nil {
		if util.ResourceNotFoundError(err) {

			return c.createDesiredEndpoint(redis, ipAddress, spec.MasterServiceEndpoint)
		}

		return err
	}

	obj, err := api.Scheme.DeepCopy(actual)

	if err != nil {
		return err
	}

	return c.updateEndpointToDesired(redis, obj.(*apiv1.Endpoints), ipAddress, spec.MasterServiceEndpoint)

}

func (c *RedisController) createDesiredDeployment(redis *redisv1.Redis, desiredFunction func(*redisv1.Redis) *appsv1beta1.Deployment) (err error) {

	_, err = c.kubernetesClient.AppsV1beta1().Deployments(redis.Namespace).Create(desiredFunction(redis))

	return

}

func (c *RedisController) createDesiredStatefulSet(redis *redisv1.Redis, desiredFunction func(*redisv1.Redis) *appsv1beta1.StatefulSet) (err error) {

	_, err = c.kubernetesClient.AppsV1beta1().StatefulSets(redis.Namespace).Create(desiredFunction(redis))

	return

}

func (c *RedisController) createDesiredConfigMap(redis *redisv1.Redis, desiredFunction func(*redisv1.Redis) *apiv1.ConfigMap) (err error) {

	_, err = c.kubernetesClient.CoreV1().ConfigMaps(redis.Namespace).Create(desiredFunction(redis))

	return
}

func (c *RedisController) createDesiredService(redis *redisv1.Redis, desiredFunction func(*redisv1.Redis) *apiv1.Service) (err error) {

	_, err = c.kubernetesClient.CoreV1().Services(redis.Namespace).Create(desiredFunction(redis))

	return

}

func (c *RedisController) createDesiredEndpoint(redis *redisv1.Redis, ipAddress string, desiredFunction func(*redisv1.Redis, string) *apiv1.Endpoints) (err error) {

	_, err = c.kubernetesClient.CoreV1().Endpoints(redis.Namespace).Create(desiredFunction(redis, ipAddress))

	return

}

func (c *RedisController) updateDeploymentToDesired(redis *redisv1.Redis, actual *appsv1beta1.Deployment, desiredFunction func(*redisv1.Redis) *appsv1beta1.Deployment) error {

	desired := desiredFunction(redis)

	if !reflect.DeepEqual(actual, desired) {

		actualJSON, err := json.Marshal(actual)
		if err != nil {
			return err
		}
		desiredJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}

		patch, err := strategicpatch.CreateTwoWayMergePatch(actualJSON, desiredJSON, appsv1beta1.Deployment{})

		if err != nil {
			return err
		}

		_, err = c.kubernetesClient.AppsV1beta1().Deployments(redis.Namespace).Patch(actual.Name, types.StrategicMergePatchType, patch)

		return err
	}

	return nil
}

func (c *RedisController) updateStateFulSetToDesired(redis *redisv1.Redis, actual *appsv1beta1.StatefulSet, desiredFunction func(*redisv1.Redis) *appsv1beta1.StatefulSet) error {

	desired := desiredFunction(redis)

	if !reflect.DeepEqual(actual, desired) {

		actualJSON, err := json.Marshal(actual)
		if err != nil {
			return err
		}
		desiredJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}

		patch, err := strategicpatch.CreateTwoWayMergePatch(actualJSON, desiredJSON, appsv1beta1.StatefulSet{})

		if err != nil {
			return err
		}

		_, err = c.kubernetesClient.AppsV1beta1().StatefulSets(redis.Namespace).Patch(actual.Name, types.StrategicMergePatchType, patch)

		return err
	}

	return nil
}

func (c *RedisController) updateServiceToDesired(redis *redisv1.Redis, actual *apiv1.Service, desiredFunction func(*redisv1.Redis) *apiv1.Service) error {

	desired := desiredFunction(redis)

	if !reflect.DeepEqual(actual, desired) {

		desiredJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}

		_, err = c.kubernetesClient.CoreV1().Services(redis.Namespace).Patch(actual.Name, types.StrategicMergePatchType, desiredJSON)

		return err
	}

	return nil
}

func (c *RedisController) updateConfigMapToDesired(redis *redisv1.Redis, actual *apiv1.ConfigMap, desiredFunction func(*redisv1.Redis) *apiv1.ConfigMap) error {

	desired := desiredFunction(redis)

	if !reflect.DeepEqual(actual, desired) {
		desiredJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}

		_, err = c.kubernetesClient.CoreV1().ConfigMaps(redis.Namespace).Patch(actual.Name, types.StrategicMergePatchType, desiredJSON)

		return err
	}

	return nil
}

func (c *RedisController) updateEndpointToDesired(redis *redisv1.Redis, actual *apiv1.Endpoints, ipAddress string, desiredFunction func(*redisv1.Redis, string) *apiv1.Endpoints) error {

	desired := desiredFunction(redis, ipAddress)

	if !reflect.DeepEqual(actual, desired) {
		desiredJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}

		_, err = c.kubernetesClient.CoreV1().Endpoints(redis.Namespace).Patch(actual.Name, types.StrategicMergePatchType, desiredJSON)

		return err
	}

	return nil
}

func (c *RedisController) deleteResources(namespace, name string) error {

	deletePolicy := metav1.DeletePropagationBackground

	err := c.kubernetesClient.CoreV1().Pods(namespace).Delete(spec.GetMasterPodName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.AppsV1beta1().Deployments(namespace).Delete(spec.GetSlaveStatefulSetName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.AppsV1beta1().Deployments(namespace).Delete(spec.GetSentinelDeploymentName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.CoreV1().Services(namespace).Delete(spec.GetSentinelServiceName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.CoreV1().ConfigMaps(namespace).Delete(spec.GetSentinelConfigMapName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.AppsV1beta1().StatefulSets(namespace).Delete(spec.GetSlaveStatefulSetName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	err = c.kubernetesClient.CoreV1().Services(namespace).Delete(spec.GetMasterServiceName(name), &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

	if err != nil && !util.ResourceNotFoundError(err) {
		return err
	}

	return nil
}
