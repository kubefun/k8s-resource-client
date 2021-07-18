package cache

import (
	"sync"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var Resources *ResourceCache
var Namespaces []string
var Access resource.ResourceAccess

func init() {
	Resources = NewResourceCache()
	Namespaces = make([]string, 0)
	Access = nil
}

func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		_map: &sync.Map{}, // key:string, value:[]subjectaccess.Resource
	}
}

type ResourceCache struct {
	_map *sync.Map
}

func (r *ResourceCache) AddResources(key string, resources ...resource.Resource) {
	resources = unique(resources)

	v, loaded := r._map.LoadOrStore(key, resources)
	if loaded {
		existingResources, _ := v.([]resource.Resource)
		existingResources = append(existingResources, resources...)
		existingResources = unique(existingResources)
		r._map.Store(key, existingResources)
	}
}

func (r *ResourceCache) GetResources(key string) []resource.Resource {
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
