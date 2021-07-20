package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
)

func TestWatchIsRunning(t *testing.T) {
	stopCh := make(chan struct{})
	w := &cache.WatchDetails{StopCh: stopCh}
	assert.True(t, w.IsRunning())

	close(stopCh)

	assert.False(t, w.IsRunning())
}
