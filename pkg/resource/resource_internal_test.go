package resource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func TestResourceAccessTypeCast(t *testing.T) {
	failCast := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "unable to type convert status to int") && e.Level == zap.WarnLevel {
			failCast = true
		}
		return nil
	})))
	ra := &resourceAccess{
		logger: logger,
	}

	ra.access.Store("key", 1)

	key := resourceVerbKey(deploymentResource.Key(), "list")
	ra.access.Store(key, "test")
	ra.Allowed(deploymentResource, "list")
	assert.True(t, failCast)

	r := ra.String()
	assert.Equal(t, "key: 1\n", r)
}

func TestResourceAccessStringBadKey(t *testing.T) {
	ra := &resourceAccess{
		logger: zap.NewNop(),
	}

	ra.access.Store(1, 1)
	r := ra.String()
	assert.Equal(t, "", r)
}

var deploymentResource = Resource{
	Namespace:        "default",
	GroupVersionKind: schema.GroupVersionKind{Version: "v1", Group: "apps", Kind: "deployment"},
	APIResource: metav1.APIResource{
		Name:         "deployments",
		SingularName: "deployment",
		Namespaced:   true,
		Group:        "apps",
		Kind:         "deployment",
		Version:      "v1",
		Verbs:        metav1.Verbs{"get", "list", "watch", "delete", "create"},
	},
}
