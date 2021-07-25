package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
)

func TestWatchIsRunning(t *testing.T) {
	stopCh := make(chan struct{})
	w := &cache.WatchDetail{StopCh: stopCh}
	assert.Equal(t, w.IsRunning(), 1)

	close(stopCh)

	assert.Equal(t, w.IsRunning(), 0)
}

func TestWatchDrainStopMain(t *testing.T) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	w := &cache.WatchDetail{StopCh: make(chan struct{}), Queue: workqueue.New(), Logger: zap.NewNop()}
	assert.Equal(t, w.IsRunning(), 1)

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
	assert.Equal(t, w.IsRunning(), 1)

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
	assert.Equal(t, w.IsRunning(), 1)

	i := "shutdown"
	w.Queue.Add(i)
	w.Queue.ShutDown()

	w.Drain(eventCh, stopCh)
	s := <-eventCh
	assert.Equal(t, "shutdown", s.(string))
}
