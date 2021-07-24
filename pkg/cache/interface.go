package cache

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceLister interface {
	// List will return all objects in this namespace
	List(selector labels.Selector) (ret []runtime.Object, err error)
	// Get will attempt to retrieve by namespace and name
	Get(name string) (runtime.Object, error)
	// Drain will get events from the queue and send them to the provided channel
	Drain(ch chan<- interface{}, stopCh chan struct{})
	// Stop
	Stop()
	// Namespace
	Namespace() string
	// Key
	Key() string
}
