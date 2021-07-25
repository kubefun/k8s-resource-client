package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewResourceCache(t *testing.T) {
	c := cache.NewResourceCache()
	c.Add("default", testResource, testResource)

	resources := c.Get("default")
	assert.Len(t, resources, 1)

	c.Add("default", testResource)
	assert.Len(t, resources, 1)

	empty := c.Get("not-found")
	assert.Len(t, empty, 0)
}

var testResource = resource.Resource{
	GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "deployment"},
}
