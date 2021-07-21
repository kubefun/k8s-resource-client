package testing

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var _ dynamicinformer.DynamicSharedInformerFactory = (*FakeDynamicSharedInformerFactory)(nil)

type FakeDynamicSharedInformerFactory struct{}

func (d FakeDynamicSharedInformerFactory) Start(stopCh <-chan struct{}) {}
func (d FakeDynamicSharedInformerFactory) ForResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return FakeGenericInformer{}
}
func (d FakeDynamicSharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[schema.GroupVersionResource]bool {
	return map[schema.GroupVersionResource]bool{}
}

type FakeGenericInformer struct{}

func (s FakeGenericInformer) Informer() cache.SharedIndexInformer { return FakeSharedIndexInformer{} }
func (s FakeGenericInformer) Lister() cache.GenericLister         { return nil }

type FakeSharedIndexInformer struct{}

func (s FakeSharedIndexInformer) AddEventHandler(handler cache.ResourceEventHandler) {}
func (s FakeSharedIndexInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
}
func (s FakeSharedIndexInformer) GetStore() cache.Store           { return nil }
func (s FakeSharedIndexInformer) GetController() cache.Controller { return nil }
func (s FakeSharedIndexInformer) Run(stopCh <-chan struct{})      {}
func (s FakeSharedIndexInformer) HasSynced() bool                 { return true }
func (s FakeSharedIndexInformer) LastSyncResourceVersion() string { return "" }
func (s FakeSharedIndexInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	return nil
}
func (s FakeSharedIndexInformer) AddIndexers(indexers cache.Indexers) error { return nil }
func (s FakeSharedIndexInformer) GetIndexer() cache.Indexer                 { return nil }
