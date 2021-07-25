package cache

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// FilteredWatcheDetail is a thin wrapped around a WatchDetail
// used when a NamespaceAll watcher is available it scopes that
// WatchDetail List and Get calls to a single namespace.
type FilteredWatchDetail struct {
	Detail    *WatchDetail
	namespace string
}

var _ ResourceLister = (*FilteredWatchDetail)(nil)

func (w FilteredWatchDetail) List(selector labels.Selector) ([]runtime.Object, error) {
	return w.Detail.informer.Lister().ByNamespace(w.namespace).List(selector)
}

func (w *FilteredWatchDetail) Key() string {
	return w.Detail.Key()
}

func (w *FilteredWatchDetail) Namespace() string {
	return w.namespace
}

func (w *FilteredWatchDetail) Drain(ch chan<- interface{}, stopCh chan struct{}) {
	w.Detail.Drain(ch, stopCh)
}

func (w *FilteredWatchDetail) Get(name string) (runtime.Object, error) {
	return w.Detail.informer.Lister().ByNamespace(w.namespace).Get(name)
}

func (w *FilteredWatchDetail) Stop() {
	w.Detail.Stop()
}

func (w *FilteredWatchDetail) IsRunning() int {
	return w.Detail.IsRunning()
}
