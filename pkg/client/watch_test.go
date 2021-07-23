package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
)

func TestWatchResource(t *testing.T) {
	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	w, err := client.WatchResource(context.TODO(), c, resource.Resource{}, "", false)
	assert.Nil(t, err)
	assert.NotNil(t, w)
}

func TestWatchAllResource(t *testing.T) {
	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	client.WatchAllResources(context.TODO(), c, "", false)
}
