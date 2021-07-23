package client

import (
	"context"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
)

func WatchResource(ctx context.Context, client *Client, res resource.Resource, namespace string, queueEvents bool) (*cache.WatchDetail, error) {
	client.Logger.Info("creating ListWatch",
		zap.String("resource", res.APIResource.Name),
		zap.String("namespace", namespace),
	)
	res.Namespace = namespace
	return client.watcher.Watch(ctx, res, queueEvents)
}

func WatchAllResources(ctx context.Context, client *Client, namespace string, queueEvents bool) {
	for _, res := range cache.Resources.Get("namespace") {
		WatchResource(ctx, client, res, namespace, queueEvents)
	}
}
