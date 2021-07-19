package client

import (
	"context"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

// AutoDiscoverResources makes a best-effort attempt using the discover client to list all the resources for all of the namespaces
// that were provided and update the cache.ResourceMap. This operation is expensive on large clusters and should be considered part
// of a startup routine and a long-duration periodic task.
func AutoDiscoverResources(ctx context.Context, client *Client) error {
	client.Logger.Info("discovering resources")
	resources, err := ResourceListForNamespace(ctx, client, "")
	if err != nil {
		return &errors.ResourceDiscoveryError{Err: []error{err}}
	}
	for _, resource := range resources {
		if resource.APIResource.Namespaced {
			cache.Resources.AddResources("namespace", resource)
		} else {
			cache.Resources.AddResources("cluster", resource)
		}
	}

	return nil
}

// ResourceListForNamespace uses a Discovery Client and attempts to list all of the known resources for the given namespace.
// This method can be used to populate initial resource lists as well as refresh existing caches.
func ResourceListForNamespace(ctx context.Context, client *Client, namespace string) ([]resource.Resource, error) {
	scopedResources, err := resource.ResourceList(ctx, client.Logger, client.clientset.Discovery(), namespace)
	if err != nil {
		// TODO: consider if we want a typed-error
		return nil, err
	}
	return scopedResources, nil
}
