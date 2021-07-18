package resource_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
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

func TestResourceAccessChecks(t *testing.T) {
	r := resource.Resource{
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
	authFake := rtesting.SubjectAccessFake{}
	ra := resource.NewResourceAccess(context.TODO(), authFake, []resource.Resource{r},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.False(t, ra.Allowed(r, "list"))
	assert.False(t, ra.AllowedAll(r, []string{"list", "watch"}))
	assert.False(t, ra.AllowedAny(r, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 3")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 3")
}
