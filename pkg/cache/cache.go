package cache

import (
	"sync"

	"github.com/wwitzel3/subjectaccess/pkg/subjectaccess"
)

var Resources *ResourceCache
var Namespaces []string

func init() {
	Resources = NewResourceCache()
	Namespaces = make([]string, 0)
}

func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		_map: &sync.Map{}, // key:string, value:[]subjectaccess.Resource
	}
}

type ResourceCache struct {
	_map *sync.Map
}

func (r *ResourceCache) AddResources(namespace string, resources ...subjectaccess.Resource) {
	v, loaded := r._map.LoadOrStore(namespace, resources)
	if loaded {
		existingResources, _ := v.([]subjectaccess.Resource)
		existingResources = append(existingResources, resources...)
		existingResources = unique(existingResources)
		r._map.Store(namespace, existingResources)
	}
}

func (r *ResourceCache) GetResources(namespace string) []subjectaccess.Resource {
	v, loaded := r._map.Load(namespace)
	if !loaded {
		return []subjectaccess.Resource{}
	}
	resources, _ := v.([]subjectaccess.Resource)
	return resources
}

func unique(resources []subjectaccess.Resource) []subjectaccess.Resource {
	keys := make(map[string]struct{})
	list := []subjectaccess.Resource{}
	for _, resource := range resources {
		if _, value := keys[resource.Key()]; !value {
			keys[resource.Key()] = struct{}{}
			list = append(list, resource)
		}
	}
	return list
}
