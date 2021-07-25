package resource_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	rtesting "github.com/wwitzel3/k8s-resource-client/pkg/resource/testing"
)

func TestResourceListNilClient(t *testing.T) {
	_, err := resource.ResourceList(context.TODO(), nil, nil, false)
	assert.EqualError(t, err, "discoveryClient is nil")
}

func TestResourceListClusterScoped(t *testing.T) {

	client := rtesting.ServerResourcesFake{}

	r, err := resource.ResourceList(context.TODO(), nil, client, false)
	assert.Nil(t, err)

	assert.Len(t, r, 1)
	resource := r[0]

	assert.Equal(t, resource.GroupVersionKind, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "deployment"})
	assert.Equal(t, resource.Key(), "apps.v1.deployment")
}

func TestResourceListNamespaceScoped(t *testing.T) {

	client := rtesting.ServerResourcesFake{Namespaced: true}

	r, err := resource.ResourceList(context.TODO(), nil, client, true)
	assert.Nil(t, err)

	assert.Len(t, r, 1)
	resource := r[0]

	assert.Equal(t, resource.GroupVersionKind, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "deployment"})
	assert.Equal(t, resource.Key(), "apps.v1.deployment")
	resource.GroupVersionKind.Group = ""
	assert.Equal(t, resource.Key(), "v1.deployment")
}

func TestResourceListEmpty(t *testing.T) {

	client := rtesting.ServerResourcesFake{Namespaced: true, Empty: true}

	r, err := resource.ResourceList(context.TODO(), nil, client, true)
	assert.Nil(t, err)

	assert.Len(t, r, 0)
}

func TestResourceListNoVerbs(t *testing.T) {

	client := rtesting.ServerResourcesFake{Namespaced: true, NoVerbs: true}

	r, err := resource.ResourceList(context.TODO(), nil, client, true)
	assert.Nil(t, err)

	assert.Len(t, r, 0)
}
func TestResourceListBadGroupVersion(t *testing.T) {
	client := rtesting.ServerResourcesFake{Namespaced: true}
	client.ServerPreferredResourcesFn = func(fake *rtesting.ServerResourcesFake) ([]*v1.APIResourceList, error) {
		resources := []v1.APIResource{
			{Name: "deployments", SingularName: "deployment", Group: "apps", Version: "v1", Verbs: v1.Verbs{"list", "watch"}, Kind: "deployment", Namespaced: fake.Namespaced},
		}
		resource_list := &v1.APIResourceList{
			TypeMeta: v1.TypeMeta{
				Kind: "deployment",
			},
			GroupVersion: "_arston/---artei/o.-ar/t",
			APIResources: resources,
		}
		return []*v1.APIResourceList{resource_list}, nil
	}

	groupVersionWarning := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "unable to parse groupVersion") && e.Level == zap.WarnLevel {
			groupVersionWarning = true
		}
		return nil
	})))

	_, err := resource.ResourceList(context.TODO(), logger, client, true)
	assert.Nil(t, err)
	assert.True(t, groupVersionWarning)
}

func TestResourceListErr(t *testing.T) {
	client := rtesting.ServerResourcesFake{Namespaced: true}
	client.ServerPreferredResourcesFn = func(fake *rtesting.ServerResourcesFake) ([]*v1.APIResourceList, error) {
		return nil, fmt.Errorf("bad lookup err")
	}

	_, err := resource.ResourceList(context.TODO(), nil, client, true)
	assert.EqualError(t, err, "get preferred resources: bad lookup err")
}

func TestResourceListPartialListErr(t *testing.T) {
	client := rtesting.ServerResourcesFake{Namespaced: true}
	client.ServerPreferredResourcesFn = func(fake *rtesting.ServerResourcesFake) ([]*v1.APIResourceList, error) {
		resources := []v1.APIResource{
			{Name: "deployments", SingularName: "deployment", Group: "apps", Version: "v1", Verbs: v1.Verbs{"list", "watch"}, Kind: "deployment", Namespaced: fake.Namespaced},
		}
		resource_list := &v1.APIResourceList{
			TypeMeta: v1.TypeMeta{
				Kind: "deployment",
			},
			GroupVersion: "apps/v1",
			APIResources: resources,
		}
		return []*v1.APIResourceList{resource_list}, fmt.Errorf("recoverable error with lookup")

	}

	resourceListWarning := false
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "full resource list") && e.Level == zap.WarnLevel {
			resourceListWarning = true
		}
		return nil
	})))

	_, err := resource.ResourceList(context.TODO(), logger, client, true)
	assert.Nil(t, err)
	assert.True(t, resourceListWarning)
}
func TestResourceGVR(t *testing.T) {
	gvr := deploymentResource.GroupVersionResource()
	assert.Equal(t, "apps", gvr.Group)
	assert.Equal(t, "v1", gvr.Version)
	assert.Equal(t, "deployments", gvr.Resource)
}
