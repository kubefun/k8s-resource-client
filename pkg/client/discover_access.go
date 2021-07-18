package client

import (
	"context"
	"fmt"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

func AutoDiscoverAccess(ctx context.Context, client *Client, resources ...resource.Resource) error {
	cache.Access = resource.NewResourceAccess(ctx, client.clientset.AuthorizationV1().SelfSubjectAccessReviews(), resources)
	return nil
}

func UpdateResourceAccess(ctx context.Context, client *Client, resource resource.Resource) error {
	if cache.Access == nil {
		return fmt.Errorf("nil cache.Access")
	}
	cache.Access.Update(ctx, client.clientset.AuthorizationV1().SelfSubjectAccessReviews(), resource, "list")
	cache.Access.Update(ctx, client.clientset.AuthorizationV1().SelfSubjectAccessReviews(), resource, "watch")
	return nil
}
