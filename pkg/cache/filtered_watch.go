package cache

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type FilteredWatchDetail struct {
	detail    *WatchDetail
	namespace string
}

var _ ResourceLister = (*FilteredWatchDetail)(nil)

func (w FilteredWatchDetail) List(selector labels.Selector) ([]runtime.Object, error) {
	if w.namespace == metav1.NamespaceAll {
		return w.detail.informer.Lister().List(selector)
	}
	return w.detail.informer.Lister().ByNamespace(w.namespace).List(selector)
}

func (w *FilteredWatchDetail) Key() string {
	return w.detail.Key()
}

func (w *FilteredWatchDetail) Namespace() string {
	return w.namespace
}

func (w *FilteredWatchDetail) Drain(ch chan<- interface{}, stopCh chan struct{}) {
	w.detail.Drain(ch, stopCh)
}

func (w *FilteredWatchDetail) Get(name string) (runtime.Object, error) {
	return w.detail.Get(name)
}

func (w *FilteredWatchDetail) Stop() {
	w.detail.Stop()
}
