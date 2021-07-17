package client

import (
	"go.uber.org/zap"
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
}

func NewClient(options ...ClientOption) (*Client, error) {
	defer logging.Logger.Sync()

	c := &Client{
		ResourceMode:            Auto,
		NamespaceMode:           Auto,
		SkipSubjectAccessChecks: false,
		Logger:                  logging.Logger,
	}

	for _, opt := range options {
		opt(c)
	}

	if c.RESTConfig == nil {
		return nil, &errors.NilRESTConfig{}
	} else {
		if c.RESTConfig.QPS < 400 {
			c.Logger.Warn("config QPS below 400",
				// key-value pairs
				zap.String("recommended", ">=400"),
				zap.Float32("qps", c.RESTConfig.QPS),
			)
		}
		if c.RESTConfig.Burst < 800 {
			c.Logger.Warn("config Burst below 800",
				// key-value pairs
				zap.String("recommended", ">=800"),
				zap.Int("burst", c.RESTConfig.Burst),
			)
		}
	}

	return c, nil
}
