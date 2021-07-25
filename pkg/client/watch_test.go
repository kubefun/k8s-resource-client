package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

func TestWatchResource(t *testing.T) {
	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	w, err := client.WatchResource(context.TODO(), c, resource.Resource{}, false, []string{""})
	assert.Nil(t, err)
	assert.NotNil(t, w)

	w, err = client.WatchResource(context.TODO(), c, resource.Resource{}, false, []string{"default"})
	assert.Nil(t, err)
	assert.NotNil(t, w)
}

func TestWatchAllResource(t *testing.T) {
	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	client.WatchAllResources(context.TODO(), c, false, []string{""})
}
