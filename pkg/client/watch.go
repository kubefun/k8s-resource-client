package client

import (
	"context"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
)

// WatchResource creates a watch for the Resource in the provided namespaces.
// To watch across all namespaces you can pass in metav1.NamespaceAll.
func WatchResource(ctx context.Context, client *Client, res resource.Resource, queueEvents bool, namespaces []string) ([]*cache.WatchDetail, error) {
	if hasNamespaceAll(namespaces) {
		w, err := client.watcher.Watch(ctx, "", res, queueEvents)
		if err != nil {
			return nil, err
		}
		return []*cache.WatchDetail{w}, nil
	}

	watchDetails := make([]*cache.WatchDetail, len(namespaces))
	for i, ns := range namespaces {
		client.Logger.Info("creating watch",
			zap.String("resource", res.Key()),
			zap.String("namespace", ns),
		)
		w, err := client.watcher.Watch(ctx, ns, res, queueEvents)
		if err != nil {
			return nil, err
		}
		watchDetails[i] = w
	}
	return watchDetails, nil
}

func WatchAllResources(ctx context.Context, client *Client, queueEvents bool, namespaces []string) {
	for _, res := range cache.Resources.Get("namespace") {
		WatchResource(ctx, client, res, queueEvents, namespaces)
	}
}

func hasNamespaceAll(namespaces []string) bool {
	for _, ns := range namespaces {
		if ns == "" {
			return true
		}
	}
	return false
}
