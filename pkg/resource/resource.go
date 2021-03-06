// Package subjectaccess provides functions for listing resource access in a Kubernetes cluster.
package resource

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	authClient "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	Denied int = iota
	Allowed
	Unused
	Error
)

type Resource struct {
	GroupVersionKind schema.GroupVersionKind
	APIResource      metav1.APIResource
}

func (r Resource) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    r.GroupVersionKind.Group,
		Version:  r.GroupVersionKind.Version,
		Resource: r.APIResource.Name,
	}
}

func (r Resource) Key() string {
	gvk := r.GroupVersionKind
	if gvk.Group == "" {
		return fmt.Sprintf("%s.%s", gvk.Version, gvk.Kind)
	}
	return fmt.Sprintf("%s.%s.%s", gvk.Group, gvk.Version, gvk.Kind)
}

// ResourceList creates a list of Resource objects using the Discovery client.
func ResourceList(_ context.Context, logger *zap.Logger, client discovery.ServerResourcesInterface, namespaced bool) ([]Resource, error) {
	if client == nil {
		return nil, fmt.Errorf("discoveryClient is nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	var resourceList func() ([]*metav1.APIResourceList, error)
	if namespaced {
		resourceList = client.ServerPreferredResources
	} else {
		resourceList = client.ServerPreferredNamespacedResources
	}

	resources, err := resourceList()
	if err != nil {
		if resources == nil {
			return nil, fmt.Errorf("get preferred resources: %w", err)
		}
		logger.Warn("unable to get full resource list",
			zap.Error(err),
		)
	}

	var result []Resource
	for _, resp := range resources {
		if len(resp.APIResources) == 0 {
			continue
		}

		groupVersion, err := schema.ParseGroupVersion(resp.GroupVersion)
		if err != nil {
			logger.Warn("unable to parse groupVersion", zap.Error(err))
			continue
		}

		for _, r := range resp.APIResources {
			if len(r.Verbs) == 0 {
				continue
			}

			result = append(result, Resource{
				GroupVersionKind: schema.GroupVersionKind{
					Version: groupVersion.Version,
					Group:   groupVersion.Group,
					Kind:    r.Kind,
				},
				APIResource: r,
			})
		}
	}

	return result, nil
}

func resourceVerbKey(namespace, key, verb string) string {
	return fmt.Sprintf("%s.%s.%s", namespace, key, verb)
}

// ResourceAccess provides a way to check if a given resource and verb are allowed to be performed by
// the current Kubernetes client.
type ResourceAccess interface {
	Update(context.Context, authClient.SelfSubjectAccessReviewInterface, string, Resource, string)
	Allowed(namespace string, resource Resource, verb string) bool
	AllowedAll(namespace string, resource Resource, verbs []string) bool
	AllowedAny(namespace string, resource Resource, verbs []string) bool
	String() string
}

var _ ResourceAccess = (*resourceAccess)(nil)

// NewResourceAccess provides a ResourceAccess object with an access map popluated from issuing SelfSubjectAccessReview
// requests for the list of resources and verbs provided.
func NewResourceAccess(ctx context.Context, client authClient.SelfSubjectAccessReviewInterface, namespace string, resources []Resource, options ...ResourceAccessOption) *resourceAccess {
	ra := &resourceAccess{
		access:       sync.Map{},
		logger:       zap.NewNop(),
		minimumVerbs: metav1.Verbs{"list", "watch"},
		namespace:    namespace,
	}

	for _, o := range options {
		o(ra)
	}

	group := sync.WaitGroup{}
	for _, resource := range resources {
		group.Add(1)

		r := resource
		go func() {
			defer group.Done()

			for _, verb := range ra.minimumVerbs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				ra.Update(ctx, client, namespace, r, verb)
			}
		}()
	}

	group.Wait()

	return ra
}

type resourceAccess struct {
	access       sync.Map
	logger       *zap.Logger
	minimumVerbs metav1.Verbs
	namespace    string
}

// Allowed checks if the given verb is allowed for the GVK.
func (r *resourceAccess) Allowed(namespace string, resource Resource, verb string) bool {
	key := resourceVerbKey(namespace, resource.Key(), verb)

	v, found := r.access.Load(key)
	if !found {
		r.logger.Debug("not found",
			zap.String("key", key),
		)
		return false
	}

	s, ok := v.(int)
	if !ok {
		r.logger.Warn("unable to type convert status to int, malformed access map",
			zap.String("value", fmt.Sprintf("%v", v)),
		)
		return false
	}

	return statusIntAsBool(s)
}

// AllowedAll checks if all of the given verbs are allowed for the GVK.
func (r *resourceAccess) AllowedAll(namespace string, resource Resource, verbs []string) bool {
	for _, verb := range verbs {
		if !r.Allowed(namespace, resource, verb) {
			return false
		}
	}
	return true
}

// AllowedAny checks if any of the given verbs are allowed for the GVK.
func (r *resourceAccess) AllowedAny(namespace string, resource Resource, verbs []string) bool {
	for _, verb := range verbs {
		if r.Allowed(namespace, resource, verb) {
			return true
		}
	}
	return false
}

func (ra *resourceAccess) Update(ctx context.Context, client authClient.SelfSubjectAccessReviewInterface, namespace string, resource Resource, verb string) {
	apiVerbs := sets.NewString(resource.APIResource.Verbs...)
	key := resourceVerbKey(namespace, resource.Key(), verb)

	if !apiVerbs.Has(verb) {
		ra.access.Store(key, Unused)
		return
	}

	sar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:      verb,
				Resource:  resource.APIResource.Name,
				Group:     resource.GroupVersionKind.Group,
				Namespace: namespace,
			},
		},
	}

	if result, err := client.Create(ctx, sar, metav1.CreateOptions{}); err != nil {
		ra.logger.Error("error SelfSubjectAccessReview", zap.Error(err))
		ra.access.Store(key, Error)
	} else {
		if result.Status.Allowed {
			ra.access.Store(key, Allowed)
		} else {
			ra.logger.Warn("resource failed minimum RBAC requirement",
				zap.String("reason", result.Status.Reason),
				zap.String("evaluation_error", result.Status.EvaluationError),
				zap.String("resource", fmt.Sprintf("%v", resource.APIResource)),
				zap.String("minimum_verbs", fmt.Sprintf("%v", ra.minimumVerbs)),
			)
			ra.access.Store(key, Denied)
		}
	}
}

func (r *resourceAccess) String() string {
	result := ""
	printer := func(key, value interface{}) bool {
		s, ok := key.(string)
		if !ok {
			return false
		}

		v, ok := value.(int)
		if !ok {
			return false
		}

		result += fmt.Sprintf("%s: %d\n", s, v)

		return true
	}
	r.access.Range(printer)
	return result
}

func statusIntAsBool(i int) bool {
	if i == Denied || i == Unused {
		return false
	}
	if i == Allowed {
		return true
	}
	return false
}
