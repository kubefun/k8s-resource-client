package resource_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	rtesting "github.com/wwitzel3/k8s-resource-client/pkg/resource/testing"
)

func TestResourceListNilClient(t *testing.T) {
	_, err := resource.ResourceList(context.TODO(), nil, "")
	assert.EqualError(t, err, "discoveryClient is nil")
}

func TestResourceListClusterScoped(t *testing.T) {

	client := rtesting.ServerResourcesFake{}

	r, err := resource.ResourceList(context.TODO(), client, "")
	assert.Nil(t, err)

	assert.Len(t, r, 1)
	resource := r[0]

	assert.Equal(t, resource.GroupVersionKind, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "deployment"})
	assert.Equal(t, resource.Key(), "apps.v1.deployment")
}

func TestResourceListNamespaceScoped(t *testing.T) {

	client := rtesting.ServerResourcesFake{Namespaced: true}

	r, err := resource.ResourceList(context.TODO(), client, "default")
	assert.Nil(t, err)

	assert.Len(t, r, 1)
	resource := r[0]

	assert.Equal(t, resource.GroupVersionKind, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "deployment"})
	assert.Equal(t, resource.Key(), "default.apps.v1.deployment")
	resource.GroupVersionKind.Group = ""
	assert.Equal(t, resource.Key(), "default.v1.deployment")
}

func TestNewResourceAccessDoneCtx(t *testing.T) {
	ctx, cfn := context.WithCancel(context.TODO())
	cfn()

	resource.NewResourceAccess(ctx, nil, resource.Resource{Namespace: "default"})
}
