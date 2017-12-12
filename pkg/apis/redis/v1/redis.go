package v1

import (
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServerSpec   `json:"spec"`
	Status            ServerStatus `json:"status"`
}

func (c *Redis) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       CRDResourceKind,
		Name:       c.GetName(),
		UID:        c.GetUID(),
		Controller: &trueVar,
	}
}

const (
	defaultBaseImage = "redis"
	defaultVersion   = "4.0-alpine"
	DefaultPVSize				   = "500Mi"
)

type ServerSpec struct {
	Sentinels SentinelSpec `json:"sentinels"`

	Slaves SlaveSpec `json:"slaves"`

	BaseImage string `json:"baseImage,omitempty"`

	Version string `json:"version,omitempty"`

	Paused bool `json:"paused,omitempty"`

	Pod *PodPolicy `json:"pod,omitempty"`
}

type SentinelSpec struct {
	Replicas  int32     `json:"replicas"`
	Quorum    int32     `json:"quorum"`
	ConfigMap ConfigMap `json:"configMap"`
}

type SlaveSpec struct {
	Replicas  int32     `json:"replicas"`
	ConfigMap ConfigMap `json:"configMap"`
}

type ConfigMap string

type ServerPhase string

const (
	ServerCreatingPhase ServerPhase = "Creating"
	ServerStoppingPhase ServerPhase = "Stopping"
	ServerRunningPhase  ServerPhase = "Running"
	ServerFailedPhase   ServerPhase = "Failed"
)

type ServerConditionType string

type ServerCondition struct {
	Type   ServerConditionType `json:"type"`
	Reason string              `json:"reason,omitempty"`
}

const (
	ServerConditionAddSeedMaster    ServerConditionType = "AddingSeedMaster"
	ServerConditionRemoveSeedMaster ServerConditionType = "removingSeedMaster"
	ServerConditionAddSentinel      ServerConditionType = "AddingSentinel"
	ServerConditionRemoveSentinel   ServerConditionType = "removingSentinel"
	ServerConditionAddSlave         ServerConditionType = "AddingSlave"
	ServerConditionRemoveSlave      ServerConditionType = "removingSlave"
	ServerConditionReady            ServerConditionType = "Ready"
)

type PodPolicy struct {
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type ServerStatus struct {
	Phase          ServerPhase       `json:"phase"`
	Conditions     []ServerCondition `json:"conditions"`
	SlaveStatus    SlaveStatus       `json:"slaves"`
	SentinelStatus SentinelStatus    `json:"sentinels"`
}

type SlaveStatus struct {
	Ready   []string `json:"ready,omitempty"`
	Unready []string `json:"unready,omitempty"`
}

type SentinelStatus struct {
	Ready   []string `json:"ready,omitempty"`
	Unready []string `json:"unready,omitempty"`
}

func (ss *ServerStatus) SetPhase(phase ServerPhase) {
	ss.Phase = phase
}

func (ss *ServerStatus) MarkAddSeedMasterCondition() {

	condition := ServerCondition{
		Type:   ServerConditionAddSeedMaster,
		Reason: "Adding the Redis Seed Master",
	}

	ss.markCondition(condition)

}

func (ss *ServerStatus) MarkRemoveSeedMasterCondition() {
	condition := ServerCondition{
		Type:   ServerConditionRemoveSeedMaster,
		Reason: "Removing the Redis Seed Master",
	}

	ss.markCondition(condition)
}

func (ss *ServerStatus) MarkAddSentinelCondition() {

	condition := ServerCondition{
		Type:   ServerConditionAddSentinel,
		Reason: "Adding a Redis sentinel",
	}

	ss.markCondition(condition)

}

func (ss *ServerStatus) MarkRemoveSentinelCondition() {
	condition := ServerCondition{
		Type:   ServerConditionRemoveSentinel,
		Reason: "Removing a Redis sentinel",
	}

	ss.markCondition(condition)
}

func (ss *ServerStatus) MarkAddSlaveCondition() {

	condition := ServerCondition{
		Type:   ServerConditionAddSlave,
		Reason: "Adding a Redis slave",
	}

	ss.markCondition(condition)

}

func (ss *ServerStatus) MarkRemoveSlaveCondition() {
	condition := ServerCondition{
		Type:   ServerConditionRemoveSlave,
		Reason: "Removing a Redis slave",
	}

	ss.markCondition(condition)
}

func (ss *ServerStatus) MarkReadyCondition() {
	condition := ServerCondition{
		Type:   ServerConditionReady,
		Reason: "Server ready",
	}

	ss.markCondition(condition)
}

func (ss *ServerStatus) markCondition(sc ServerCondition) {

	if len(ss.Conditions) == 10 {
		ss.Conditions = append(ss.Conditions[1:], sc)
		return
	}

	ss.Conditions = append(ss.Conditions, sc)

}

func (s *ServerSpec) ApplyDefaults() {
	if len(s.BaseImage) == 0 {
		logrus.Warnf("Using default image: %s", defaultBaseImage)
		s.BaseImage = defaultBaseImage
	}

	if len(s.Version) == 0 {
		logrus.Warnf("Using default image version: %s", defaultVersion)
		s.Version = defaultVersion
	}

	if s.Sentinels.Replicas != 0 && s.Sentinels.Replicas%2 == 0 {
		logrus.Warn("Redis sentinels should be an odd number to prevent ties!")
	}

	if s.Pod == nil {
		logrus.Warnf("Using default size for PV's: %s", DefaultPVSize)
		s.Pod = &PodPolicy{
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultPVSize),
				},
			},
		}
	}
}

func (s *SentinelSpec) ApplyDefaults(name string) {

	if len(s.ConfigMap) == 0 {
		logrus.Warnf("Using default configMap version: %s", name)
		logrus.Warnf("This configMap will be created if it doesn't already exist.")
		s.ConfigMap = ConfigMap(name)
	}
}
