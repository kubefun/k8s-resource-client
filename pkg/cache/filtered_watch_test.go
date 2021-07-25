package cache_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	wtesting "github.com/wwitzel3/k8s-resource-client/pkg/cache/testing"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	rtesting "k8s.io/apimachinery/pkg/runtime/testing"
	"k8s.io/client-go/util/workqueue"
)

func TestFilteredWatchDetail(t *testing.T) {
	cache.ResourceWatches = &sync.Map{}

	dsifFake := wtesting.NewFakeDynamicSharedInformerFactory()
	dynFake := ctesting.FakeDynamicClient{}

	dsifFake.GenericInformer.GenericLister.NamespaceLister.Objects = []runtime.Object{nil}
	dsifFake.GenericInformer.GenericLister.NamespaceLister.Object = &rtesting.MockCacheableObject{}

	filteredInfo := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		println(e.Message)
		if strings.Contains(e.Message, "found NamespaceAll creating filtered watch detail") && e.Level == zap.InfoLevel {
			filteredInfo = true
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

	podWd, err := w.Watch(context.TODO(), "", podResource, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.NotNil(t, podWd)

	lister, err := cache.WatchForResource(podResource, "testing-ns")
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, "testing-ns", lister.Namespace())
	assert.Equal(t, ".v1.Pod", lister.Key())

	objs, err := lister.List(labels.Everything())
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Len(t, objs, 1)

	obj, err := lister.Get("test-name")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.NotNil(t, obj)
	assert.True(t, filteredInfo)

	logging.Logger, _ = zap.NewProduction()
}

func TestFilteredWatchDetailDrain(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w1 := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.True(t, w1.IsRunning())

	filtered := &cache.FilteredWatchDetail{Detail: w1}
	filtered.Drain(eventCh, stopCh)

	i := "string1"
	w1.Queue.Add(i)

	s := <-eventCh
	assert.Equal(t, "string1", s.(string))

	filtered.Stop()
	filtered.Drain(eventCh, stopCh)
	assert.False(t, w1.IsRunning())
}
