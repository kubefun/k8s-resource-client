package resource_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rtesting "github.com/wwitzel3/k8s-resource-client/pkg/resource/testing"
)

func TestNewResourceAccessDoneCtx(t *testing.T) {
	ctx, cfn := context.WithCancel(context.TODO())
	cfn()

	resource.NewResourceAccess(ctx, nil, []resource.Resource{{Namespace: "default"}})
}

func TestNewResourceAccessOptions(t *testing.T) {
	ctx, cfn := context.WithCancel(context.TODO())
	cfn()

	resource.NewResourceAccess(ctx, nil, []resource.Resource{{Namespace: "default"}},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
}

func TestResourceAccessChecksFalse(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}
	ra := resource.NewResourceAccess(context.TODO(), authFake, []resource.Resource{deploymentResource},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.False(t, ra.Allowed(deploymentResource, "list"))
	assert.False(t, ra.AllowedAll(deploymentResource, []string{"list", "watch"}))
	assert.False(t, ra.AllowedAny(deploymentResource, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 3")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 3")
}

func TestResourceAccessChecksTrue(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}
	authFake.CreateFn = func(fake *rtesting.SubjectAccessFake) (*v1.SelfSubjectAccessReview, error) {
		ssar := &v1.SelfSubjectAccessReview{
			Status: v1.SubjectAccessReviewStatus{
				Allowed: true,
			},
		}
		return ssar, nil
	}

	ra := resource.NewResourceAccess(context.TODO(), authFake, []resource.Resource{deploymentResource},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.True(t, ra.Allowed(deploymentResource, "list"))
	assert.True(t, ra.AllowedAll(deploymentResource, []string{"list", "watch"}))
	assert.True(t, ra.AllowedAny(deploymentResource, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 1")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 1")
}

var deploymentResource = resource.Resource{
	Namespace:        "default",
	GroupVersionKind: schema.GroupVersionKind{Version: "v1", Group: "apps", Kind: "deployment"},
	APIResource: metav1.APIResource{
		Name:         "deployments",
		SingularName: "deployment",
		Namespaced:   true,
		Group:        "apps",
		Kind:         "deployment",
		Version:      "v1",
		Verbs:        metav1.Verbs{"get", "list", "watch", "delete", "create"},
	},
}
