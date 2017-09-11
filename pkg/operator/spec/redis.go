package spec

import (
	"encoding/json"
	"errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultBaseImage = "joelw-s/redis-operator"
	defaultVersion   = "0.1"
)

type RedisServerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisServer `json:"items"`
}

type RedisServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServerSpec   `json:"spec"`
	Status            ServerStatus `json:"status"`
}

func (c *RedisServer) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &trueVar,
	}
}

type ServerSpec struct {
	Size int `json:"size"`

	BaseImage string `json:"baseImage"`

	Version string `json:"version,omitempty"`

	Paused bool `json:"paused,omitempty"`

	Pod *PodPolicy `json:"pod,omitempty"`
}

type PodPolicy struct {
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be eligible
	// to run on a node, the node must have each of the indicated key-value pairs as
	// labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	AntiAffinity bool `json:"antiAffinity,omitempty"`
}

func (c *ServerSpec) Validate() error {
	if c.Pod != nil {
		for k := range c.Pod.Labels {
			if k == "app" || strings.HasPrefix(k, "redis_") {
				return errors.New("spec: pod labels contains reserved label")
			}
		}
	}
	return nil
}

func (c *ServerSpec) Cleanup() {
	if len(c.BaseImage) == 0 {
		c.BaseImage = defaultBaseImage
	}

	if len(c.Version) == 0 {
		c.Version = defaultVersion
	}

	c.Version = strings.TrimLeft(c.Version, "v")
}

type ServerPhase string

const (
	ServerPhaseNone     ServerPhase = ""
	ServerPhaseCreating             = "Creating"
	ServerPhaseRunning              = "Running"
	ServerPhaseFailed               = "Failed"
)

type ServerStatus struct {
	// Phase is the cluster running phase
	Phase  ServerPhase `json:"phase"`
	Reason string      `json:"reason"`

	ControlPaused bool `json:"controlPaused"`

	Size int `json:"size"`

	Sentinels MembersStatus `json:"sentinels"`

	Slaves MembersStatus `json:"slaves"`

	CurrentVersion string `json:"currentVersion"`

	TargetVersion string `json:"targetVersion"`
}

type MembersStatus struct {
	Ready   []string `json:"ready,omitempty"`
	Unready []string `json:"unready,omitempty"`
}

func (cs ServerStatus) Copy() ServerStatus {
	newCS := ServerStatus{}
	b, err := json.Marshal(cs)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &newCS)
	if err != nil {
		panic(err)
	}
	return newCS
}

func (cs *ServerStatus) IsFailed() bool {
	if cs == nil {
		return false
	}
	return cs.Phase == ServerPhaseFailed
}

func (cs *ServerStatus) SetPhase(p ServerPhase) {
	cs.Phase = p
}

func (cs *ServerStatus) PauseControl() {
	cs.ControlPaused = true
}

func (cs *ServerStatus) Control() {
	cs.ControlPaused = false
}

func (cs *ServerStatus) UpgradeVersionTo(v string) {
	cs.TargetVersion = v
}

func (cs *ServerStatus) SetVersion(v string) {
	cs.TargetVersion = ""
	cs.CurrentVersion = v
}

func (cs *ServerStatus) SetReason(r string) {
	cs.Reason = r
}
