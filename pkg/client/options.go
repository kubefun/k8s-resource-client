package client

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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

func WithClientsetFn(fn func(context.Context, *rest.Config) (*kubernetes.Clientset, error)) ClientOption {
	return func(c *Client) {
		c.ClientsetFn = fn
	}
}

func WithDynamicClientFn(fn func(context.Context, *rest.Config) (dynamic.Interface, error)) ClientOption {
	return func(c *Client) {
		c.DynamicClientFn = fn
	}
}

func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}
