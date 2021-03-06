package resource_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rtesting "github.com/wwitzel3/k8s-resource-client/pkg/resource/testing"
)

func TestNewResourceAccessDoneCtx(t *testing.T) {
	ctx, cfn := context.WithCancel(context.TODO())
	cfn()

	resource.NewResourceAccess(ctx, nil, "default", []resource.Resource{})
}

func TestNewResourceAccessOptions(t *testing.T) {
	ctx, cfn := context.WithCancel(context.TODO())
	cfn()

	resource.NewResourceAccess(ctx, nil, "default", []resource.Resource{},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
}

func TestResourceAccessChecksFalse(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}
	ra := resource.NewResourceAccess(context.TODO(), authFake, "default", []resource.Resource{deploymentResource},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.False(t, ra.Allowed("default", deploymentResource, "list"))
	assert.False(t, ra.AllowedAll("default", deploymentResource, []string{"list", "watch"}))
	assert.False(t, ra.AllowedAny("default", deploymentResource, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 3")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 3")
}

func TestResourceAccessChecksNotFound(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}

	notFound := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "not found") && e.Level == zap.DebugLevel {
			notFound = true
		}
		return nil
	})))

	ra := resource.NewResourceAccess(context.TODO(), authFake, "default", []resource.Resource{},
		resource.WithLogger(logger),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	ra.Allowed("default", deploymentResource, "list")
	assert.True(t, notFound)
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

	ra := resource.NewResourceAccess(context.TODO(), authFake, "default", []resource.Resource{deploymentResource},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.True(t, ra.Allowed("default", deploymentResource, "list"))
	assert.True(t, ra.AllowedAll("default", deploymentResource, []string{"list", "watch"}))
	assert.True(t, ra.AllowedAny("default", deploymentResource, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 1")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 1")
}

func TestResourceAccessContextClosed(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}
	authFake.CreateFn = func(fake *rtesting.SubjectAccessFake) (*v1.SelfSubjectAccessReview, error) {
		ssar := &v1.SelfSubjectAccessReview{
			Status: v1.SubjectAccessReviewStatus{
				Allowed: true,
			},
		}
		return ssar, nil
	}

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	ra := resource.NewResourceAccess(ctx, authFake, "default", []resource.Resource{deploymentResource},
		resource.WithLogger(zap.NewNop()),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
}

func TestResourceAccessChecksDenied(t *testing.T) {
	authFake := rtesting.SubjectAccessFake{}
	authFake.CreateFn = func(fake *rtesting.SubjectAccessFake) (*v1.SelfSubjectAccessReview, error) {
		ssar := &v1.SelfSubjectAccessReview{
			Status: v1.SubjectAccessReviewStatus{
				Allowed: false,
			},
		}
		return ssar, nil
	}

	rbacDenied := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "resource failed minimum RBAC requirement") && e.Level == zap.WarnLevel {
			rbacDenied = true
		}
		return nil
	})))

	ra := resource.NewResourceAccess(context.TODO(), authFake, "default", []resource.Resource{deploymentResource},
		resource.WithLogger(logger),
		resource.WithMinimumRBAC([]string{"list", "watch"}),
	)
	assert.NotNil(t, ra)
	assert.True(t, rbacDenied)

	assert.False(t, ra.Allowed("default", deploymentResource, "list"))
	assert.False(t, ra.AllowedAll("default", deploymentResource, []string{"list", "watch"}))
	assert.False(t, ra.AllowedAny("default", deploymentResource, []string{"list", "watch"}))

	assert.Contains(t, ra.String(), "default.apps.v1.deployment.list: 0")
	assert.Contains(t, ra.String(), "default.apps.v1.deployment.watch: 0")

	deploymentResource.APIResource.Namespaced = false
	ra.Update(context.TODO(), authFake, "default", deploymentResource, "patch")
	assert.Contains(t, ra.String(), "apps.v1.deployment.patch: 2")
}

var deploymentResource = resource.Resource{
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
