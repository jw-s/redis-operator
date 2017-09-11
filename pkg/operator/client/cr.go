package client

import (
	"github.com/JoelW-S/redis-operator/pkg/operator/spec"
	"github.com/JoelW-S/redis-operator/pkg/operator/util"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"reflect"
)

func CreateCustomResourceDefinition(clientset apiextensionsclient.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.RedisCRDName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   spec.RedisGroupName,
			Version: spec.RedisSchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   spec.RedisNamePlural,
				Singular: spec.RedisNameSingular,
				Kind:     reflect.TypeOf(spec.RedisServer{}).Name(),
			},
		},
	}
	crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)

	if err != nil {
		return nil, err
	}

	return crd, nil
}

type RedisServerCR struct {
	Client     *rest.RESTClient
	CrScheme   *runtime.Scheme
	ParamCodec runtime.ParameterCodec
}

func NewClientInCluster() *RedisServerCR {
	cfg, err := util.InClusterConfig()
	if err != nil {
		panic(err)
	}
	cli, err := NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return cli
}

func NewClient(cfg *rest.Config) (*RedisServerCR, error) {

	cli, crScheme, err := newRedisClient(cfg)
	if err != nil {
		return nil, err
	}
	return &RedisServerCR{
		Client:     cli,
		CrScheme:   crScheme,
		ParamCodec: runtime.NewParameterCodec(crScheme),
	}, nil
}

func newRedisClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	crScheme := runtime.NewScheme()
	if err := spec.AddKnownTypes(crScheme); err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &spec.RedisSchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(crScheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, crScheme, nil
}
