package cache_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	wtesting "github.com/wwitzel3/k8s-resource-client/pkg/cache/testing"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

func TestNewWatcherErr(t *testing.T) {
	_, err := cache.NewWatcher(context.TODO())
	assert.EqualError(t, err, "dynamic client nil, use WithDynamicClient option")
}

func TestWatchNamespaceMismatch(t *testing.T) {
	dynFake := ctesting.FakeDynamicClient{}
	w, err := cache.NewWatcher(context.TODO(), cache.WithDynamicClient(dynFake), cache.WithNamespace("default"))
	assert.Nil(t, err)

	_, err = w.Watch(context.TODO(), resource.Resource{Namespace: "test"}, false)
	assert.EqualError(t, err, "unable to create watch, resource namespace:test does not match watcher namespace:default")
}

func TestNewWatcherDefaultInformerFactory(t *testing.T) {
	dynFake := ctesting.FakeDynamicClient{}

	_, err := cache.NewWatcher(context.TODO(), cache.WithDynamicClient(dynFake))
	assert.Nil(t, err)
}

func TestWatcherQueueEvents(t *testing.T) {
	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(&dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd, err := w.Watch(context.TODO(), deploymentResource, true)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	assert.Len(t, dsifFake.GenericInformer.SharedIndexInformer.Handlers, 1)
	handler := dsifFake.GenericInformer.SharedIndexInformer.Handlers[0]

	handler.OnAdd("")
	handler.OnDelete("")
	handler.OnUpdate("", "")
}

func TestWatcherHelpers(t *testing.T) {
	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd, err := w.Watch(context.TODO(), deploymentResource, false)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	v, ok := cache.WatchForResource(deploymentResource)
	assert.True(t, ok)
	assert.NotNil(t, v)

	_, ok = cache.WatchForResource(resource.Resource{})
	assert.False(t, ok)

	podWatcher, err := w.Watch(context.TODO(), podResource, false)
	assert.Nil(t, err)
	watchers := cache.WatchList(false)
	assert.Len(t, watchers, 2)
	assert.Equal(t, cache.WatchCount(false), 2)

	podWatcher.Stop()
	watchers = cache.WatchList(true)
	assert.Len(t, watchers, 1)
	assert.Equal(t, cache.WatchCount(true), 1)
}

func TestWatchIsRunning(t *testing.T) {
	stopCh := make(chan struct{})
	w := &cache.WatchDetail{StopCh: stopCh}
	assert.True(t, w.IsRunning())

	close(stopCh)

	assert.False(t, w.IsRunning())
}

func TestWatchStopAll(t *testing.T) {
	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd1, err := w.Watch(context.TODO(), resource.Resource{Namespace: "foo"}, false)
	assert.Nil(t, err)
	wd2, err := w.Watch(context.TODO(), resource.Resource{Namespace: "bar"}, false)
	assert.Nil(t, err)

	cache.WatcherStop()

	assert.False(t, wd1.IsRunning())
	assert.False(t, wd2.IsRunning())
}

func TestWatchDrainStopMain(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w.IsRunning())

	w.Drain(eventCh, stopCh)

	i := "string1"
	w.Queue.Add(i)
	s := <-eventCh
	assert.Equal(t, "string1", s.(string))

	w.Stop()
	w.Drain(eventCh, stopCh)
}

func TestWatchDrainStopLocal(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w.IsRunning())

	w.Drain(eventCh, stopCh)

	i := "string1"
	w.Queue.Add(i)
	s := <-eventCh
	assert.Equal(t, "string1", s.(string))

	close(stopCh)
	w.Drain(eventCh, stopCh)
}

func TestWatchDrainShutdown(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w.IsRunning())

	i := "shutdown"
	w.Queue.Add(i)
	w.Queue.ShutDown()

	w.Drain(eventCh, stopCh)
	s := <-eventCh
	assert.Equal(t, "shutdown", s.(string))
}

func TestWatchErrorHandlerFactory(t *testing.T) {
	type test struct {
		err error
		fn  func(*kcache.Reflector, error)
	}
	tests := []test{
		{
			err: &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonExpired}},
			fn:  cache.WatchErrorHandlerFactory(zap.NewNop(), "", make(chan struct{})),
		},
		{
			err: &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonGone}},
			fn:  cache.WatchErrorHandlerFactory(zap.NewNop(), "", make(chan struct{})),
		},
		{
			err: io.EOF,
			fn:  cache.WatchErrorHandlerFactory(zap.NewNop(), "", make(chan struct{})),
		},
		{
			err: io.ErrUnexpectedEOF,
			fn:  cache.WatchErrorHandlerFactory(zap.NewNop(), "", make(chan struct{})),
		},
		{
			err: nil,
			fn:  cache.WatchErrorHandlerFactory(zap.NewNop(), "", make(chan struct{})),
		},
	}

	for _, tc := range tests {
		tc.fn(nil, tc.err)
	}
}

func TestWatcherHelpersBad(t *testing.T) {
	cache.Watches = &sync.Map{}
	cache.Watches.Store("test", "test")
	assert.Equal(t, 0, cache.WatchCount(false))
	assert.Len(t, cache.WatchList(false), 0)
	cache.WatcherStop()
	cache.Watches = &sync.Map{}
}

var deploymentResource = resource.Resource{
	Namespace:        "default",
	GroupVersionKind: schema.GroupVersionKind{Version: "v1", Group: "apps", Kind: "Deployment"},
	APIResource: metav1.APIResource{
		Name:         "deployments",
		SingularName: "deployment",
		Namespaced:   true,
		Group:        "apps",
		Kind:         "Deployment",
		Version:      "v1",
		Verbs:        metav1.Verbs{"get", "list", "watch", "delete", "create"},
	},
}

var podResource = resource.Resource{
	Namespace:        "default",
	GroupVersionKind: schema.GroupVersionKind{Version: "v1", Group: "", Kind: "Pod"},
	APIResource: metav1.APIResource{
		Name:         "pods",
		SingularName: "pod",
		Namespaced:   true,
		Group:        "",
		Kind:         "Pod",
		Version:      "v1",
		Verbs:        metav1.Verbs{"get", "list", "watch", "delete", "create"},
	},
}
