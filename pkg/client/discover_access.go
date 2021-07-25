package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var AutoAccessVerbs = metav1.Verbs{"list", "watch"}

func AutoDiscoverAccess(ctx context.Context, client *Client, namespace string, resources ...resource.Resource) error {
	cache.Access = resource.NewResourceAccess(
		ctx,
		client.subjectAccess,
		namespace,
		resources,
		resource.WithMinimumRBAC(AutoAccessVerbs),
	)
	return nil
}

func UpdateResourceAccess(ctx context.Context, client *Client, res resource.Resource, namespaces []string) error {
	if cache.Access == nil {
		return fmt.Errorf("nil cache.Access")
	}
	for _, ns := range namespaces {
		cache.Access.Update(ctx, client.subjectAccess, ns, res, "list")
		cache.Access.Update(ctx, client.subjectAccess, ns, res, "watch")
	}
	return nil
}
