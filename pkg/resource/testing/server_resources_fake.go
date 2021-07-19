package testing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

var _ discovery.ServerResourcesInterface = (*ServerResourcesFake)(nil)

type ServerResourcesFake struct {
	Empty           bool
	NoVerbs         bool
	Namespaced      bool
	APIResourceList []*metav1.APIResourceList

	ServerPreferredResourcesFn func(*ServerResourcesFake) ([]*metav1.APIResourceList, error)
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (s ServerResourcesFake) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return nil, nil
}

// ServerResources returns the supported resources for all groups and versions.
//
// The returned resource list might be non-nil with partial results even in the case of
// non-nil error.
//
// Deprecated: use ServerGroupsAndResources instead.
func (s ServerResourcesFake) ServerResources() ([]*metav1.APIResourceList, error) { return nil, nil }

// ServerResources returns the supported groups and resources for all groups and versions.
//
// The returned group and resource lists might be non-nil with partial results even in the
// case of non-nil error.
func (s ServerResourcesFake) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return nil, nil, nil
}

// ServerPreferredResources returns the supported resources with the version preferred by the
// server.
//
// The returned group and resource lists might be non-nil with partial results even in the
// case of non-nil error.
func (s ServerResourcesFake) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	if s.ServerPreferredResourcesFn != nil {
		return s.ServerPreferredResourcesFn(&s)
	}
	return ServerPreferredResources(&s)
}

func ServerPreferredResources(fake *ServerResourcesFake) ([]*metav1.APIResourceList, error) {
	resources := []metav1.APIResource{
		{Name: "deployments", SingularName: "deployment", Group: "apps", Version: "v1", Verbs: metav1.Verbs{"list", "watch"}, Kind: "deployment", Namespaced: fake.Namespaced},
	}
	resource_list := &metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind: "deployment",
		},
		GroupVersion: "apps/v1",
		APIResources: resources,
	}

	if fake.NoVerbs {
		resources[0].Verbs = metav1.Verbs{}
	}

	if fake.Empty {
		return []*metav1.APIResourceList{{APIResources: []metav1.APIResource{}}}, nil
	}
	return []*metav1.APIResourceList{resource_list}, nil
}

// ServerPreferredNamespacedResources returns the supported namespaced resources with the
// version preferred by the server.
//
// The returned resource list might be non-nil with partial results even in the case of
// non-nil error.
func (s ServerResourcesFake) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	if s.ServerPreferredResourcesFn != nil {
		return s.ServerPreferredResourcesFn(&s)
	}
	return ServerPreferredResources(&s)
}
