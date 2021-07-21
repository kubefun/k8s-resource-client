package cache_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	wtesting "github.com/wwitzel3/k8s-resource-client/pkg/cache/testing"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

func TestWatcherHelpers(t *testing.T) {
	dsifFake := wtesting.FakeDynamicSharedInformerFactory{}
	dynFake := ctesting.FakeDynamicClient{}

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
	)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	wd := w.Watch(context.TODO(), deploymentResource, false)
	assert.NotNil(t, wd)

	v, ok := cache.WatcherForResource(deploymentResource)
	assert.True(t, ok)
	assert.NotNil(t, v)

	_, ok = cache.WatcherForResource(resource.Resource{})
	assert.False(t, ok)

	podWatcher := w.Watch(context.TODO(), podResource, false)
	watchers := cache.WatcherList(false)
	assert.Len(t, watchers, 2)
	assert.Equal(t, cache.WatchCount(false), 2)

	podWatcher.Stop()
	watchers = cache.WatcherList(true)
	assert.Len(t, watchers, 1)
	assert.Equal(t, cache.WatchCount(true), 1)
}

func TestWatchIsRunning(t *testing.T) {
	stopCh := make(chan struct{})
	w := &cache.WatchDetails{StopCh: stopCh}
	assert.True(t, w.IsRunning())

	close(stopCh)

	assert.False(t, w.IsRunning())
}

func TestWatchDrain(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetails{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w.IsRunning())

	w.Drain(eventCh, stopCh)

	i := "string1"
	w.Queue.Add(i)
	s := <-eventCh
	assert.Equal(t, "string1", s.(string))

	close(stopCh)
	w.Drain(eventCh, stopCh)

	w.Stop()
	w.Drain(eventCh, stopCh)
}

func TestWatchDrainShutdown(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetails{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w.IsRunning())

	i := "shutdown"
	w.Queue.Add(i)
	w.Queue.ShutDown()

	w.Drain(eventCh, stopCh)
	s := <-eventCh
	assert.Equal(t, "shutdown", s.(string))
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
