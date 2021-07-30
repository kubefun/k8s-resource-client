package cache

import (
	"sync"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var Resources *ResourceCache
var Namespaces []string
var Access resource.ResourceAccess

type ResourceKey string

const (
	NamespacedResources    = ResourceKey("namespace")
	ClusterScopedResources = ResourceKey("cluster")
)

func init() {
	Resources = NewResourceCache()
	Namespaces = make([]string, 0)
	Access = nil
}

func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		_map: &sync.Map{}, // key:ResourceKey, value:[]subjectaccess.Resource
	}
}

type ResourceCache struct {
	_map *sync.Map
}

func (r *ResourceCache) Add(key ResourceKey, resources ...resource.Resource) {
	resources = unique(resources)

	v, loaded := r._map.LoadOrStore(key, resources)
	if loaded {
		existingResources, _ := v.([]resource.Resource)
		existingResources = append(existingResources, resources...)
		existingResources = unique(existingResources)
		r._map.Store(key, existingResources)
	}
}

// Get resturns all the resources for the given key
func (r *ResourceCache) Get(key ResourceKey) []resource.Resource {
	v, loaded := r._map.Load(key)
	if !loaded {
		return []resource.Resource{}
	}
	resources, _ := v.([]resource.Resource)
	return resources
}

func unique(resources []resource.Resource) []resource.Resource {
	keys := make(map[string]struct{})
	list := []resource.Resource{}
	for _, resource := range resources {
		if _, value := keys[resource.Key()]; !value {
			keys[resource.Key()] = struct{}{}
			list = append(list, resource)
		}
	}
	return list
}
