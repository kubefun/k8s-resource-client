package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var AutoAccessVerbs = metav1.Verbs{"list", "watch"}

func AutoDiscoverAccess(ctx context.Context, client *Client, resources ...resource.Resource) error {
	cache.Access = resource.NewResourceAccess(
		ctx,
		client.subjectAccess,
		resources,
		resource.WithMinimumRBAC(AutoAccessVerbs),
	)
	return nil
}

func UpdateResourceAccess(ctx context.Context, client *Client, resource resource.Resource) error {
	if cache.Access == nil {
		return fmt.Errorf("nil cache.Access")
	}
	cache.Access.Update(ctx, client.subjectAccess, resource, "list")
	cache.Access.Update(ctx, client.subjectAccess, resource, "watch")
	return nil
}
