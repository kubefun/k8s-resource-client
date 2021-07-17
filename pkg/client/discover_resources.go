package client

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
	"github.com/wwitzel3/subjectaccess/pkg/subjectaccess"
)

// AutoDiscoverResources makes a best-effort attempt using the discover client to list all the resources for all of the namespaces
// that were provided and update the cache.ResourceMap. This operation is expensive on large clusters and should be considered part
// of a startup routine and a long-duration periodic task.
func AutoDiscoverResources(ctx context.Context, client *Client, namespaces ...string) error {
	wg := sync.WaitGroup{}

	discoveryErr := &errors.ResourceDiscoveryError{}

	for _, n := range namespaces {
		wg.Add(1)
		namespace := n
		go func() {
			defer wg.Done()
			client.Logger.Info("discovering resources",
				zap.String("namespace", namespace),
			)
			resources, err := ResourceListForNamespace(ctx, client, namespace)
			if err != nil {
				discoveryErr.Add(err)
				return
			}
			cache.Resources.AddResources(namespace, resources...)
		}()
	}
	wg.Wait()
	if len(discoveryErr.Err) > 0 {
		return discoveryErr
	}
	return nil
}

// ResourceListForNamespace uses a Discovery Client and attempts to list all of the known resources for the given namespace.
// This method can be used to populate initial resource lists as well as refresh existing caches.
func ResourceListForNamespace(ctx context.Context, client *Client, namespace string) ([]subjectaccess.Resource, error) {
	scopedResources, err := subjectaccess.ResourceList(ctx, client.clientset.DiscoveryClient, namespace)
	if err != nil {
		// TODO: consider if we want a typed-error
		return nil, err
	}
	return scopedResources, nil
}
