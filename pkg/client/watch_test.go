package client_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	wtesting "github.com/wwitzel3/k8s-resource-client/pkg/cache/testing"
	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
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

func TestWatchResourceErr(t *testing.T) {
	cache.ResourceWatches = &sync.Map{}

	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
		cache.WithNamespace("not-matching"),
	)
	if err != nil {
		t.Fatal(err)
	}
	watcherFn := func(context.Context, *zap.Logger, dynamic.Interface) (*cache.Watcher, error) {
		return w, nil
	}

	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
		client.WithWatcherFn(watcherFn),
	)
	assert.Nil(t, err)
	wr, err := client.WatchResource(context.TODO(), c, resource.Resource{}, false, []string{"different-ns"})
	assert.EqualError(t, err, "unable to create watch, resource namespace:different-ns does not match watcher namespace:not-matching")
	assert.Nil(t, wr)

}

func TestWatchNamespaceAllErr(t *testing.T) {
	cache.ResourceWatches = &sync.Map{}

	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	watcherFn := func(context.Context, *zap.Logger, dynamic.Interface) (*cache.Watcher, error) {
		return w, nil
	}

	c, err := client.NewClient(context.TODO(),
		client.WithRESTConfig(config),
		client.WithLogger(zap.NewNop()),
		client.WithWatcherFn(watcherFn),
	)
	if err != nil {
		t.Fatal(err)
	}

	cache.ResourceWatches.Store("Version.Kind", "bad-string-should-be-map")

	wr, err := client.WatchResource(context.TODO(), c, resource.Resource{GroupVersionKind: schema.GroupVersionKind{Version: "Version", Kind: "Kind"}}, false, []string{""})
	assert.EqualError(t, err, "append, found key: Version.Kind, unable to cast to []*WatchDetail")
	assert.Nil(t, wr)
	cache.ResourceWatches = &sync.Map{}
}
