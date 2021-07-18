package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusIntAsBool(t *testing.T) {
	b := statusIntAsBool(Denied)
	assert.False(t, b)

	b = statusIntAsBool(Unused)
	assert.False(t, b)

	b = statusIntAsBool(Allowed)
	assert.True(t, b)

	b = statusIntAsBool(Error)
	assert.False(t, b)
}

func TestResourceVerbKey(t *testing.T) {
	verbKey := resourceVerbKey("pod", "watch")
	assert.Equal(t, "pod.watch", verbKey)
}
