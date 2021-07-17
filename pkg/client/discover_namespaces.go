package client

import (
	"context"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AutoDiscoverNamespaces makes a best-effort attempt using the dynamic client to list all the namespaces in the cluster
// and update the cache.Namespaces list with the results. This is commonly used as a startup routine.
func AutoDiscoverNamespaces(ctx context.Context, client *Client) error {
	client.Logger.Info("discovering namespaces")

	res := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}

	nri := client.dynamic.Resource(res)

	list, err := nri.List(ctx, metav1.ListOptions{})
	if err != nil {
		return &errors.NamespaceDiscoveryError{Err: err}
	}

	for _, ns := range list.Items {
		cache.Namespaces = append(cache.Namespaces, ns.GetName())
	}

	return nil
}
