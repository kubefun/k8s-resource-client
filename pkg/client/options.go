package client

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedAuthv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

type ClientOption func(*Client)

func WithResourceMode(mode ModeType) ClientOption {
	return func(c *Client) {
		c.ResourceMode = mode
	}
}

func WithNamespaceMode(mode ModeType) ClientOption {
	return func(c *Client) {
		c.NamespaceMode = mode
	}
}

func WithSkipSubjectAccessChecks(skip bool) ClientOption {
	return func(c *Client) {
		c.SkipSubjectAccessChecks = skip
	}
}

func WithRESTConfig(config *rest.Config) ClientOption {
	return func(c *Client) {
		c.RESTConfig = config
	}
}

func WithClientsetFn(fn func(context.Context, *rest.Config) (kubernetes.Interface, error)) ClientOption {
	return func(c *Client) {
		c.ClientsetFn = fn
	}
}

func WithDynamicClientFn(fn func(context.Context, *rest.Config) (dynamic.Interface, error)) ClientOption {
	return func(c *Client) {
		c.DynamicClientFn = fn
	}
}

func WithServerResourcesFn(fn func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error)) ClientOption {
	return func(c *Client) {
		c.ServerResourcesFn = fn
	}
}

func WithSubjectAccessFn(fn func(context.Context, kubernetes.Interface) (typedAuthv1.SelfSubjectAccessReviewInterface, error)) ClientOption {
	return func(c *Client) {
		c.SubjectAccessFn = fn
	}
}

func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}
