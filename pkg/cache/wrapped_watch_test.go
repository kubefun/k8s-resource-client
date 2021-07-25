package cache_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	rtesting "k8s.io/apimachinery/pkg/runtime/testing"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	wtesting "github.com/wwitzel3/k8s-resource-client/pkg/cache/testing"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
)

func TestWrappedWatchDetails(t *testing.T) {
	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	listErr := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		println(e.Message)
		if strings.Contains(e.Message, "failed to list") && e.Level == zap.ErrorLevel {
			listErr = true
		}
		return nil
	})))
	logging.Logger = logger

	w, err := cache.NewWatcher(context.TODO(),
		cache.WithDynamicClient(&dynFake),
		cache.WithDynamicSharedInformerFactory(dsifFake),
		cache.WithLogger(zap.NewNop()),
	)
	assert.Nil(t, err)

	podWd, err := w.Watch(context.TODO(), "default", podResource, false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	deployWd, err := w.Watch(context.TODO(), "different-ns", deploymentResource, false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	wrapped := &cache.WrappedWatchDetails{Listers: []cache.ResourceLister{podWd, deployWd}}
	dsifFake.GenericInformer.GenericLister.NamespaceLister.Objects = []runtime.Object{nil}
	dsifFake.GenericInformer.GenericLister.NamespaceLister.Object = &rtesting.MockCacheableObject{}

	key := wrapped.Key()
	assert.Equal(t, "default.v1.Pod,different-ns.apps.v1.Deployment", key)

	namespace := wrapped.Namespace()
	assert.Equal(t, "default,different-ns", namespace)

	objs, err := wrapped.List(labels.Everything())
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Len(t, objs, 2)

	_, err = wrapped.Get("test-obj")
	assert.Nil(t, err)

	dsifFake.GenericInformer.GenericLister.NamespaceLister.ListErr = fmt.Errorf("test lister error")
	dsifFake.GenericInformer.GenericLister.NamespaceLister.GetErr = fmt.Errorf("test get error")
	objs, err = wrapped.List(labels.Everything())
	assert.EqualError(t, err, "test lister error,test lister error")
	assert.Len(t, objs, 0)
	// set true by zap.Logger
	assert.True(t, listErr)

	_, err = wrapped.Get("test-obj")
	assert.EqualError(t, err, "unable to find object test-obj in any namespace of: [default different-ns]")

	assert.Equal(t, podWd.IsRunning(), 1)
	assert.Equal(t, deployWd.IsRunning(), 1)
	wrapped.Stop()
	assert.Equal(t, podWd.IsRunning(), 0)
	assert.Equal(t, deployWd.IsRunning(), 0)

	logging.Logger, _ = zap.NewProduction()
}

func TestWrappedWatchDrainStopMain(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w1 := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.Equal(t, w1.IsRunning(), 1)

	w2 := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.Equal(t, w2.IsRunning(), 1)

	wrapped := &cache.WrappedWatchDetails{Listers: []cache.ResourceLister{w1, w2}}
	wrapped.Drain(eventCh, stopCh)

	i := "string1"
	w1.Queue.Add(i)

	j := "string2"
	w2.Queue.Add(j)

	values := []string{}
	s := <-eventCh
	values = append(values, s.(string))

	s = <-eventCh
	values = append(values, s.(string))

	assert.Contains(t, values, "string1")
	assert.Contains(t, values, "string2")

	wrapped.Stop()
	wrapped.Drain(eventCh, stopCh)
}
