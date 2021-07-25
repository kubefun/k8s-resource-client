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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	rtesting "k8s.io/apimachinery/pkg/runtime/testing"
	kcache "k8s.io/client-go/tools/cache"

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

	_, err = w.Watch(context.TODO(), "test", resource.Resource{}, false)
	assert.EqualError(t, err, "unable to create watch, resource namespace:test does not match watcher namespace:default")
}

func TestNewWatcherDefaultInformerFactory(t *testing.T) {
	dynFake := ctesting.FakeDynamicClient{}

	_, err := cache.NewWatcher(context.TODO(), cache.WithDynamicClient(dynFake))
	assert.Nil(t, err)
}

func TestResourceWatchesBadKey(t *testing.T) {
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

	cache.ResourceWatches.Store("Version.Kind", "bad-string-should-be-map")
	_, err = cache.WatchForResource(resource.Resource{GroupVersionKind: schema.GroupVersionKind{Version: "Version", Kind: "Kind"}})
	assert.EqualError(t, err, "watch, found key:Version.Kind, unable to cast to *sync.Map")

	_, err = w.Watch(context.TODO(), "", resource.Resource{GroupVersionKind: schema.GroupVersionKind{Version: "Version", Kind: "Kind"}}, false)
	assert.EqualError(t, err, "append, found key: Version.Kind, unable to cast to []*WatchDetail")
}

func TestResourceWatchesAddReuse(t *testing.T) {
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

	// Add new watcher to cache
	cache.ResourceWatches.Store("Version.Kind", &sync.Map{})
	w1, err := w.Watch(context.TODO(), "", resource.Resource{GroupVersionKind: schema.GroupVersionKind{Version: "Version", Kind: "Kind"}}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Use already created Watcher
	w2, err := w.Watch(context.TODO(), "", resource.Resource{GroupVersionKind: schema.GroupVersionKind{Version: "Version", Kind: "Kind"}}, false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, w1.Key(), w2.Key())
}

func TestWatcherNamespaceAll(t *testing.T) {
	cache.ResourceWatches = &sync.Map{}

	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	dsifFake.GenericInformer.GenericLister.Objects = []runtime.Object{nil}
	dsifFake.GenericInformer.GenericLister.Object = &rtesting.MockCacheableObject{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd, err := w.Watch(context.TODO(), "", deploymentResource, false)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	v, err := cache.WatchForResource(deploymentResource)
	assert.Nil(t, err)
	assert.NotNil(t, v)

	objs, err := v.List(labels.Everything())
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Len(t, objs, 1)
	obj, err := v.Get(labels.Everything().String())
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.NotNil(t, obj)
}

func TestWatcherAppendMulti(t *testing.T) {
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

	wd, err := w.Watch(context.TODO(), "default", deploymentResource, false)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	wd, err = w.Watch(context.TODO(), "different-ns", deploymentResource, false)
	assert.Nil(t, err)
	assert.NotNil(t, wd)
}

func TestWatcherQueueEvents(t *testing.T) {
	cache.ResourceWatches = &sync.Map{}

	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(&dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd, err := w.Watch(context.TODO(), "default", deploymentResource, true)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	assert.Len(t, dsifFake.GenericInformer.SharedIndexInformer.Handlers, 1)
	handler := dsifFake.GenericInformer.SharedIndexInformer.Handlers[0]

	handler.OnAdd("")
	handler.OnDelete("")
	handler.OnUpdate("", "")
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

	wd1, err := w.Watch(context.TODO(), "foo", resource.Resource{}, false)
	assert.Nil(t, err)
	wd2, err := w.Watch(context.TODO(), "bar", resource.Resource{}, false)
	assert.Nil(t, err)

	cache.WatcherStop()

	assert.Equal(t, wd1.IsRunning(), 0)
	assert.Equal(t, wd2.IsRunning(), 0)
}

func TestWatcherHelpers(t *testing.T) {
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

	wd, err := w.Watch(context.TODO(), "default", deploymentResource, false)
	assert.Nil(t, err)
	assert.NotNil(t, wd)

	v, err := cache.WatchForResource(deploymentResource)
	assert.Nil(t, err)
	assert.NotNil(t, v)

	_, err = cache.WatchForResource(resource.Resource{})
	assert.EqualError(t, err, "no watch found for resource: {GroupVersionKind:/, Kind= APIResource:{Name: SingularName: Namespaced:false Group: Version: Kind: Verbs:[] ShortNames:[] Categories:[] StorageVersionHash:}}")

	podWatcher, err := w.Watch(context.TODO(), "default", podResource, false)
	assert.Nil(t, err)

	v, err = cache.WatchForResource(podResource, "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.NotNil(t, v)

	v, err = cache.WatchForResource(podResource, "default", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.NotNil(t, v)

	watchers := cache.WatchList(false)
	assert.Len(t, watchers, 2)
	assert.Equal(t, cache.WatchCount(false), 2)

	podWatcher.Stop()
	watchers = cache.WatchList(true)
	assert.Len(t, watchers, 1)
	assert.Equal(t, cache.WatchCount(true), 1)
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
	cache.ResourceWatches = &sync.Map{}
	cache.ResourceWatches.Store("test", "test")
	assert.Equal(t, 0, cache.WatchCount(false))
	assert.Len(t, cache.WatchList(false), 0)
	cache.WatcherStop()

	badWatchMap := &sync.Map{}
	badWatchMap.Store("key", "value")
	cache.ResourceWatches.Store("test", badWatchMap)

	assert.Equal(t, 0, cache.WatchCount(false))
	assert.Len(t, cache.WatchList(false), 0)
	cache.WatcherStop()
	cache.ResourceWatches = &sync.Map{}
}

var deploymentResource = resource.Resource{
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
