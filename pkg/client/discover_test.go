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
	rtesting "github.com/wwitzel3/k8s-resource-client/pkg/resource/testing"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestAutoDiscoverAccess(t *testing.T) {
	assert.Nil(t, cache.Access)

	fakeClient := ctesting.NewFakeClient(nil, false)
	client.AutoDiscoverAccess(context.TODO(), fakeClient, "")

	assert.NotNil(t, cache.Access)
}

func TestUpdateResourceAccess(t *testing.T) {
	cache.Access = nil

	r := resource.Resource{}
	fakeClient := ctesting.NewFakeClient(nil, false)
	err := client.UpdateResourceAccess(context.TODO(), fakeClient, r, []string{""})
	assert.EqualError(t, err, "nil cache.Access")

	client.AutoDiscoverAccess(context.TODO(), fakeClient, "")
	client.UpdateResourceAccess(context.TODO(), fakeClient, r, []string{""})
	err = client.UpdateResourceAccess(context.TODO(), fakeClient, r, []string{""})
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

func TestResourceListForNamespace(t *testing.T) {
	assert.Len(t, cache.Resources.Get("default"), 0)
	ctx := context.TODO()

	config := &rest.Config{QPS: 400, Burst: 800}
	clientset, err := client.NewClientset(ctx, config)

	clientsetFn := func(context.Context, *rest.Config) (kubernetes.Interface, error) {
		return clientset, err
	}

	serverResourcesFn := func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return &rtesting.ServerResourcesFake{}, nil
	}

	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config), client.WithClientsetFn(clientsetFn), client.WithServerResourcesFn(serverResourcesFn))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	resources, err := client.ResourceList(context.TODO(), c, true)
	assert.Nil(t, err)
	assert.Len(t, resources, 1)
}

func TestAutoDiscoverResources(t *testing.T) {
	assert.Len(t, cache.Resources.Get("default"), 0)
	ctx := context.TODO()

	config := &rest.Config{QPS: 400, Burst: 800}
	clientset, err := client.NewClientset(ctx, config)

	clientsetFn := func(context.Context, *rest.Config) (kubernetes.Interface, error) {
		return clientset, err
	}

	serverResourcesFn := func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return &rtesting.ServerResourcesFake{Namespaced: true}, nil
	}

	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config), client.WithClientsetFn(clientsetFn), client.WithServerResourcesFn(serverResourcesFn))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = client.AutoDiscoverResources(context.TODO(), c)
	assert.Nil(t, err)
}

func TestAutoDiscoverResourcesCluster(t *testing.T) {
	assert.Len(t, cache.Resources.Get("default"), 0)
	ctx := context.TODO()

	config := &rest.Config{QPS: 400, Burst: 800}
	clientset, err := client.NewClientset(ctx, config)

	clientsetFn := func(context.Context, *rest.Config) (kubernetes.Interface, error) {
		return clientset, err
	}

	serverResourcesFn := func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return &rtesting.ServerResourcesFake{Namespaced: false}, nil
	}

	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config), client.WithClientsetFn(clientsetFn), client.WithServerResourcesFn(serverResourcesFn))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = client.AutoDiscoverResources(context.TODO(), c)
	assert.Nil(t, err)
}

func TestAutoDiscoverResourcesErr(t *testing.T) {
	assert.Len(t, cache.Resources.Get("default"), 0)
	ctx := context.TODO()

	config := &rest.Config{QPS: 400, Burst: 800}
	clientset, err := client.NewClientset(ctx, config)

	clientsetFn := func(context.Context, *rest.Config) (kubernetes.Interface, error) {
		return clientset, err
	}

	serverResourcesFn := func(context.Context, kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return &rtesting.ServerResourcesFake{Err: true}, nil
	}

	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config), client.WithClientsetFn(clientsetFn), client.WithServerResourcesFn(serverResourcesFn))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = client.AutoDiscoverResources(context.TODO(), c)
	assert.EqualError(t, err, "ResourceDiscoveryError - [get preferred resources: fake server resources error]")
}
