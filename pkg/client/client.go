package client

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedAuthv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"

	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
)

type ModeType uint

const (
	Auto ModeType = iota
	Explicit
)

type Client struct {
	ResourceMode            ModeType
	NamespaceMode           ModeType
	SkipSubjectAccessChecks bool
	RESTConfig              *rest.Config
	Logger                  *zap.Logger

	ClientsetFn func(context.Context, *rest.Config) (kubernetes.Interface, error)
	clientset   kubernetes.Interface

	DynamicClientFn func(context.Context, *rest.Config) (dynamic.Interface, error)
	dynamic         dynamic.Interface

	ServerResourcesFn func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error)
	serverResources   discovery.ServerResourcesInterface

	SubjectAccessFn func(context.Context, kubernetes.Interface) (typedAuthv1.SelfSubjectAccessReviewInterface, error)
	subjectAccess   typedAuthv1.SelfSubjectAccessReviewInterface

	mu sync.Mutex
}

func NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	defer logging.Logger.Sync()

	c := &Client{
		ResourceMode:            Auto,
		NamespaceMode:           Auto,
		SkipSubjectAccessChecks: false,
		Logger:                  logging.Logger,
		ClientsetFn:             NewClientset,
		DynamicClientFn:         NewDynamicClient,
		ServerResourcesFn:       NewServerResources,
		SubjectAccessFn:         NewSubjectAccess,
	}

	for _, opt := range options {
		opt(c)
	}

	if err := c.UpdateRESTConfig(ctx, c.RESTConfig); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) UpdateRESTConfig(ctx context.Context, config *rest.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := CheckRestConfig(ctx, config, c.Logger); err != nil {
		return err
	}
	c.RESTConfig = config

	clientset, err := c.ClientsetFn(ctx, c.RESTConfig)
	if err != nil {
		return err
	}
	c.clientset = clientset

	dynclient, err := c.DynamicClientFn(ctx, c.RESTConfig)
	if err != nil {
		return err
	}
	c.dynamic = dynclient

	serverResources, err := c.ServerResourcesFn(ctx, c.clientset)
	if err != nil {
		return err
	}
	c.serverResources = serverResources

	subjectAccess, err := c.SubjectAccessFn(ctx, c.clientset)
	if err != nil {
		return err
	}
	c.subjectAccess = subjectAccess
	return nil
}

func CheckRestConfig(ctx context.Context, config *rest.Config, logger *zap.Logger) error {
	if config == nil {
		return &errors.NilRESTConfig{}
	} else {
		if config.QPS < 400 {
			logger.Warn("rest.Config QPS below 400",
				// key-value pairs
				zap.String("recommended", ">=400"),
				zap.Float32("qps", config.QPS),
			)
		}
		if config.Burst < 800 {
			logger.Warn("rest.Config Burst below 800",
				// key-value pairs
				zap.String("recommended", ">=800"),
				zap.Int("burst", config.Burst),
			)
		}
		return nil
	}
}

func NewClientset(ctx context.Context, config *rest.Config) (kubernetes.Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, &errors.K8SNewForConfig{Err: err}
	}
	return clientset, nil
}

func NewDynamicClient(ctx context.Context, config *rest.Config) (dynamic.Interface, error) {
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, &errors.K8SNewForConfig{Err: err}
	}
	return dc, nil
}

func NewServerResources(ctx context.Context, clientset kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
	if clientset == nil {
		return nil, fmt.Errorf("nil client.clientset")
	}
	return clientset.Discovery(), nil
}

func NewSubjectAccess(ctx context.Context, clientset kubernetes.Interface) (typedAuthv1.SelfSubjectAccessReviewInterface, error) {
	if clientset == nil {
		return nil, fmt.Errorf("nil client.clientset")
	}
	return clientset.AuthorizationV1().SelfSubjectAccessReviews(), nil
}
