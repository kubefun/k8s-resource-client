package client

import (
	"go.uber.org/zap"
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

func WithRestClientConfig(config *rest.Config) ClientOption {
	return func(c *Client) {
		c.RESTConfig = config
	}
}

func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}
