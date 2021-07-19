package testing

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/wwitzel3/k8s-resource-client/pkg/client"
)

func NewFakeClient(listResults *unstructured.UnstructuredList, listErr bool) *client.Client {
	client, err := client.NewClient(
		context.TODO(),
		client.WithRESTConfig(FakeConfig),
		client.WithClientsetFn(FakeClientset),
		client.WithDynamicClientFn(FakeDynamicFactory(listResults, listErr)),
	)
	if err != nil {
		panic(err)
	}
	return client
}

func FakeClientset(ctx context.Context, config *rest.Config) (kubernetes.Interface, error) {
	cs := &kubernetes.Clientset{}
	return cs, nil
}

func FakeDynamicFactory(listResults *unstructured.UnstructuredList, listErr bool) func(context.Context, *rest.Config) (dynamic.Interface, error) {
	fakeDynamic := func(ctx context.Context, config *rest.Config) (dynamic.Interface, error) {
		return &FakeDynamicClient{ListErr: listErr, ListResults: listResults}, nil
	}
	return fakeDynamic
}

var FakeConfig = &rest.Config{QPS: 400, Burst: 800}

type FakeDynamicClient struct {
	ListErr     bool
	ListResults *unstructured.UnstructuredList
}

func (d FakeDynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &NamespaceableResource{ListErr: d.ListErr, ListResults: d.ListResults}
}

type NamespaceableResource struct {
	ListErr     bool
	ListResults *unstructured.UnstructuredList
}

func (n NamespaceableResource) Namespace(string) dynamic.ResourceInterface {
	return nil
}

func (n NamespaceableResource) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	if n.ListErr {
		return nil, fmt.Errorf("fake list error")
	}
	if n.ListResults == nil {
		panic("nil list results")
	}
	return n.ListResults, nil
}
func (n NamespaceableResource) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (n NamespaceableResource) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (n NamespaceableResource) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (n NamespaceableResource) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}
func (n NamespaceableResource) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}
func (n NamespaceableResource) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
func (n NamespaceableResource) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (n NamespaceableResource) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}
