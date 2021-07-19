package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	ctesting "github.com/wwitzel3/k8s-resource-client/pkg/client/testing"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAutoDiscoverAccess(t *testing.T) {
	assert.Nil(t, cache.Access)

	fakeClient := ctesting.NewFakeClient(nil, false)
	client.AutoDiscoverAccess(context.TODO(), fakeClient)

	assert.NotNil(t, cache.Access)
}

func TestUpdateResourceAccess(t *testing.T) {
	cache.Access = nil

	r := resource.Resource{}
	fakeClient := ctesting.NewFakeClient(nil, false)
	err := client.UpdateResourceAccess(context.TODO(), fakeClient, r)
	assert.EqualError(t, err, "nil cache.Access")

	client.AutoDiscoverAccess(context.TODO(), fakeClient)
	client.UpdateResourceAccess(context.TODO(), fakeClient, r)
	err = client.UpdateResourceAccess(context.TODO(), fakeClient, r)
	assert.Nil(t, err)

}

func TestAutoDiscoverNamespacesErr(t *testing.T) {
	assert.Len(t, cache.Namespaces, 0)

	fakeClient := ctesting.NewFakeClient(nil, true)
	err := client.AutoDiscoverNamespaces(context.TODO(), fakeClient)
	assert.EqualError(t, err, "NamespaceDiscoveryError - fake list error")
}

func TestAutoDiscoverNamespaces(t *testing.T) {
	assert.Len(t, cache.Namespaces, 0)

	v := map[string]interface{}{
		"Name": "default",
	}
	list := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{
			{Object: v},
		},
	}

	for _, item := range list.Items {
		fmt.Println(item)
	}
	fakeClient := ctesting.NewFakeClient(list, false)
	err := client.AutoDiscoverNamespaces(context.TODO(), fakeClient)
	assert.Nil(t, err)

	assert.Len(t, cache.Namespaces, 1)
}

// func TestResourceListForNamespace(t *testing.T) {
// 	assert.Len(t, cache.Resources.GetResources("default"), 0)

// 	fakeClient := ctesting.NewFakeClient(nil, false)
// 	resources, err := client.ResourceListForNamespace(context.TODO(), fakeClient, "default")
// 	assert.Nil(t, err)
// 	assert.Len(t, resources, 1)
// }
